package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

// ErrNotFound is returned when a requested record does not exist.
var ErrNotFound = errors.New("not found")

// ErrDuplicateEmail is returned when an email is already in use.
var ErrDuplicateEmail = errors.New("duplicate email")

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

// ListUsers returns all users ordered by creation time.
func (s *Store) ListUsers() ([]*User, error) {
	rows, err := s.DB.Query(`SELECT ` + userColumns + ` FROM users ORDER BY created_at, id`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]*User, 0)
	for rows.Next() {
		var user User
		if err := rows.Scan(
			&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, &user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

// CreateUser inserts a new user and returns it.
func (s *Store) CreateUser(name, email, passwordHash, role string) (*User, error) {
	var user User
	err := s.DB.QueryRow(
		`INSERT INTO users (name, email, password_hash, role)
		 VALUES ($1, $2, $3, $4)
		 RETURNING `+userColumns,
		name, NormalizeEmail(email), passwordHash, role,
	).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt)
	if isUniqueViolation(err) {
		return nil, ErrDuplicateEmail
	}
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &user, nil
}

// UpdateUser updates a user's name and role and returns the updated record.
func (s *Store) UpdateUser(id int64, name, role string) (*User, error) {
	var user User
	err := s.DB.QueryRow(
		`UPDATE users SET name = $2, role = $3 WHERE id = $1 RETURNING `+userColumns,
		id, name, role,
	).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return &user, nil
}

// ResetUserPassword replaces a user's password hash.
func (s *Store) ResetUserPassword(id int64, passwordHash string) error {
	res, err := s.DB.Exec(`UPDATE users SET password_hash = $2 WHERE id = $1`, id, passwordHash)
	if err != nil {
		return fmt.Errorf("reset password: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("reset password rows: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteUser removes a user by id.
func (s *Store) DeleteUser(id int64) error {
	res, err := s.DB.Exec(`DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete user rows: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
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
