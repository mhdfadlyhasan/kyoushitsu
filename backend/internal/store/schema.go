package store

import "fmt"

var migrations = []string{
	`
CREATE TABLE IF NOT EXISTS users (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	name          TEXT    NOT NULL,
	email         TEXT    NOT NULL UNIQUE,
	password_hash TEXT    NOT NULL,
	role          TEXT    NOT NULL CHECK (role IN ('super_admin', 'admin', 'teacher', 'student')),
	created_at    TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
`,
	`
CREATE UNIQUE INDEX IF NOT EXISTS users_email_normalized ON users (lower(email));
`,
}

// Migrate applies the database schema. It is safe to call on every boot.
func (s *Store) Migrate() error {
	for _, migration := range migrations {
		if _, err := s.DB.Exec(migration); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}
