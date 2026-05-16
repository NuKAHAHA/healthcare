// @title           Healthcare API
// @version         1.0
// @description     Healthcare Management System REST API
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
// @description     Enter: Bearer {your_access_token}
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"healthcare-api/internal/auth"
	"healthcare-api/internal/config"
	"healthcare-api/internal/handlers"
	"healthcare-api/internal/logger"
	"healthcare-api/internal/middleware"
	"healthcare-api/internal/repositories"
	"healthcare-api/internal/services"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "healthcare-api/docs" // swagger spec registration

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v\n", err)
	}

	appLogger := logger.New(cfg.Logger.Level)
	appLogger.Info("Starting Healthcare API Server")

	// ── Database ──────────────────────────────────────────────
	db, err := initDB(cfg)
	if err != nil {
		log.Fatalf("database error: %v\n", err)
	}
	defer db.Close()
	appLogger.Info("Database connection established")

	// ── Redis ─────────────────────────────────────────────────
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		// Redis unavailable at startup — log warning but continue (rate limiting fails open)
		appLogger.WarnWithContext("Redis unavailable — rate limiting and token blacklist disabled",
			"startup", 0, err.Error())
	} else {
		appLogger.Info("Redis connection established")
	}
	defer func() { _ = rdb.Close() }()

	// ── Repositories ──────────────────────────────────────────
	userRepo := repositories.NewUserRepository(db)
	patientRepo := repositories.NewPatientRepository(db)
	appointmentRepo := repositories.NewAppointmentRepository(db)
	treatmentRepo := repositories.NewTreatmentRepository(db)
	auditRepo := repositories.NewAuditLogRepository(db)
	refreshRepo := repositories.NewRefreshTokenRepository(db)

	// ── Auth components ───────────────────────────────────────
	jwtManager := auth.NewJWTManager(&cfg.JWT)
	passwordMgr := auth.NewPasswordManager()
	blacklist := auth.NewTokenBlacklist(rdb)
	loginLimiter := middleware.NewLoginLimiter(rdb,
		cfg.RateLimit.LoginMaxAttempts,
		cfg.RateLimit.LoginWindowMin,
		cfg.RateLimit.LoginBlockMin,
	)

	// ── Services ──────────────────────────────────────────────
	authService := services.NewAuthService(userRepo, auditRepo, refreshRepo, jwtManager, passwordMgr, appLogger)
	patientService := services.NewPatientService(patientRepo, auditRepo, appLogger)
	appointmentService := services.NewAppointmentService(appointmentRepo, patientRepo, userRepo, auditRepo, appLogger)
	treatmentService := services.NewTreatmentService(treatmentRepo, appointmentRepo, patientRepo, auditRepo, appLogger)
	auditService := services.NewAuditService(auditRepo, appLogger)

	// ── Handlers ──────────────────────────────────────────────
	authHandler := handlers.NewAuthHandler(authService, appLogger, cfg.Server.MaxBodySize,
		loginLimiter, blacklist, jwtManager, cfg.IsDevelopment())
	patientHandler := handlers.NewPatientHandler(patientService, appointmentRepo, userRepo, appLogger, cfg.Server.MaxBodySize)
	appointmentHandler := handlers.NewAppointmentHandler(appointmentService, appointmentRepo, appLogger, cfg.Server.MaxBodySize)
	treatmentHandler := handlers.NewTreatmentHandler(treatmentService, appLogger, cfg.Server.MaxBodySize)
	auditHandler := handlers.NewAuditHandler(auditService, appLogger)
	userHandler := handlers.NewUserHandler(userRepo, appLogger)

	// ── Middleware shortcuts ───────────────────────────────────
	requireAuth := func(h http.HandlerFunc, roles ...string) http.Handler {
		handler := http.Handler(h)
		if len(roles) > 0 {
			handler = middleware.RoleBasedAuthorization(appLogger, roles...)(handler)
		}
		return middleware.AuthMiddleware(jwtManager, blacklist, appLogger)(handler)
	}

	// ── Router ────────────────────────────────────────────────
	mux := http.NewServeMux()

	// Public
	mux.HandleFunc("/health", healthHandler)
	mux.Handle("/auth/login", method(http.MethodPost, http.HandlerFunc(authHandler.Login)))
	mux.Handle("/auth/refresh", method(http.MethodPost, http.HandlerFunc(authHandler.Refresh)))

	// Swagger UI — CSP is relaxed for this prefix only in dev
	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// Auth (authenticated)
	mux.Handle("/auth/register", method(http.MethodPost, requireAuth(authHandler.Register, "admin")))
	mux.Handle("/auth/logout", method(http.MethodPost, requireAuth(authHandler.Logout)))

	// Patients
	mux.Handle("/patients", methodSwitch(map[string]http.Handler{
		http.MethodPost: requireAuth(patientHandler.RegisterPatient, "registrar", "admin"),
		http.MethodGet:  requireAuth(patientHandler.ListPatients, "registrar", "admin", "doctor"),
	}))
	mux.Handle("/patients/", method(http.MethodGet, requireAuth(patientHandler.GetPatient)))

	// Appointments
	mux.Handle("/appointments", methodSwitch(map[string]http.Handler{
		http.MethodPost: requireAuth(appointmentHandler.CreateAppointment, "registrar", "admin"),
		http.MethodGet:  requireAuth(appointmentHandler.ListAppointments),
	}))
	mux.Handle("PATCH /appointments/{id}/status", requireAuth(appointmentHandler.UpdateStatus, "admin", "registrar"))

	// Treatments
	mux.Handle("/treatments", method(http.MethodPost, requireAuth(treatmentHandler.AddTreatment, "doctor")))

	// Reports
	mux.Handle("/reports/", method(http.MethodGet, requireAuth(treatmentHandler.GetReport, "doctor", "admin")))

	// Users (for scheduling dropdowns)
	mux.Handle("/users", method(http.MethodGet, requireAuth(userHandler.ListUsers, "admin", "registrar")))

	// Audit logs
	mux.Handle("/audit-logs/", method(http.MethodGet, requireAuth(auditHandler.GetAuditLogs, "admin")))

	// ── Global middleware chain ────────────────────────────────
	// Order (outer → inner): CORS → SecurityHeaders → RateLimit → PanicRecovery → RequestLogger → mux
	var handler http.Handler = mux
	handler = middleware.PanicRecovery(appLogger)(handler)
	handler = middleware.GlobalRateLimit(rdb, cfg.RateLimit.GlobalRPM)(handler)
	handler = middleware.SecurityHeaders()(handler)
	handler = middleware.CORS(cfg.CORS.AllowedOrigins)(handler)
	handler = middleware.RequestLogger(appLogger)(handler)

	// ── Server ────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      handler,
		ReadTimeout:  time.Duration(cfg.Server.RequestTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.RequestTimeout) * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		appLogger.Info(fmt.Sprintf("Server listening on :%d — Swagger UI: http://localhost:%d/swagger/index.html", cfg.Server.Port, cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v\n", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	appLogger.Info("Shutting down...")
	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Fatalf("graceful shutdown error: %v\n", err)
	}
	appLogger.Info("Shutdown complete")
}

func initDB(cfg *config.Config) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return pool, nil
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "healthcare-api",
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

// method restricts a handler to a single HTTP method.
func method(m string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != m {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = w.Write([]byte(`{"code":"method_not_allowed","message":"Method Not Allowed"}`))
			return
		}
		h.ServeHTTP(w, r)
	})
}

// methodSwitch dispatches to different handlers based on HTTP method.
func methodSwitch(handlers map[string]http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h, ok := handlers[r.Method]; ok {
			h.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte(`{"code":"method_not_allowed","message":"Method Not Allowed"}`))
	})
}
