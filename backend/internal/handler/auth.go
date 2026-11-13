package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mhdfadlyhasan/study-companion/internal/auth"
	"github.com/mhdfadlyhasan/study-companion/internal/store"
)

const tokenTTL = 24 * time.Hour

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string      `json:"token"`
	User  *store.User `json:"user"`
}

func (h *handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	email := strings.TrimSpace(req.Email)
	if email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	user, err := h.store.GetUserByEmail(email)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load user")
		return
	}
	if !auth.CheckPassword(user.PasswordHash, req.Password) {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	token, err := auth.IssueToken(h.jwtSecret, user.ID, user.Role, tokenTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create token")
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{Token: token, User: user})
}

func (h *handler) me(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.store.GetUserByID(id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load user")
		return
	}

	writeJSON(w, http.StatusOK, user)
}
