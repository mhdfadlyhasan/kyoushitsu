package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mhdfadlyhasan/study-companion/internal/auth"
	"github.com/mhdfadlyhasan/study-companion/internal/store"
)

func tokenFor(t *testing.T, id int64, role string) string {
	t.Helper()

	token, err := auth.IssueToken("test-secret", id, role, time.Hour)
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}
	return token
}

func TestAdminListUsers(t *testing.T) {
	st, h := newTestRouter(t)
	superID := insertUser(t, st, "Owner", "owner@example.com", "secret", auth.RoleSuperAdmin)
	insertUser(t, st, "Ana", "ana@example.com", "secret", auth.RoleTeacher)
	insertUser(t, st, "Bob", "bob@example.com", "secret", auth.RoleStudent)

	rec := doJSON(t, h, http.MethodGet, "/api/admin/users", nil, tokenFor(t, superID, auth.RoleSuperAdmin))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %s)", rec.Code, rec.Body)
	}
	if strings.Contains(rec.Body.String(), "password_hash") {
		t.Fatal("response must not expose password_hash")
	}

	var users []store.User
	if err := json.Unmarshal(rec.Body.Bytes(), &users); err != nil {
		t.Fatalf("decode users: %v", err)
	}
	if len(users) != 3 {
		t.Fatalf("len(users) = %d, want 3", len(users))
	}

	adminID := insertUser(t, st, "Admin", "admin@example.com", "secret", auth.RoleAdmin)
	if rec := doJSON(t, h, http.MethodGet, "/api/admin/users", nil, tokenFor(t, adminID, auth.RoleAdmin)); rec.Code != http.StatusOK {
		t.Fatalf("admin status = %d, want 200", rec.Code)
	}

	teacherID := users[1].ID
	if rec := doJSON(t, h, http.MethodGet, "/api/admin/users", nil, tokenFor(t, teacherID, auth.RoleTeacher)); rec.Code != http.StatusForbidden {
		t.Fatalf("teacher status = %d, want 403", rec.Code)
	}
	studentID := users[2].ID
	if rec := doJSON(t, h, http.MethodGet, "/api/admin/users", nil, tokenFor(t, studentID, auth.RoleStudent)); rec.Code != http.StatusForbidden {
		t.Fatalf("student status = %d, want 403", rec.Code)
	}
	if rec := doJSON(t, h, http.MethodGet, "/api/admin/users", nil, ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous status = %d, want 401", rec.Code)
	}
}

func TestAdminCreateUserRoleRules(t *testing.T) {
	st, h := newTestRouter(t)
	superID := insertUser(t, st, "Owner", "owner@example.com", "secret", auth.RoleSuperAdmin)
	adminID := insertUser(t, st, "Admin", "admin@example.com", "secret", auth.RoleAdmin)

	super := tokenFor(t, superID, auth.RoleSuperAdmin)
	admin := tokenFor(t, adminID, auth.RoleAdmin)

	rec := doJSON(t, h, http.MethodPost, "/api/admin/users", map[string]string{
		"name": "Ana", "email": " ANA@example.com ", "password": "password1", "role": auth.RoleTeacher,
	}, admin)
	if rec.Code != http.StatusCreated {
		t.Fatalf("admin create teacher status = %d, want 201 (body %s)", rec.Code, rec.Body)
	}
	var created store.User
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created: %v", err)
	}
	if created.Email != "ana@example.com" || created.Role != auth.RoleTeacher {
		t.Fatalf("created = %+v, want normalized teacher", created)
	}

	if rec := doJSON(t, h, http.MethodPost, "/api/admin/users", map[string]string{
		"name": "Bob", "email": "bob@example.com", "password": "password1", "role": auth.RoleStudent,
	}, admin); rec.Code != http.StatusCreated {
		t.Fatalf("admin create student status = %d, want 201", rec.Code)
	}
	if rec := doJSON(t, h, http.MethodPost, "/api/admin/users", map[string]string{
		"name": "Eve", "email": "eve@example.com", "password": "password1", "role": auth.RoleAdmin,
	}, admin); rec.Code != http.StatusForbidden {
		t.Fatalf("admin create admin status = %d, want 403", rec.Code)
	}
	if rec := doJSON(t, h, http.MethodPost, "/api/admin/users", map[string]string{
		"name": "Eve", "email": "eve@example.com", "password": "password1", "role": auth.RoleAdmin,
	}, super); rec.Code != http.StatusCreated {
		t.Fatalf("super create admin status = %d, want 201 (body %s)", rec.Code, rec.Body)
	}

	badCases := []struct {
		name string
		body map[string]string
		want int
	}{
		{"super admin role", map[string]string{"name": "X", "email": "x@example.com", "password": "password1", "role": auth.RoleSuperAdmin}, http.StatusBadRequest},
		{"invalid role", map[string]string{"name": "X", "email": "x@example.com", "password": "password1", "role": "owner"}, http.StatusBadRequest},
		{"short password", map[string]string{"name": "X", "email": "x@example.com", "password": "short", "role": auth.RoleStudent}, http.StatusBadRequest},
		{"missing name", map[string]string{"email": "x@example.com", "password": "password1", "role": auth.RoleStudent}, http.StatusBadRequest},
		{"missing email", map[string]string{"name": "X", "password": "password1", "role": auth.RoleStudent}, http.StatusBadRequest},
		{"duplicate email", map[string]string{"name": "X", "email": "ana@example.com", "password": "password1", "role": auth.RoleStudent}, http.StatusConflict},
	}
	for _, tc := range badCases {
		t.Run(tc.name, func(t *testing.T) {
			if rec := doJSON(t, h, http.MethodPost, "/api/admin/users", tc.body, super); rec.Code != tc.want {
				t.Fatalf("status = %d, want %d (body %s)", rec.Code, tc.want, rec.Body)
			}
		})
	}
}

func TestAdminUpdateUserGuards(t *testing.T) {
	st, h := newTestRouter(t)
	superID := insertUser(t, st, "Owner", "owner@example.com", "secret", auth.RoleSuperAdmin)
	adminID := insertUser(t, st, "Admin", "admin@example.com", "secret", auth.RoleAdmin)
	teacherID := insertUser(t, st, "Ana", "ana@example.com", "secret", auth.RoleTeacher)

	super := tokenFor(t, superID, auth.RoleSuperAdmin)
	admin := tokenFor(t, adminID, auth.RoleAdmin)

	rec := doJSON(t, h, http.MethodPatch, "/api/admin/users/"+strconv.FormatInt(teacherID, 10),
		map[string]string{"name": "Ana B", "role": auth.RoleStudent}, super)
	if rec.Code != http.StatusOK {
		t.Fatalf("super update status = %d, want 200 (body %s)", rec.Code, rec.Body)
	}

	if rec := doJSON(t, h, http.MethodPatch, "/api/admin/users/"+strconv.FormatInt(teacherID, 10),
		map[string]string{"name": "Ana C"}, admin); rec.Code != http.StatusOK {
		t.Fatalf("admin update status = %d, want 200 (body %s)", rec.Code, rec.Body)
	}
	if rec := doJSON(t, h, http.MethodPatch, "/api/admin/users/"+strconv.FormatInt(teacherID, 10),
		map[string]string{"role": auth.RoleAdmin}, admin); rec.Code != http.StatusForbidden {
		t.Fatalf("admin promote status = %d, want 403", rec.Code)
	}
	if rec := doJSON(t, h, http.MethodPatch, "/api/admin/users/"+strconv.FormatInt(adminID, 10),
		map[string]string{"name": "Nope"}, admin); rec.Code != http.StatusForbidden {
		t.Fatalf("admin modify admin status = %d, want 403", rec.Code)
	}
	if rec := doJSON(t, h, http.MethodPatch, "/api/admin/users/"+strconv.FormatInt(adminID, 10),
		map[string]string{"name": "Admin B"}, super); rec.Code != http.StatusOK {
		t.Fatalf("super modify admin status = %d, want 200 (body %s)", rec.Code, rec.Body)
	}
	if rec := doJSON(t, h, http.MethodPatch, "/api/admin/users/"+strconv.FormatInt(superID, 10),
		map[string]string{"name": "Owner B"}, super); rec.Code != http.StatusOK {
		t.Fatalf("super self rename status = %d, want 200 (body %s)", rec.Code, rec.Body)
	}
	if rec := doJSON(t, h, http.MethodPatch, "/api/admin/users/"+strconv.FormatInt(superID, 10),
		map[string]string{"role": auth.RoleStudent}, super); rec.Code != http.StatusForbidden {
		t.Fatalf("super self demote status = %d, want 403", rec.Code)
	}
	if rec := doJSON(t, h, http.MethodPatch, "/api/admin/users/"+strconv.FormatInt(teacherID, 10),
		map[string]string{"role": auth.RoleSuperAdmin}, super); rec.Code != http.StatusBadRequest {
		t.Fatalf("assign super admin status = %d, want 400", rec.Code)
	}
	if rec := doJSON(t, h, http.MethodPatch, "/api/admin/users/"+strconv.FormatInt(teacherID, 10),
		map[string]string{}, super); rec.Code != http.StatusBadRequest {
		t.Fatalf("empty patch status = %d, want 400", rec.Code)
	}
	if rec := doJSON(t, h, http.MethodPatch, "/api/admin/users/999999",
		map[string]string{"name": "Ghost"}, super); rec.Code != http.StatusNotFound {
		t.Fatalf("missing user status = %d, want 404", rec.Code)
	}
}

func TestAdminResetPassword(t *testing.T) {
	st, h := newTestRouter(t)
	superID := insertUser(t, st, "Owner", "owner@example.com", "secret", auth.RoleSuperAdmin)
	adminID := insertUser(t, st, "Admin", "admin@example.com", "secret", auth.RoleAdmin)
	teacherID := insertUser(t, st, "Ana", "ana@example.com", "secret", auth.RoleTeacher)

	super := tokenFor(t, superID, auth.RoleSuperAdmin)
	admin := tokenFor(t, adminID, auth.RoleAdmin)

	if rec := doJSON(t, h, http.MethodPost, "/api/admin/users/"+strconv.FormatInt(teacherID, 10)+"/password",
		map[string]string{"password": "brand-new-pass"}, admin); rec.Code != http.StatusNoContent {
		t.Fatalf("reset status = %d, want 204 (body %s)", rec.Code, rec.Body)
	}

	login := doJSON(t, h, http.MethodPost, "/api/auth/login",
		map[string]string{"email": "ana@example.com", "password": "brand-new-pass"}, "")
	if login.Code != http.StatusOK {
		t.Fatalf("login after reset status = %d, want 200", login.Code)
	}

	if rec := doJSON(t, h, http.MethodPost, "/api/admin/users/"+strconv.FormatInt(adminID, 10)+"/password",
		map[string]string{"password": "brand-new-pass"}, admin); rec.Code != http.StatusForbidden {
		t.Fatalf("admin reset admin status = %d, want 403", rec.Code)
	}
	if rec := doJSON(t, h, http.MethodPost, "/api/admin/users/"+strconv.FormatInt(superID, 10)+"/password",
		map[string]string{"password": "brand-new-pass"}, super); rec.Code != http.StatusNoContent {
		t.Fatalf("super self reset status = %d, want 204", rec.Code)
	}
	if rec := doJSON(t, h, http.MethodPost, "/api/admin/users/"+strconv.FormatInt(teacherID, 10)+"/password",
		map[string]string{"password": "short"}, super); rec.Code != http.StatusBadRequest {
		t.Fatalf("short password status = %d, want 400", rec.Code)
	}
}

func TestAdminDeleteUserGuards(t *testing.T) {
	st, h := newTestRouter(t)
	superID := insertUser(t, st, "Owner", "owner@example.com", "secret", auth.RoleSuperAdmin)
	adminID := insertUser(t, st, "Admin", "admin@example.com", "secret", auth.RoleAdmin)
	teacherID := insertUser(t, st, "Ana", "ana@example.com", "secret", auth.RoleTeacher)
	otherAdminID := insertUser(t, st, "Admin2", "admin2@example.com", "secret", auth.RoleAdmin)

	super := tokenFor(t, superID, auth.RoleSuperAdmin)
	admin := tokenFor(t, adminID, auth.RoleAdmin)

	if rec := doJSON(t, h, http.MethodDelete, "/api/admin/users/"+strconv.FormatInt(teacherID, 10), nil, admin); rec.Code != http.StatusNoContent {
		t.Fatalf("admin delete teacher status = %d, want 204 (body %s)", rec.Code, rec.Body)
	}
	if rec := doJSON(t, h, http.MethodDelete, "/api/admin/users/"+strconv.FormatInt(superID, 10), nil, super); rec.Code != http.StatusForbidden {
		t.Fatalf("delete super status = %d, want 403", rec.Code)
	}
	if rec := doJSON(t, h, http.MethodDelete, "/api/admin/users/"+strconv.FormatInt(adminID, 10), nil, admin); rec.Code != http.StatusForbidden {
		t.Fatalf("self delete status = %d, want 403", rec.Code)
	}
	if rec := doJSON(t, h, http.MethodDelete, "/api/admin/users/"+strconv.FormatInt(otherAdminID, 10), nil, admin); rec.Code != http.StatusForbidden {
		t.Fatalf("admin delete admin status = %d, want 403", rec.Code)
	}
	if rec := doJSON(t, h, http.MethodDelete, "/api/admin/users/"+strconv.FormatInt(otherAdminID, 10), nil, super); rec.Code != http.StatusNoContent {
		t.Fatalf("super delete admin status = %d, want 204", rec.Code)
	}
	if rec := doJSON(t, h, http.MethodDelete, "/api/admin/users/999999", nil, super); rec.Code != http.StatusNotFound {
		t.Fatalf("delete missing status = %d, want 404", rec.Code)
	}
}
