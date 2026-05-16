package services

import (
	"context"
	"testing"
	"time"

	"healthcare-api/internal/auth"
	"healthcare-api/internal/config"
	"healthcare-api/internal/logger"
	"healthcare-api/internal/models"
)

// MockUserRepository for testing
type MockUserRepository struct {
	users map[string]*models.User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*models.User),
	}
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	user.ID = 1
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	m.users[user.Email] = user
	return user, nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user, exists := m.users[email]
	if !exists {
		return nil, nil
	}
	return user, nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, nil
}

func (m *MockUserRepository) GetAll(ctx context.Context) ([]*models.User, error) {
	var users []*models.User
	for _, user := range m.users {
		users = append(users, user)
	}
	return users, nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) (*models.User, error) {
	m.users[user.Email] = user
	return user, nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id int64) error {
	for email, user := range m.users {
		if user.ID == id {
			delete(m.users, email)
			return nil
		}
	}
	return nil
}

// MockAuditRepository for testing
type MockAuditRepository struct {
	logs []*models.AuditLog
}

func NewMockAuditRepository() *MockAuditRepository {
	return &MockAuditRepository{
		logs: make([]*models.AuditLog, 0),
	}
}

func (m *MockAuditRepository) Create(ctx context.Context, auditLog *models.AuditLog) (*models.AuditLog, error) {
	auditLog.ID = int64(len(m.logs) + 1)
	m.logs = append(m.logs, auditLog)
	return auditLog, nil
}

func (m *MockAuditRepository) GetByID(ctx context.Context, id int64) (*models.AuditLog, error) {
	for _, log := range m.logs {
		if log.ID == id {
			return log, nil
		}
	}
	return nil, nil
}

func (m *MockAuditRepository) GetByResourceID(ctx context.Context, resourceID int64) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog
	for _, log := range m.logs {
		if log.ResourceID == resourceID {
			logs = append(logs, log)
		}
	}
	return logs, nil
}

// MockRefreshTokenRepository for testing
type MockRefreshTokenRepository struct {
	tokens map[string]*models.RefreshToken
}

func NewMockRefreshTokenRepository() *MockRefreshTokenRepository {
	return &MockRefreshTokenRepository{
		tokens: make(map[string]*models.RefreshToken),
	}
}

func (m *MockRefreshTokenRepository) Create(ctx context.Context, token *models.RefreshToken) (*models.RefreshToken, error) {
	token.ID = int64(len(m.tokens) + 1)
	token.CreatedAt = time.Now()
	m.tokens[token.TokenHash] = token
	return token, nil
}

func (m *MockRefreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	token, exists := m.tokens[tokenHash]
	if !exists {
		return nil, nil
	}
	return token, nil
}

func (m *MockRefreshTokenRepository) RevokeByTokenHash(ctx context.Context, tokenHash string) error {
	token, exists := m.tokens[tokenHash]
	if exists {
		now := time.Now()
		token.RevokedAt = &now
	}
	return nil
}

func (m *MockRefreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID int64) error {
	now := time.Now()
	for _, token := range m.tokens {
		if token.UserID == userID && token.RevokedAt == nil {
			token.RevokedAt = &now
		}
	}
	return nil
}

func newTestAuthService(userRepo *MockUserRepository, auditRepo *MockAuditRepository, refreshRepo *MockRefreshTokenRepository, jwtManager *auth.JWTManager, passwordMgr *auth.PasswordManager, appLogger *logger.Logger) *AuthService {
	return NewAuthService(userRepo, auditRepo, refreshRepo, jwtManager, passwordMgr, appLogger)
}

// TestAuthService_Login_Success tests successful login
func TestAuthService_Login_Success(t *testing.T) {
	userRepo := NewMockUserRepository()
	auditRepo := NewMockAuditRepository()
	refreshRepo := NewMockRefreshTokenRepository()
	jwtCfg := &config.JWTConfig{
		Secret:           "test-secret-key-minimum-32-characters-long-!!!",
		AccessExpireMin:  15,
		RefreshExpireDay: 7,
	}
	jwtManager := auth.NewJWTManager(jwtCfg)
	passwordMgr := auth.NewPasswordManager()
	appLogger := logger.New("info")

	authService := newTestAuthService(userRepo, auditRepo, refreshRepo, jwtManager, passwordMgr, appLogger)
	ctx := context.Background()

	// Create test user
	password := "TestPassword123!"
	hash, _ := passwordMgr.HashPassword(password)
	user := &models.User{
		ID:           1,
		Email:        "test@example.com",
		FirstName:    "Test",
		LastName:     "User",
		Role:         "admin",
		PasswordHash: hash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	userRepo.users["test@example.com"] = user

	// Test login
	response, err := authService.Login(ctx, "test@example.com", password)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected non-nil response")
	}

	if response.AccessToken == "" {
		t.Fatal("Expected non-empty access token")
	}

	if response.RefreshToken == "" {
		t.Fatal("Expected non-empty refresh token")
	}

	if response.User.Email != "test@example.com" {
		t.Fatalf("Expected email test@example.com, got %v", response.User.Email)
	}

	// Verify audit log was created
	if len(auditRepo.logs) == 0 {
		t.Fatal("Expected audit log to be created")
	}

	auditLog := auditRepo.logs[0]
	if auditLog.Action != "login" {
		t.Fatalf("Expected action 'login', got %v", auditLog.Action)
	}

	if auditLog.Status != "success" {
		t.Fatalf("Expected status 'success', got %v", auditLog.Status)
	}
}

// TestAuthService_Login_InvalidCredentials tests login with wrong password
func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	userRepo := NewMockUserRepository()
	auditRepo := NewMockAuditRepository()
	refreshRepo := NewMockRefreshTokenRepository()
	jwtCfg := &config.JWTConfig{
		Secret:           "test-secret-key-minimum-32-characters-long-!!!",
		AccessExpireMin:  15,
		RefreshExpireDay: 7,
	}
	jwtManager := auth.NewJWTManager(jwtCfg)
	passwordMgr := auth.NewPasswordManager()
	appLogger := logger.New("info")

	authService := newTestAuthService(userRepo, auditRepo, refreshRepo, jwtManager, passwordMgr, appLogger)
	ctx := context.Background()

	// Create test user
	password := "TestPassword123!"
	hash, _ := passwordMgr.HashPassword(password)
	user := &models.User{
		ID:           1,
		Email:        "test@example.com",
		FirstName:    "Test",
		LastName:     "User",
		Role:         "admin",
		PasswordHash: hash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	userRepo.users["test@example.com"] = user

	// Test login with wrong password
	_, err := authService.Login(ctx, "test@example.com", "WrongPassword")
	if err == nil {
		t.Fatal("Expected login to fail with wrong password")
	}

	// Verify audit log shows failure
	if len(auditRepo.logs) == 0 {
		t.Fatal("Expected audit log to be created")
	}

	auditLog := auditRepo.logs[0]
	if auditLog.Status != "failure" {
		t.Fatalf("Expected status 'failure', got %v", auditLog.Status)
	}
}

// TestAuthService_Login_UserNotFound tests login with non-existent user
func TestAuthService_Login_UserNotFound(t *testing.T) {
	userRepo := NewMockUserRepository()
	auditRepo := NewMockAuditRepository()
	refreshRepo := NewMockRefreshTokenRepository()
	jwtCfg := &config.JWTConfig{
		Secret:           "test-secret-key-minimum-32-characters-long-!!!",
		AccessExpireMin:  15,
		RefreshExpireDay: 7,
	}
	jwtManager := auth.NewJWTManager(jwtCfg)
	passwordMgr := auth.NewPasswordManager()
	appLogger := logger.New("info")

	authService := newTestAuthService(userRepo, auditRepo, refreshRepo, jwtManager, passwordMgr, appLogger)
	ctx := context.Background()

	// Test login with non-existent user
	_, err := authService.Login(ctx, "nonexistent@example.com", "password")
	if err == nil {
		t.Fatal("Expected login to fail with non-existent user")
	}
}

// TestAuthService_Login_EmptyEmail tests login with empty email
func TestAuthService_Login_EmptyEmail(t *testing.T) {
	userRepo := NewMockUserRepository()
	auditRepo := NewMockAuditRepository()
	refreshRepo := NewMockRefreshTokenRepository()
	jwtCfg := &config.JWTConfig{
		Secret:           "test-secret-key-minimum-32-characters-long-!!!",
		AccessExpireMin:  15,
		RefreshExpireDay: 7,
	}
	jwtManager := auth.NewJWTManager(jwtCfg)
	passwordMgr := auth.NewPasswordManager()
	appLogger := logger.New("info")

	authService := newTestAuthService(userRepo, auditRepo, refreshRepo, jwtManager, passwordMgr, appLogger)
	ctx := context.Background()

	_, err := authService.Login(ctx, "", "password")
	if err == nil {
		t.Fatal("Expected login to fail with empty email")
	}
}

// TestAuthService_CreateUser_Success tests successful user creation
func TestAuthService_CreateUser_Success(t *testing.T) {
	userRepo := NewMockUserRepository()
	auditRepo := NewMockAuditRepository()
	refreshRepo := NewMockRefreshTokenRepository()
	jwtCfg := &config.JWTConfig{
		Secret:           "test-secret-key-minimum-32-characters-long-!!!",
		AccessExpireMin:  15,
		RefreshExpireDay: 7,
	}
	jwtManager := auth.NewJWTManager(jwtCfg)
	passwordMgr := auth.NewPasswordManager()
	appLogger := logger.New("info")

	authService := newTestAuthService(userRepo, auditRepo, refreshRepo, jwtManager, passwordMgr, appLogger)
	ctx := context.Background()

	user, err := authService.CreateUser(ctx, "newuser@example.com", "TestPassword123!", "John", "Doe", "admin")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if user == nil {
		t.Fatal("Expected non-nil user")
	}

	if user.Email != "newuser@example.com" {
		t.Fatalf("Expected email newuser@example.com, got %v", user.Email)
	}

	if user.FirstName != "John" {
		t.Fatalf("Expected firstName John, got %v", user.FirstName)
	}

	if user.Role != "admin" {
		t.Fatalf("Expected role admin, got %v", user.Role)
	}

	if user.PasswordHash == "" {
		t.Fatal("Expected non-empty password hash")
	}
}

// TestAuthService_CreateUser_ShortPassword tests user creation with password too short
func TestAuthService_CreateUser_ShortPassword(t *testing.T) {
	userRepo := NewMockUserRepository()
	auditRepo := NewMockAuditRepository()
	refreshRepo := NewMockRefreshTokenRepository()
	jwtCfg := &config.JWTConfig{
		Secret:           "test-secret-key-minimum-32-characters-long-!!!",
		AccessExpireMin:  15,
		RefreshExpireDay: 7,
	}
	jwtManager := auth.NewJWTManager(jwtCfg)
	passwordMgr := auth.NewPasswordManager()
	appLogger := logger.New("info")

	authService := newTestAuthService(userRepo, auditRepo, refreshRepo, jwtManager, passwordMgr, appLogger)
	ctx := context.Background()

	_, err := authService.CreateUser(ctx, "newuser@example.com", "short", "John", "Doe", "admin")
	if err == nil {
		t.Fatal("Expected CreateUser to fail with short password")
	}
}

// TestAuthService_RefreshAccessToken_Success tests successful token refresh
func TestAuthService_RefreshAccessToken_Success(t *testing.T) {
	userRepo := NewMockUserRepository()
	auditRepo := NewMockAuditRepository()
	refreshRepo := NewMockRefreshTokenRepository()
	jwtCfg := &config.JWTConfig{
		Secret:           "test-secret-key-minimum-32-characters-long-!!!",
		AccessExpireMin:  15,
		RefreshExpireDay: 7,
	}
	jwtManager := auth.NewJWTManager(jwtCfg)
	passwordMgr := auth.NewPasswordManager()
	appLogger := logger.New("info")

	authService := newTestAuthService(userRepo, auditRepo, refreshRepo, jwtManager, passwordMgr, appLogger)
	ctx := context.Background()

	// Create test user
	user := &models.User{
		ID:           1,
		Email:        "test@example.com",
		FirstName:    "Test",
		LastName:     "User",
		Role:         "admin",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	userRepo.users["test@example.com"] = user

	// Generate tokens
	_, refreshToken, _, err := jwtManager.GenerateTokens(user)
	if err != nil {
		t.Fatalf("GenerateTokens failed: %v", err)
	}
	_, err = refreshRepo.Create(ctx, &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: hashToken(refreshToken),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Create refresh token failed: %v", err)
	}

	// Test refresh
	newAccessToken, newRefreshToken, expiresAt, _, err := authService.RefreshAccessToken(ctx, refreshToken, "")
	if err != nil {
		t.Fatalf("RefreshAccessToken failed: %v", err)
	}

	if newAccessToken == "" {
		t.Fatal("Expected non-empty access token")
	}

	if newRefreshToken == "" {
		t.Fatal("Expected non-empty refresh token")
	}

	if expiresAt.IsZero() {
		t.Fatal("Expected non-zero expiration time")
	}

	// Verify new token is valid
	claims, err := jwtManager.ValidateToken(newAccessToken)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims["userId"] != float64(1) {
		t.Fatalf("Expected userId 1, got %v", claims["userId"])
	}
}

func TestAuthService_RefreshAccessToken_RejectsReusedToken(t *testing.T) {
	userRepo := NewMockUserRepository()
	auditRepo := NewMockAuditRepository()
	refreshRepo := NewMockRefreshTokenRepository()
	jwtCfg := &config.JWTConfig{
		Secret:           "test-secret-key-minimum-32-characters-long-!!!",
		AccessExpireMin:  15,
		RefreshExpireDay: 7,
	}
	jwtManager := auth.NewJWTManager(jwtCfg)
	passwordMgr := auth.NewPasswordManager()
	appLogger := logger.New("info")

	authService := newTestAuthService(userRepo, auditRepo, refreshRepo, jwtManager, passwordMgr, appLogger)
	ctx := context.Background()

	user := &models.User{
		ID:           1,
		Email:        "test@example.com",
		FirstName:    "Test",
		LastName:     "User",
		Role:         "admin",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	userRepo.users["test@example.com"] = user

	_, refreshToken, _, err := jwtManager.GenerateTokens(user)
	if err != nil {
		t.Fatalf("GenerateTokens failed: %v", err)
	}
	_, err = refreshRepo.Create(ctx, &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: hashToken(refreshToken),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Create refresh token failed: %v", err)
	}

	_, _, _, _, err = authService.RefreshAccessToken(ctx, refreshToken, "")
	if err != nil {
		t.Fatalf("First refresh failed: %v", err)
	}

	_, _, _, _, err = authService.RefreshAccessToken(ctx, refreshToken, "")
	if err == nil {
		t.Fatal("Expected reused refresh token to be rejected")
	}
}

func TestAuthService_Logout_RevokesUserRefreshTokens(t *testing.T) {
	userRepo := NewMockUserRepository()
	auditRepo := NewMockAuditRepository()
	refreshRepo := NewMockRefreshTokenRepository()
	jwtCfg := &config.JWTConfig{
		Secret:           "test-secret-key-minimum-32-characters-long-!!!",
		AccessExpireMin:  15,
		RefreshExpireDay: 7,
	}
	jwtManager := auth.NewJWTManager(jwtCfg)
	passwordMgr := auth.NewPasswordManager()
	appLogger := logger.New("info")

	authService := newTestAuthService(userRepo, auditRepo, refreshRepo, jwtManager, passwordMgr, appLogger)
	ctx := context.Background()

	_, err := refreshRepo.Create(ctx, &models.RefreshToken{
		UserID:    1,
		TokenHash: "token-hash",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Create refresh token failed: %v", err)
	}

	if err := authService.Logout(ctx, 1); err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	token, err := refreshRepo.GetByTokenHash(ctx, "token-hash")
	if err != nil {
		t.Fatalf("GetByTokenHash failed: %v", err)
	}
	if token.RevokedAt == nil {
		t.Fatal("Expected refresh token to be revoked")
	}
}
