package auth

import (
	"testing"
	"time"

	"healthcare-api/internal/config"
	"healthcare-api/internal/models"
)

// TestPasswordManager_HashPassword tests bcrypt password hashing
func TestPasswordManager_HashPassword(t *testing.T) {
	pm := NewPasswordManager()
	password := "TestPassword123!"

	hash, err := pm.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Fatal("Expected non-empty hash, got empty string")
	}

	if hash == password {
		t.Fatal("Password hash should not equal plaintext password")
	}
}

// TestPasswordManager_VerifyPassword_Success tests successful password verification
func TestPasswordManager_VerifyPassword_Success(t *testing.T) {
	pm := NewPasswordManager()
	password := "TestPassword123!"

	hash, err := pm.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if !pm.VerifyPassword(hash, password) {
		t.Fatal("Expected password verification to succeed")
	}
}

// TestPasswordManager_VerifyPassword_Failure tests failed password verification
func TestPasswordManager_VerifyPassword_Failure(t *testing.T) {
	pm := NewPasswordManager()
	password := "TestPassword123!"
	wrongPassword := "WrongPassword456!"

	hash, err := pm.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if pm.VerifyPassword(hash, wrongPassword) {
		t.Fatal("Expected password verification to fail with wrong password")
	}
}

// TestPasswordManager_VerifyPassword_EmptyPassword tests verification with empty password
func TestPasswordManager_VerifyPassword_EmptyPassword(t *testing.T) {
	pm := NewPasswordManager()
	password := "TestPassword123!"
	emptyPassword := ""

	hash, err := pm.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if pm.VerifyPassword(hash, emptyPassword) {
		t.Fatal("Expected password verification to fail with empty password")
	}
}

// TestJWTManager_GenerateTokens tests JWT token generation
func TestJWTManager_GenerateTokens(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:           "test-secret-key-minimum-32-characters-long-!!!",
		AccessExpireMin:  15,
		RefreshExpireDay: 7,
	}
	jm := NewJWTManager(cfg)

	// Mock user for testing
	user := &models.User{
		ID:    int64(1),
		Email: "test@example.com",
		Role:  "admin",
	}

	accessToken, refreshToken, expiresAt, err := jm.GenerateTokens(user)
	if err != nil {
		t.Fatalf("GenerateTokens failed: %v", err)
	}

	if accessToken == "" {
		t.Fatal("Expected non-empty access token")
	}

	if refreshToken == "" {
		t.Fatal("Expected non-empty refresh token")
	}

	if expiresAt.IsZero() {
		t.Fatal("Expected non-zero expiration time")
	}

	if expiresAt.Before(time.Now()) {
		t.Fatal("Expected expiration time to be in the future")
	}
}

// TestJWTManager_ValidateToken_Success tests successful token validation
func TestJWTManager_ValidateToken_Success(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:           "test-secret-key-minimum-32-characters-long-!!!",
		AccessExpireMin:  15,
		RefreshExpireDay: 7,
	}
	jm := NewJWTManager(cfg)

	user := &models.User{
		ID:    int64(1),
		Email: "test@example.com",
		Role:  "admin",
	}

	accessToken, _, _, err := jm.GenerateTokens(user)
	if err != nil {
		t.Fatalf("GenerateTokens failed: %v", err)
	}

	claims, err := jm.ValidateToken(accessToken)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims == nil {
		t.Fatal("Expected non-nil claims")
	}

	if claims["userId"] != float64(1) {
		t.Fatalf("Expected userId 1, got %v", claims["userId"])
	}

	if claims["email"] != "test@example.com" {
		t.Fatalf("Expected email test@example.com, got %v", claims["email"])
	}

	if claims["role"] != "admin" {
		t.Fatalf("Expected role admin, got %v", claims["role"])
	}
}

// TestJWTManager_ValidateToken_InvalidToken tests token validation with invalid token
func TestJWTManager_ValidateToken_InvalidToken(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:           "test-secret-key-minimum-32-characters-long-!!!",
		AccessExpireMin:  15,
		RefreshExpireDay: 7,
	}
	jm := NewJWTManager(cfg)

	invalidToken := "invalid.token.format"

	_, err := jm.ValidateToken(invalidToken)
	if err == nil {
		t.Fatal("Expected error for invalid token")
	}
}

// TestJWTManager_GetTokenType_AccessToken tests token type validation for access token
func TestJWTManager_GetTokenType_AccessToken(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:           "test-secret-key-minimum-32-characters-long-!!!",
		AccessExpireMin:  15,
		RefreshExpireDay: 7,
	}
	jm := NewJWTManager(cfg)

	user := &models.User{
		ID:    int64(1),
		Email: "test@example.com",
		Role:  "admin",
	}

	accessToken, _, _, err := jm.GenerateTokens(user)
	if err != nil {
		t.Fatalf("GenerateTokens failed: %v", err)
	}

	tokenType, err := jm.GetTokenType(accessToken)
	if err != nil {
		t.Fatalf("GetTokenType failed: %v", err)
	}

	if tokenType != "access" {
		t.Fatalf("Expected token type 'access', got %v", tokenType)
	}
}

// TestJWTManager_GetTokenType_RefreshToken tests token type validation for refresh token
func TestJWTManager_GetTokenType_RefreshToken(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:           "test-secret-key-minimum-32-characters-long-!!!",
		AccessExpireMin:  15,
		RefreshExpireDay: 7,
	}
	jm := NewJWTManager(cfg)

	user := &models.User{
		ID:    int64(1),
		Email: "test@example.com",
		Role:  "admin",
	}

	_, refreshToken, _, err := jm.GenerateTokens(user)
	if err != nil {
		t.Fatalf("GenerateTokens failed: %v", err)
	}

	tokenType, err := jm.GetTokenType(refreshToken)
	if err != nil {
		t.Fatalf("GetTokenType failed: %v", err)
	}

	if tokenType != "refresh" {
		t.Fatalf("Expected token type 'refresh', got %v", tokenType)
	}
}
