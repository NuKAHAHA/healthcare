package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"healthcare-api/internal/config"
	"healthcare-api/internal/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	issuer   = "healthcare-api"
	audience = "healthcare-client"
)

type JWTManager struct {
	secret           string
	accessExpireMin  int
	refreshExpireDay int
}

func NewJWTManager(cfg *config.JWTConfig) *JWTManager {
	return &JWTManager{
		secret:           cfg.Secret,
		accessExpireMin:  cfg.AccessExpireMin,
		refreshExpireDay: cfg.RefreshExpireDay,
	}
}

// GenerateTokens creates a signed access + refresh token pair.
// Access tokens include jti (for blacklisting), iss, and aud claims.
func (jm *JWTManager) GenerateTokens(user *models.User) (accessToken, refreshToken string, expiresAt time.Time, err error) {
	now := time.Now()

	// ── Access token ──────────────────────────────────────────
	accessExpiration := now.Add(time.Duration(jm.accessExpireMin) * time.Minute)
	accessJTI, err := generateTokenID()
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to generate access token id: %w", err)
	}
	accessClaims := jwt.MapClaims{
		"userId": user.ID,
		"email":  user.Email,
		"role":   user.Role,
		"exp":    accessExpiration.Unix(),
		"iat":    now.Unix(),
		"iss":    issuer,
		"aud":    []string{audience},
		"type":   "access",
		"jti":    accessJTI,
	}
	accessJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessJWT.SignedString([]byte(jm.secret))
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to sign access token: %w", err)
	}

	// ── Refresh token ─────────────────────────────────────────
	refreshExpiration := now.Add(time.Duration(jm.refreshExpireDay) * 24 * time.Hour)
	refreshJTI, err := generateTokenID()
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to generate refresh token id: %w", err)
	}
	refreshClaims := jwt.MapClaims{
		"userId": user.ID,
		"email":  user.Email,
		"role":   user.Role,
		"exp":    refreshExpiration.Unix(),
		"iat":    now.Unix(),
		"iss":    issuer,
		"aud":    []string{audience},
		"type":   "refresh",
		"jti":    refreshJTI,
	}
	refreshJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshJWT.SignedString([]byte(jm.secret))
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return accessToken, refreshToken, accessExpiration, nil
}

func (jm *JWTManager) GetRefreshExpiration() time.Time {
	return time.Now().Add(time.Duration(jm.refreshExpireDay) * 24 * time.Hour)
}

// ValidateToken validates signature, expiry, signing method, issuer, and audience.
func (jm *JWTManager) ValidateToken(tokenString string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid token signing method: %v", token.Header["alg"])
		}
		return []byte(jm.secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validate issuer
	iss, _ := claims["iss"].(string)
	if iss != issuer {
		return nil, fmt.Errorf("invalid token issuer")
	}

	// Validate audience
	switch aud := claims["aud"].(type) {
	case string:
		if aud != audience {
			return nil, fmt.Errorf("invalid token audience")
		}
	case []interface{}:
		found := false
		for _, a := range aud {
			if s, ok := a.(string); ok && s == audience {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("invalid token audience")
		}
	default:
		return nil, fmt.Errorf("invalid token audience claim type")
	}

	return claims, nil
}

// GetJTI extracts the jti (JWT ID) claim from a raw token string.
func (jm *JWTManager) GetJTI(tokenString string) (string, error) {
	claims, err := jm.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}
	jti, ok := claims["jti"].(string)
	if !ok || jti == "" {
		return "", fmt.Errorf("jti not found in token")
	}
	return jti, nil
}

// GetExpirationTime returns the exp claim as time.Time.
func (jm *JWTManager) GetExpirationTime(tokenString string) (time.Time, error) {
	claims, err := jm.ValidateToken(tokenString)
	if err != nil {
		return time.Time{}, err
	}
	exp, ok := claims["exp"].(float64)
	if !ok {
		return time.Time{}, fmt.Errorf("exp not found in token")
	}
	return time.Unix(int64(exp), 0), nil
}

// GetTokenType returns "access" or "refresh".
func (jm *JWTManager) GetTokenType(tokenString string) (string, error) {
	claims, err := jm.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}
	tokenType, ok := claims["type"].(string)
	if !ok {
		return "", fmt.Errorf("token type not found")
	}
	return tokenType, nil
}

func generateTokenID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
