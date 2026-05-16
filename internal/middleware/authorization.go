package middleware

import (
	"fmt"
	"healthcare-api/internal/logger"
	"healthcare-api/internal/models"
	"net/http"
)

// RoleBasedAuthorization enforces role membership.
// 403 Forbidden attempts are logged including user ID, endpoint, and method.
func RoleBasedAuthorization(appLogger *logger.Logger, allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := GetAuthUser(r)
			if err != nil {
				sendError(w, http.StatusUnauthorized, "unauthorized", "User not authenticated")
				return
			}

			for _, role := range allowedRoles {
				if user.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			appLogger.WarnWithContext(
				"access denied — insufficient role",
				"authz_failure",
				user.ID,
				fmt.Sprintf("role=%s required=%v endpoint=%s method=%s ip=%s",
					user.Role, allowedRoles, r.URL.Path, r.Method, ClientIP(r)),
			)
			sendError(w, http.StatusForbidden, "forbidden", "Insufficient permissions")
		})
	}
}

// ObjectLevelAuthorization checks resource-level ownership.
func ObjectLevelAuthorization(appLogger *logger.Logger, checkFunc func(*models.AuthUser, *http.Request) bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := GetAuthUser(r)
			if err != nil {
				sendError(w, http.StatusUnauthorized, "unauthorized", "User not authenticated")
				return
			}

			if !checkFunc(user, r) {
				appLogger.WarnWithContext(
					"object-level access denied",
					"authz_failure",
					user.ID,
					fmt.Sprintf("endpoint=%s method=%s ip=%s", r.URL.Path, r.Method, ClientIP(r)),
				)
				sendError(w, http.StatusForbidden, "forbidden", "Access denied")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
