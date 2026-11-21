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
	return s.getUser(`SELECT `+userColumns+` FROM users WHERE lower(email) = $1`, NormalizeEmail(email))
}

// GetUserByID returns the user with the given id.
func (s *Store) GetUserByID(id int64) (*User, error) {
	return s.getUser(`SELECT `+userColumns+` FROM users WHERE id = $1`, id)
}

// HasSuperAdmin reports whether a super admin account already exists.
func (s *Store) HasSuperAdmin() (bool, error) {
	var count int
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM users WHERE role = 'super_admin'`).Scan(&count); err != nil {
		return false, fmt.Errorf("count super admins: %w", err)
	}
	return count > 0, nil
}

// CreateSuperAdminIfNone inserts a super admin only when no super admin exists.
// It reports whether a row was created.
func (s *Store) CreateSuperAdminIfNone(name, email, passwordHash string) (bool, error) {
	res, err := s.DB.Exec(
		`INSERT INTO users (name, email, password_hash, role)
		 SELECT $1, $2, $3, 'super_admin'
		 WHERE NOT EXISTS (SELECT 1 FROM users WHERE role = 'super_admin')`,
		name, NormalizeEmail(email), passwordHash,
	)
	if err != nil {
		return false, fmt.Errorf("seed super admin: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("seed super admin rows: %w", err)
	}
	return affected > 0, nil
}

func (s *Store) getUser(query string, arg any) (*User, error) {
	var user User

	err := s.DB.QueryRow(query, arg).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	return &user, nil
}
