package middleware

import (
	"context"
	"fmt"
	"healthcare-api/internal/auth"
	"healthcare-api/internal/logger"
	"healthcare-api/internal/models"
	"net/http"
	"strings"
)

// AuthMiddleware validates JWT access tokens and checks the token blacklist.
// Requests with revoked tokens (logged-out sessions) are rejected with 401.
func AuthMiddleware(jwtManager *auth.JWTManager, blacklist *auth.TokenBlacklist, appLogger *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				ip := ClientIP(r)
				appLogger.WarnWithContext("missing authorization header", "auth_failure",
					0, fmt.Sprintf("ip=%s endpoint=%s method=%s reason=missing_header", ip, r.URL.Path, r.Method))
				sendError(w, http.StatusUnauthorized, "unauthorized", "Missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				ip := ClientIP(r)
				appLogger.WarnWithContext("invalid authorization header format", "auth_failure",
					0, fmt.Sprintf("ip=%s endpoint=%s reason=bad_header_format", ip, r.URL.Path))
				sendError(w, http.StatusUnauthorized, "unauthorized", "Invalid authorization header")
				return
			}

			token := parts[1]

			claims, err := jwtManager.ValidateToken(token)
			if err != nil {
				ip := ClientIP(r)
				appLogger.WarnWithContext("invalid token", "auth_failure",
					0, fmt.Sprintf("ip=%s endpoint=%s reason=%s", ip, r.URL.Path, err.Error()))
				sendError(w, http.StatusUnauthorized, "unauthorized", "Invalid token")
				return
			}

			tokenType, ok := claims["type"].(string)
			if !ok || tokenType != "access" {
				appLogger.WarnWithContext("wrong token type used", "auth_failure",
					0, fmt.Sprintf("endpoint=%s type=%v", r.URL.Path, claims["type"]))
				sendError(w, http.StatusUnauthorized, "unauthorized", "Invalid token type")
				return
			}

			// ── Blacklist check ───────────────────────────────────
			jti, _ := claims["jti"].(string)
			if jti != "" && blacklist != nil {
				revoked, err := blacklist.IsRevoked(r.Context(), jti)
				if err != nil {
					// Redis unavailable — fail open, log warning
					appLogger.WarnWithContext("blacklist check failed (Redis unavailable)", "auth_warning",
						0, fmt.Sprintf("jti=%s err=%s", jti, err.Error()))
				} else if revoked {
					appLogger.WarnWithContext("revoked token used", "auth_failure",
						0, fmt.Sprintf("jti=%s endpoint=%s", jti, r.URL.Path))
					sendError(w, http.StatusUnauthorized, "unauthorized", "Token has been revoked")
					return
				}
			}

			userID, ok := claims["userId"].(float64)
			if !ok {
				sendError(w, http.StatusUnauthorized, "unauthorized", "Invalid token")
				return
			}
			email, _ := claims["email"].(string)
			role, _ := claims["role"].(string)

			authUser := &models.AuthUser{
				ID:    int64(userID),
				Email: email,
				Role:  role,
			}
			ctx := context.WithValue(r.Context(), models.UserContextKey, authUser)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetAuthUser extracts the authenticated user from request context.
func GetAuthUser(r *http.Request) (*models.AuthUser, error) {
	user, ok := r.Context().Value(models.UserContextKey).(*models.AuthUser)
	if !ok {
		return nil, fmt.Errorf("user not found in context")
	}
	return user, nil
}
