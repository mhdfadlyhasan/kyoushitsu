package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/mhdfadlyhasan/study-companion/internal/auth"
	"github.com/mhdfadlyhasan/study-companion/internal/store"
)

const minPasswordLength = 8

type actorIdentity struct {
	id   int64
	role string
}

type createUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type updateUserRequest struct {
	Name *string `json:"name"`
	Role *string `json:"role"`
}

type resetPasswordRequest struct {
	Password string `json:"password"`
}

func (h *handler) listUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.store.ListUsers()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load users")
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func (h *handler) createUser(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	name := strings.TrimSpace(req.Name)
	email := store.NormalizeEmail(req.Email)
	role := req.Role

	switch {
	case name == "":
		writeError(w, http.StatusBadRequest, "name is required")
		return
	case email == "":
		writeError(w, http.StatusBadRequest, "email is required")
		return
	case len(req.Password) < minPasswordLength:
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	case role == auth.RoleSuperAdmin:
		writeError(w, http.StatusBadRequest, "cannot create a super admin")
		return
	case !isAssignableRole(role):
		writeError(w, http.StatusBadRequest, "invalid role")
		return
	case role == auth.RoleAdmin && actor.role != auth.RoleSuperAdmin:
		writeError(w, http.StatusForbidden, "only the super admin can create admins")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not hash password")
		return
	}

	user, err := h.store.CreateUser(name, email, hash, role)
	if errors.Is(err, store.ErrDuplicateEmail) {
		writeError(w, http.StatusConflict, "email is already in use")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create user")
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func (h *handler) updateUser(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, ok := pathID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req updateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == nil && req.Role == nil {
		writeError(w, http.StatusBadRequest, "nothing to update")
		return
	}

	target, err := h.store.GetUserByID(id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load user")
		return
	}
	if !canManageUser(actor, target) {
		writeError(w, http.StatusForbidden, "not allowed to manage this user")
		return
	}

	name := target.Name
	if req.Name != nil {
		name = strings.TrimSpace(*req.Name)
		if name == "" {
			writeError(w, http.StatusBadRequest, "name is required")
			return
		}
	}

	role := target.Role
	if req.Role != nil {
		newRole := *req.Role
		switch {
		case target.Role == auth.RoleSuperAdmin && newRole != auth.RoleSuperAdmin:
			writeError(w, http.StatusForbidden, "the super admin cannot be demoted")
			return
		case target.Role != auth.RoleSuperAdmin && newRole == auth.RoleSuperAdmin:
			writeError(w, http.StatusBadRequest, "cannot assign the super admin role")
			return
		case target.Role != auth.RoleSuperAdmin && !isAssignableRole(newRole):
			writeError(w, http.StatusBadRequest, "invalid role")
			return
		case newRole == auth.RoleAdmin && actor.role != auth.RoleSuperAdmin:
			writeError(w, http.StatusForbidden, "only the super admin can promote to admin")
			return
		case actor.id == target.ID && newRole != target.Role:
			writeError(w, http.StatusForbidden, "you cannot change your own role")
			return
		}
		role = newRole
	}

	user, err := h.store.UpdateUser(id, name, role)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not update user")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *handler) resetUserPassword(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, ok := pathID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req resetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.Password) < minPasswordLength {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	target, err := h.store.GetUserByID(id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load user")
		return
	}
	if !canManageUser(actor, target) {
		writeError(w, http.StatusForbidden, "not allowed to manage this user")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not hash password")
		return
	}

	if err := h.store.ResetUserPassword(id, hash); err != nil {
		writeError(w, http.StatusInternalServerError, "could not reset password")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, ok := pathID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	target, err := h.store.GetUserByID(id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load user")
		return
	}
	if target.Role == auth.RoleSuperAdmin {
		writeError(w, http.StatusForbidden, "the super admin cannot be deleted")
		return
	}
	if actor.id == target.ID {
		writeError(w, http.StatusForbidden, "you cannot delete your own account")
		return
	}
	if target.Role == auth.RoleAdmin && actor.role != auth.RoleSuperAdmin {
		writeError(w, http.StatusForbidden, "only the super admin can delete admins")
		return
	}

	if err := h.store.DeleteUser(id); err != nil {
		writeError(w, http.StatusInternalServerError, "could not delete user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func actorFromRequest(r *http.Request) (actorIdentity, bool) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		return actorIdentity{}, false
	}
	id, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return actorIdentity{}, false
	}
	return actorIdentity{id: id, role: claims.Role}, true
}

// canManageUser reports whether actor may edit or reset target. Admins manage
// teachers and students; the super admin also manages admins and themselves.
func canManageUser(actor actorIdentity, target *store.User) bool {
	switch target.Role {
	case auth.RoleSuperAdmin:
		return actor.role == auth.RoleSuperAdmin && actor.id == target.ID
	case auth.RoleAdmin:
		return actor.role == auth.RoleSuperAdmin
	default:
		return actor.role == auth.RoleSuperAdmin || actor.role == auth.RoleAdmin
	}
}

func isAssignableRole(role string) bool {
	return role == auth.RoleAdmin || role == auth.RoleTeacher || role == auth.RoleStudent
}

func pathID(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	return id, err == nil && id > 0
}
