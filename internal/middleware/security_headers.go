package middleware

import "net/http"

// SecurityHeaders adds defensive HTTP headers to every response.
func SecurityHeaders() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("X-Frame-Options", "DENY")
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-XSS-Protection", "1; mode=block")
			h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			h.Set("Content-Security-Policy", "default-src 'self'")
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
			next.ServeHTTP(w, r)
		})
	}
}
