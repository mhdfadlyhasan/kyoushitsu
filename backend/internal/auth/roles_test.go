package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func serveWithRole(handler http.Handler, claims *Claims) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if claims != nil {
		req = req.WithContext(WithClaims(req.Context(), claims))
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestRequireRoleMissingClaims(t *testing.T) {
	handler := RequireRole(RoleTeacher)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next should not be called")
	}))

	if rec := serveWithRole(handler, nil); rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestRequireRoleRejectsWrongRole(t *testing.T) {
	handler := RequireRole(RoleTeacher, RoleAdmin)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next should not be called")
	}))

	if rec := serveWithRole(handler, &Claims{Role: RoleStudent}); rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestRequireRoleAllowsListedRole(t *testing.T) {
	called := false
	handler := RequireRole(RoleTeacher, RoleAdmin)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))

	if rec := serveWithRole(handler, &Claims{Role: RoleTeacher}); rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !called {
		t.Fatal("next was not called for an allowed role")
	}
}
