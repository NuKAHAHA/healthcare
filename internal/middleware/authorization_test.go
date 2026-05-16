package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"healthcare-api/internal/logger"
	"healthcare-api/internal/models"
)

var testLogger = logger.New("error") // suppress log output during tests

// TestRoleBasedAuthorization_AllowedRole tests successful role authorization
func TestRoleBasedAuthorization_AllowedRole(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Create middleware that allows 'admin' role
	middleware := RoleBasedAuthorization(testLogger, "admin")

	// Create request with admin user context
	req := httptest.NewRequest("GET", "/protected", nil)
	req = setAuthUserContext(req, &models.AuthUser{
		ID:    1,
		Email: "admin@example.com",
		Role:  "admin",
	})

	// Test
	w := httptest.NewRecorder()
	middleware(handler).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status code 200, got %d", w.Code)
	}
}

// TestRoleBasedAuthorization_ForbiddenRole tests denied role authorization
func TestRoleBasedAuthorization_ForbiddenRole(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Create middleware that allows only 'admin' role
	middleware := RoleBasedAuthorization(testLogger, "admin")

	// Create request with doctor user context
	req := httptest.NewRequest("GET", "/protected", nil)
	req = setAuthUserContext(req, &models.AuthUser{
		ID:    2,
		Email: "doctor@example.com",
		Role:  "doctor",
	})

	// Test
	w := httptest.NewRecorder()
	middleware(handler).ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("Expected status code 403, got %d", w.Code)
	}
}

// TestRoleBasedAuthorization_MultipleRoles tests authorization with multiple allowed roles
func TestRoleBasedAuthorization_MultipleRoles(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Create middleware that allows 'admin' or 'registrar' roles
	middleware := RoleBasedAuthorization(testLogger, "admin", "registrar")

	// Test with registrar
	req := httptest.NewRequest("GET", "/protected", nil)
	req = setAuthUserContext(req, &models.AuthUser{
		ID:    3,
		Email: "registrar@example.com",
		Role:  "registrar",
	})

	w := httptest.NewRecorder()
	middleware(handler).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status code 200, got %d", w.Code)
	}

	// Test with doctor (should be denied)
	req = httptest.NewRequest("GET", "/protected", nil)
	req = setAuthUserContext(req, &models.AuthUser{
		ID:    2,
		Email: "doctor@example.com",
		Role:  "doctor",
	})

	w = httptest.NewRecorder()
	middleware(handler).ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("Expected status code 403, got %d", w.Code)
	}
}

// TestRoleBasedAuthorization_NoUserContext tests authorization when no user context exists
func TestRoleBasedAuthorization_NoUserContext(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	middleware := RoleBasedAuthorization(testLogger, "admin")

	// Create request without user context
	req := httptest.NewRequest("GET", "/protected", nil)

	w := httptest.NewRecorder()
	middleware(handler).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Expected status code 401, got %d", w.Code)
	}
}

// TestObjectLevelAuthorization_Allowed tests successful object-level authorization
func TestObjectLevelAuthorization_Allowed(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Create check function that always allows
	checkFunc := func(user *models.AuthUser, r *http.Request) bool {
		return true
	}

	middleware := ObjectLevelAuthorization(testLogger, checkFunc)

	// Create request with user context
	req := httptest.NewRequest("GET", "/resource/1", nil)
	req = setAuthUserContext(req, &models.AuthUser{
		ID:    1,
		Email: "user@example.com",
		Role:  "admin",
	})

	w := httptest.NewRecorder()
	middleware(handler).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status code 200, got %d", w.Code)
	}
}

// TestObjectLevelAuthorization_Denied tests denied object-level authorization
func TestObjectLevelAuthorization_Denied(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Create check function that always denies
	checkFunc := func(user *models.AuthUser, r *http.Request) bool {
		return false
	}

	middleware := ObjectLevelAuthorization(testLogger, checkFunc)

	// Create request with user context
	req := httptest.NewRequest("GET", "/resource/1", nil)
	req = setAuthUserContext(req, &models.AuthUser{
		ID:    1,
		Email: "user@example.com",
		Role:  "admin",
	})

	w := httptest.NewRecorder()
	middleware(handler).ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("Expected status code 403, got %d", w.Code)
	}
}

// Helper function to set AuthUser in request context
func setAuthUserContext(r *http.Request, user *models.AuthUser) *http.Request {
	ctx := context.WithValue(r.Context(), models.UserContextKey, user)
	return r.WithContext(ctx)
}
