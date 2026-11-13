package store

import "fmt"

const schema = `
CREATE TABLE IF NOT EXISTS users (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	name          TEXT    NOT NULL,
	email         TEXT    NOT NULL UNIQUE,
	password_hash TEXT    NOT NULL,
	role          TEXT    NOT NULL CHECK (role IN ('super_admin', 'admin', 'teacher', 'student')),
	created_at    TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
`

// Migrate applies the database schema. It is safe to call on every boot.
func (s *Store) Migrate() error {
	if _, err := s.DB.Exec(schema); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}
