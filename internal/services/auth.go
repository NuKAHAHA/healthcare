package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"healthcare-api/internal/auth"
	"healthcare-api/internal/logger"
	"healthcare-api/internal/models"
	"time"
)

type userRepository interface {
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id int64) (*models.User, error)
	Create(ctx context.Context, user *models.User) (*models.User, error)
}

type auditLogRepository interface {
	Create(ctx context.Context, log *models.AuditLog) (*models.AuditLog, error)
}

type refreshTokenRepository interface {
	Create(ctx context.Context, token *models.RefreshToken) (*models.RefreshToken, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error)
	RevokeByTokenHash(ctx context.Context, tokenHash string) error
	RevokeAllByUserID(ctx context.Context, userID int64) error
}

type AuthService struct {
	userRepo    userRepository
	auditRepo   auditLogRepository
	refreshRepo refreshTokenRepository
	jwtManager  *auth.JWTManager
	passwordMgr *auth.PasswordManager
	logger      *logger.Logger
}

func NewAuthService(
	userRepo userRepository,
	auditRepo auditLogRepository,
	refreshRepo refreshTokenRepository,
	jwtManager *auth.JWTManager,
	passwordMgr *auth.PasswordManager,
	appLogger *logger.Logger,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		auditRepo:   auditRepo,
		refreshRepo: refreshRepo,
		jwtManager:  jwtManager,
		passwordMgr: passwordMgr,
		logger:      appLogger,
	}
}

// Login authenticates a user and returns tokens plus a safe user DTO.
func (s *AuthService) Login(ctx context.Context, email, password string) (*models.LoginResponse, error) {
	if email == "" || password == "" {
		return nil, fmt.Errorf("email and password are required")
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		s.writeAudit(ctx, 0, "login_attempt", "user", 0, "failure", "user not found")
		s.logger.WarnWithContext("login failed — user not found", "login_attempt", 0, "")
		return nil, fmt.Errorf("invalid credentials")
	}

	if !s.passwordMgr.VerifyPassword(user.PasswordHash, password) {
		s.writeAudit(ctx, user.ID, "login_attempt", "user", user.ID, "failure", "invalid password")
		s.logger.WarnWithContext("login failed — wrong password", "login_attempt", user.ID, "")
		return nil, fmt.Errorf("invalid credentials")
	}

	accessToken, refreshToken, expiresAt, err := s.jwtManager.GenerateTokens(user)
	if err != nil {
		s.logger.ErrorWithContext("token generation failed", "token_generation", user.ID, err.Error())
		return nil, fmt.Errorf("failed to generate tokens")
	}
	if err := s.storeRefreshToken(ctx, user.ID, refreshToken); err != nil {
		s.logger.ErrorWithContext("refresh token store failed", "token_generation", user.ID, err.Error())
		return nil, fmt.Errorf("failed to generate tokens")
	}

	s.writeAudit(ctx, user.ID, "login", "user", user.ID, "success", "role="+user.Role)
	s.logger.InfoWithContext("user logged in", "login", user.ID, "role="+user.Role)

	return &models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		User: models.UserDTO{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      user.Role,
		},
	}, nil
}

// RefreshAccessToken rotates the refresh token and issues a new access token.
// Returns the new access token, new refresh token, expiry, user DTO, and error.
func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshToken string, ip string) (string, string, time.Time, *models.UserDTO, error) {
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return "", "", time.Time{}, nil, fmt.Errorf("invalid refresh token")
	}

	tokenType, _ := claims["type"].(string)
	if tokenType != "refresh" {
		return "", "", time.Time{}, nil, fmt.Errorf("invalid token type")
	}

	userIDFloat, ok := claims["userId"].(float64)
	if !ok {
		return "", "", time.Time{}, nil, fmt.Errorf("invalid token claims")
	}
	userID := int64(userIDFloat)

	tokenHash := hashToken(refreshToken)
	stored, err := s.refreshRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil || stored == nil || stored.RevokedAt != nil || stored.ExpiresAt.Before(time.Now()) {
		// Reuse-detection: if token was already revoked, revoke ALL user tokens
		if stored != nil && stored.RevokedAt != nil {
			_ = s.refreshRepo.RevokeAllByUserID(ctx, stored.UserID)
			s.logger.WarnWithContext("refresh token reuse detected — all sessions revoked", "token_refresh",
				stored.UserID, "ip="+ip)
		}
		return "", "", time.Time{}, nil, fmt.Errorf("invalid or expired refresh token")
	}
	if stored.UserID != userID {
		_ = s.refreshRepo.RevokeByTokenHash(ctx, tokenHash)
		return "", "", time.Time{}, nil, fmt.Errorf("invalid refresh token")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return "", "", time.Time{}, nil, fmt.Errorf("user not found")
	}

	// Rotate: revoke old, issue new
	if err := s.refreshRepo.RevokeByTokenHash(ctx, tokenHash); err != nil {
		return "", "", time.Time{}, nil, fmt.Errorf("failed to rotate refresh token")
	}

	newAccess, newRefresh, expiresAt, err := s.jwtManager.GenerateTokens(user)
	if err != nil {
		return "", "", time.Time{}, nil, fmt.Errorf("failed to generate tokens")
	}
	if err := s.storeRefreshToken(ctx, user.ID, newRefresh); err != nil {
		return "", "", time.Time{}, nil, fmt.Errorf("failed to store new refresh token")
	}

	// Audit log the refresh event
	s.writeAudit(ctx, user.ID, "token_refresh", "user", user.ID, "success",
		fmt.Sprintf("ip=%s role=%s", ip, user.Role))
	s.logger.InfoWithContext("access token refreshed", "token_refresh", user.ID, "ip="+ip)

	dto := &models.UserDTO{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
	}
	return newAccess, newRefresh, expiresAt, dto, nil
}

// Logout revokes all active refresh tokens for the user.
func (s *AuthService) Logout(ctx context.Context, userID int64) error {
	return s.refreshRepo.RevokeAllByUserID(ctx, userID)
}

// CreateUser creates a new user with a hashed password.
func (s *AuthService) CreateUser(ctx context.Context, email, password, firstName, lastName, role string) (*models.User, error) {
	if email == "" || password == "" || firstName == "" || lastName == "" {
		return nil, fmt.Errorf("all fields are required")
	}
	if len(password) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters")
	}

	if existing, _ := s.userRepo.GetByEmail(ctx, email); existing != nil {
		return nil, fmt.Errorf("email already registered")
	}

	hash, err := s.passwordMgr.HashPassword(password)
	if err != nil {
		s.logger.ErrorWithContext("password hash failed", "user_creation", 0, err.Error())
		return nil, fmt.Errorf("failed to create user")
	}

	user, err := s.userRepo.Create(ctx, &models.User{
		Email:        email,
		FirstName:    firstName,
		LastName:     lastName,
		Role:         role,
		PasswordHash: hash,
	})
	if err != nil {
		return nil, err
	}
	s.logger.InfoWithContext("user created", "user_creation", 0, "role="+role)
	return user, nil
}

// Register creates a user and writes an audit log (admin action).
func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest, registrarID int64) (*models.UserDTO, error) {
	user, err := s.CreateUser(ctx, req.Email, req.Password, req.FirstName, req.LastName, req.Role)
	if err != nil {
		return nil, err
	}
	s.writeAudit(ctx, registrarID, "user_registration", "user", user.ID, "success",
		fmt.Sprintf("role=%s registered_by=%d", req.Role, registrarID))

	return &models.UserDTO{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
	}, nil
}

// ── private helpers ──────────────────────────────────────────────────────────

func (s *AuthService) storeRefreshToken(ctx context.Context, userID int64, token string) error {
	_, err := s.refreshRepo.Create(ctx, &models.RefreshToken{
		UserID:    userID,
		TokenHash: hashToken(token),
		ExpiresAt: s.jwtManager.GetRefreshExpiration(),
	})
	return err
}

func (s *AuthService) writeAudit(ctx context.Context, userID int64, action, resource string, resourceID int64, status, details string) {
	_, _ = s.auditRepo.Create(ctx, &models.AuditLog{
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Status:     status,
		Details:    details,
		CreatedAt:  time.Now(),
	})
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
