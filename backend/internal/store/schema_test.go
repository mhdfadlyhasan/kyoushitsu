package store

import (
	"testing"

	"github.com/mhdfadlyhasan/study-companion/internal/testsupport"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()

	st, err := Open(testsupport.PostgresURL(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	if err := st.Migrate(); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	return st
}

func TestMigrateUsersTable(t *testing.T) {
	st := openTestStore(t)

	if err := st.Migrate(); err != nil {
		t.Fatalf("second Migrate should be idempotent: %v", err)
	}

	_, err := st.DB.Exec(
		`INSERT INTO users (name, email, password_hash, role) VALUES ($1, $2, $3, $4)`,
		"Alice", "alice@example.com", "hash", "teacher",
	)
	if err != nil {
		t.Fatalf("insert valid user: %v", err)
	}

	_, err = st.DB.Exec(
		`INSERT INTO users (name, email, password_hash, role) VALUES ($1, $2, $3, $4)`,
		"Bob", "bob@example.com", "hash", "owner",
	)
	if err == nil {
		t.Fatal("expected invalid role to be rejected")
	}

	_, err = st.DB.Exec(
		`INSERT INTO users (name, email, password_hash, role) VALUES ($1, $2, $3, $4)`,
		"Duplicate", "alice@example.com", "hash", "student",
	)
	if err == nil {
		t.Fatal("expected duplicate email to be rejected")
	}
}
