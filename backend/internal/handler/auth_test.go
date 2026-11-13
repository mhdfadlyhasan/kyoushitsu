package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mhdfadlyhasan/study-companion/internal/auth"
	"github.com/mhdfadlyhasan/study-companion/internal/store"
)

func newTestRouter(t *testing.T) (*store.Store, http.Handler) {
	t.Helper()

	st, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	if err := st.Migrate(); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	return st, New(st, "test-secret")
}

func insertUser(t *testing.T, st *store.Store, name, email, password, role string) int64 {
	t.Helper()

	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	res, err := st.DB.Exec(
		`INSERT INTO users (name, email, password_hash, role) VALUES (?, ?, ?, ?)`,
		name, email, hash, role,
	)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("last insert id: %v", err)
	}
	return id
}

func doJSON(t *testing.T, h http.Handler, method, path string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()

	var reader *bytes.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(encoded)
	} else {
		reader = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestLoginSuccess(t *testing.T) {
	st, h := newTestRouter(t)
	id := insertUser(t, st, "Ana", "ana@example.com", "secret", auth.RoleTeacher)

	rec := doJSON(t, h, http.MethodPost, "/api/auth/login",
		map[string]string{"email": "  ANA@example.com ", "password": "secret"}, "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %s)", rec.Code, rec.Body)
	}
	if strings.Contains(rec.Body.String(), "password_hash") {
		t.Fatal("response must not expose password_hash")
	}

	var body loginResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Token == "" {
		t.Fatal("expected a token")
	}
	if body.User == nil || body.User.ID != id || body.User.Role != auth.RoleTeacher {
		t.Fatalf("user = %+v, want id %d role teacher", body.User, id)
	}

	claims, err := auth.ParseToken("test-secret", body.Token)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	if claims.Role != auth.RoleTeacher {
		t.Fatalf("token role = %q, want teacher", claims.Role)
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	st, h := newTestRouter(t)
	insertUser(t, st, "Ana", "ana@example.com", "secret", auth.RoleTeacher)

	tests := []struct {
		name string
		body map[string]string
	}{
		{"wrong password", map[string]string{"email": "ana@example.com", "password": "wrong"}},
		{"unknown email", map[string]string{"email": "nobody@example.com", "password": "secret"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doJSON(t, h, http.MethodPost, "/api/auth/login", tt.body, "")
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401", rec.Code)
			}
		})
	}
}

func TestLoginBadRequest(t *testing.T) {
	_, h := newTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader("{"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("malformed body status = %d, want 400", rec.Code)
	}

	rec = doJSON(t, h, http.MethodPost, "/api/auth/login", map[string]string{"email": ""}, "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("missing fields status = %d, want 400", rec.Code)
	}
}

func TestMe(t *testing.T) {
	st, h := newTestRouter(t)
	id := insertUser(t, st, "Ana", "ana@example.com", "secret", auth.RoleTeacher)

	login := doJSON(t, h, http.MethodPost, "/api/auth/login",
		map[string]string{"email": "ana@example.com", "password": "secret"}, "")
	var loginBody loginResponse
	if err := json.Unmarshal(login.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("decode login: %v", err)
	}

	rec := doJSON(t, h, http.MethodGet, "/api/me", nil, loginBody.Token)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %s)", rec.Code, rec.Body)
	}

	var user store.User
	if err := json.Unmarshal(rec.Body.Bytes(), &user); err != nil {
		t.Fatalf("decode me: %v", err)
	}
	if user.ID != id || user.Role != auth.RoleTeacher || user.Email != "ana@example.com" {
		t.Fatalf("user = %+v, want id %d role teacher", user, id)
	}
}

func TestMeRequiresToken(t *testing.T) {
	_, h := newTestRouter(t)

	if rec := doJSON(t, h, http.MethodGet, "/api/me", nil, ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestMeRejectsUnknownUser(t *testing.T) {
	_, h := newTestRouter(t)

	token, err := auth.IssueToken("test-secret", 999, auth.RoleStudent, time.Hour)
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}
	if rec := doJSON(t, h, http.MethodGet, "/api/me", nil, token); rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}
