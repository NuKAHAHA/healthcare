package handlers

import (
	"encoding/json"
	"healthcare-api/internal/auth"
	"healthcare-api/internal/logger"
	"healthcare-api/internal/middleware"
	"healthcare-api/internal/models"
	"healthcare-api/internal/services"
	"net/http"
	"strings"
	"time"
)

// AuthHandler handles authentication endpoints.
// @title Healthcare API Authentication
type AuthHandler struct {
	authService  *services.AuthService
	logger       *logger.Logger
	maxBodySize  int64
	loginLimiter *middleware.LoginLimiter
	blacklist    *auth.TokenBlacklist
	jwtManager   *auth.JWTManager
	isDev        bool
}

func NewAuthHandler(
	authService *services.AuthService,
	appLogger *logger.Logger,
	maxBodySize int64,
	loginLimiter *middleware.LoginLimiter,
	blacklist *auth.TokenBlacklist,
	jwtManager *auth.JWTManager,
	isDev bool,
) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		logger:       appLogger,
		maxBodySize:  maxBodySize,
		loginLimiter: loginLimiter,
		blacklist:    blacklist,
		jwtManager:   jwtManager,
		isDev:        isDev,
	}
}

// Login handles POST /auth/login
// @Summary      User login
// @Description  Authenticate with email and password. Returns access token in body and sets httpOnly refresh token cookie.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        body  body      models.LoginRequest  true  "Login credentials"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  models.ErrorResponse
// @Failure      401   {object}  models.ErrorResponse
// @Failure      429   {object}  models.ErrorResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxBodySize)
	defer func() { _ = r.Body.Close() }()

	var req models.LoginRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		sendJSONError(w, http.StatusBadRequest, "invalid_request", "Invalid request format")
		return
	}
	if req.Email == "" || req.Password == "" {
		sendJSONError(w, http.StatusBadRequest, "invalid_input", "Email and password are required")
		return
	}

	ip := middleware.ClientIP(r)

	// Redis-backed brute-force protection
	var attemptsRemaining int64 = -1
	if h.loginLimiter != nil {
		result := h.loginLimiter.Allow(r.Context(), ip, req.Email)
		if !result.Allowed {
			h.logger.WarnWithContext("login rate limit exceeded", "login_attempt", 0, "ip="+ip)
			retryAfterSec := int64(result.RetryAfter.Seconds())
			if retryAfterSec <= 0 {
				retryAfterSec = 900
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code":        "rate_limited",
				"message":     "Too many login attempts. Try again later.",
				"retry_after": retryAfterSec,
			})
			return
		}
		attemptsRemaining = result.Remaining
	}

	resp, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		body := map[string]interface{}{
			"code":    "invalid_credentials",
			"message": "Invalid email or password",
		}
		if attemptsRemaining >= 0 {
			body["attempts_remaining"] = attemptsRemaining
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(body)
		return
	}

	if h.loginLimiter != nil {
		h.loginLimiter.Reset(r.Context(), ip, req.Email)
	}

	// Set refresh token in httpOnly cookie
	h.setRefreshCookie(w, resp.RefreshToken, time.Now().Add(7*24*time.Hour))

	// Return access token + user info only (no refresh token in body)
	body := map[string]interface{}{
		"accessToken": resp.AccessToken,
		"expiresAt":   resp.ExpiresAt,
		"user":        resp.User,
	}
	sendJSONSuccess(w, http.StatusOK, body)
}

// Register handles POST /auth/register
// @Summary      Register user (Admin only)
// @Description  Create a new user account. Requires admin role.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      models.RegisterRequest  true  "User registration data"
// @Success      201   {object}  models.UserDTO
// @Failure      400   {object}  models.ErrorResponse
// @Failure      401   {object}  models.ErrorResponse
// @Failure      403   {object}  models.ErrorResponse
// @Router       /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxBodySize)
	defer func() { _ = r.Body.Close() }()

	user, err := middleware.GetAuthUser(r)
	if err != nil {
		sendJSONError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}
	if user.Role != "admin" {
		sendJSONError(w, http.StatusForbidden, "forbidden", "Only admins can register users")
		return
	}

	var req models.RegisterRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		sendJSONError(w, http.StatusBadRequest, "invalid_request", "Invalid request format")
		return
	}
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" || req.Role == "" {
		sendJSONError(w, http.StatusBadRequest, "invalid_input", "All fields are required")
		return
	}
	if req.Role != "admin" && req.Role != "registrar" && req.Role != "doctor" {
		sendJSONError(w, http.StatusBadRequest, "invalid_input", "Invalid role. Use: admin, registrar, doctor")
		return
	}

	dto, err := h.authService.Register(r.Context(), &req, user.ID)
	if err != nil {
		h.logger.WarnWithContext("user registration failed", "user_registration", user.ID, err.Error())
		sendJSONError(w, http.StatusBadRequest, "invalid_input", "Registration failed: "+err.Error())
		return
	}
	sendJSONSuccess(w, http.StatusCreated, dto)
}

// Refresh handles POST /auth/refresh
// @Summary      Refresh access token
// @Description  Exchange a valid refresh token (from httpOnly cookie) for a new access token.
// @Tags         Authentication
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  models.ErrorResponse
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	// Primary: read from httpOnly cookie
	refreshToken := ""
	if cookie, err := r.Cookie("refreshToken"); err == nil {
		refreshToken = cookie.Value
	}

	// Fallback: read from JSON body (for Postman / API clients)
	if refreshToken == "" {
		r.Body = http.MaxBytesReader(w, r.Body, h.maxBodySize)
		defer func() { _ = r.Body.Close() }()
		var req struct {
			RefreshToken string `json:"refreshToken"`
		}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		_ = dec.Decode(&req)
		refreshToken = req.RefreshToken
	}

	if refreshToken == "" {
		sendJSONError(w, http.StatusUnauthorized, "unauthorized", "Refresh token is required")
		return
	}

	newAccess, newRefresh, expiresAt, userDTO, err := h.authService.RefreshAccessToken(r.Context(), refreshToken, middleware.ClientIP(r))
	if err != nil {
		h.logger.WarnWithContext("refresh failed", "token_refresh", 0,
			"ip="+middleware.ClientIP(r)+" reason="+err.Error())
		sendJSONError(w, http.StatusUnauthorized, "invalid_token", "Invalid or expired refresh token")
		return
	}

	// Rotate cookie
	h.setRefreshCookie(w, newRefresh, time.Now().Add(7*24*time.Hour))

	sendJSONSuccess(w, http.StatusOK, map[string]interface{}{
		"accessToken": newAccess,
		"expiresAt":   expiresAt,
		"user":        userDTO,
	})
}

// Logout handles POST /auth/logout
// @Summary      Logout
// @Description  Revoke access token (blacklist JTI) and all refresh tokens.
// @Tags         Authentication
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  models.ErrorResponse
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	user, err := middleware.GetAuthUser(r)
	if err != nil {
		sendJSONError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	// Blacklist the current access token so it can't be reused
	if h.blacklist != nil && h.jwtManager != nil {
		authHeader := r.Header.Get("Authorization")
		if parts := strings.SplitN(authHeader, " ", 2); len(parts) == 2 {
			rawToken := parts[1]
			jti, err := h.jwtManager.GetJTI(rawToken)
			if err == nil {
				expAt, err := h.jwtManager.GetExpirationTime(rawToken)
				if err == nil {
					if rErr := h.blacklist.Revoke(r.Context(), jti, expAt); rErr != nil {
						h.logger.WarnWithContext("failed to blacklist access token", "logout",
							user.ID, rErr.Error())
					}
				}
			}
		}
	}

	// Revoke all refresh tokens
	if err := h.authService.Logout(r.Context(), user.ID); err != nil {
		h.logger.WarnWithContext("refresh token revocation failed", "logout", user.ID, err.Error())
	}

	// Clear refresh cookie
	h.clearRefreshCookie(w)

	h.logger.InfoWithContext("user logged out", "logout", user.ID, "role="+user.Role)
	sendJSONSuccess(w, http.StatusOK, map[string]string{"message": "Logout successful"})
}

// ── helpers ─────────────────────────────────────────────────────────────────

func (h *AuthHandler) setRefreshCookie(w http.ResponseWriter, token string, expiry time.Time) {
	sameSite := http.SameSiteStrictMode
	secure := !h.isDev
	if h.isDev {
		sameSite = http.SameSiteLaxMode
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    token,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		Path:     "/",
		Expires:  expiry,
		MaxAge:   int(time.Until(expiry).Seconds()),
	})
}

func (h *AuthHandler) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    "",
		HttpOnly: true,
		Secure:   !h.isDev,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1,
	})
}

func sendJSONError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(models.ErrorResponse{Message: message, Code: code})
}

func sendJSONSuccess(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}
