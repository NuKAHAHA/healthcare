package middleware

import (
	"encoding/json"
	"fmt"
	"healthcare-api/internal/logger"
	"healthcare-api/internal/models"
	"net/http"
	"runtime/debug"
)

// PanicRecovery recovers from panics and returns a server error
func PanicRecovery(logger *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log panic with stack trace
					logger.ErrorWithContext(
						"panic recovered",
						"panic_recovery",
						0,
						"path="+r.URL.Path+" error="+fmt.Sprint(err),
					)
					logger.Error("Stack trace: " + string(debug.Stack()))

					// Return generic error response
					sendError(w, http.StatusInternalServerError, "internal_error", "An error occurred")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// sendError sends an error response in JSON format
func sendError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := models.ErrorResponse{
		Message: message,
		Code:    code,
	}

	_ = json.NewEncoder(w).Encode(response)
}
