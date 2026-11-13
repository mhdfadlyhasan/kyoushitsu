package store

import (
	"errors"
	"testing"
)

func TestGetUserByEmailAndID(t *testing.T) {
	st := openTestStore(t)

	res, err := st.DB.Exec(
		`INSERT INTO users (name, email, password_hash, role) VALUES (?, ?, ?, ?)`,
		"Ana", "ana@example.com", "hash", "teacher",
	)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("last insert id: %v", err)
	}

	byEmail, err := st.GetUserByEmail("  ANA@example.com ")
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	if byEmail.ID != id || byEmail.Name != "Ana" || byEmail.Role != "teacher" {
		t.Fatalf("user = %+v, want id %d name Ana role teacher", byEmail, id)
	}
	if byEmail.CreatedAt.IsZero() {
		t.Fatal("expected created_at to be parsed")
	}

	byID, err := st.GetUserByID(id)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if byID.Email != "ana@example.com" {
		t.Fatalf("email = %q, want ana@example.com", byID.Email)
	}
}

func TestEmailIsCaseInsensitive(t *testing.T) {
	st := openTestStore(t)

	_, err := st.DB.Exec(
		`INSERT INTO users (name, email, password_hash, role) VALUES (?, ?, ?, ?)`,
		"Ana", "Ana@Example.com", "hash", "teacher",
	)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	user, err := st.GetUserByEmail(" ana@EXAMPLE.com ")
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	if user.Email != "Ana@Example.com" {
		t.Fatalf("email = %q, want Ana@Example.com", user.Email)
	}

	_, err = st.DB.Exec(
		`INSERT INTO users (name, email, password_hash, role) VALUES (?, ?, ?, ?)`,
		"Other", "ana@example.com", "hash", "student",
	)
	if err == nil {
		t.Fatal("expected case-insensitive duplicate email to be rejected")
	}
}

func TestNormalizeEmail(t *testing.T) {
	if got := NormalizeEmail("  Ana@Example.COM "); got != "ana@example.com" {
		t.Fatalf("NormalizeEmail = %q, want ana@example.com", got)
	}
}

func TestGetUserNotFound(t *testing.T) {
	st := openTestStore(t)

	if _, err := st.GetUserByEmail("missing@example.com"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetUserByEmail error = %v, want ErrNotFound", err)
	}
	if _, err := st.GetUserByID(999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetUserByID error = %v, want ErrNotFound", err)
	}
}
