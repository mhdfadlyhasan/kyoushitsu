package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrNotFound is returned when a requested record does not exist.
var ErrNotFound = errors.New("not found")

// User is an application account.
type User struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

const userColumns = `id, name, email, password_hash, role, created_at`

// NormalizeEmail canonicalizes an email for storage and lookup.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// GetUserByEmail returns the user with the given email, case-insensitively.
func (s *Store) GetUserByEmail(email string) (*User, error) {
	return s.getUser(`SELECT `+userColumns+` FROM users WHERE lower(email) = ?`, NormalizeEmail(email))
}

// GetUserByID returns the user with the given id.
func (s *Store) GetUserByID(id int64) (*User, error) {
	return s.getUser(`SELECT `+userColumns+` FROM users WHERE id = ?`, id)
}

func (s *Store) getUser(query string, arg any) (*User, error) {
	var (
		user      User
		createdAt string
	)

	err := s.DB.QueryRow(query, arg).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role, &createdAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	user.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	return &user, nil
}
