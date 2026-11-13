package config

import (
	"os"
)

// Config holds runtime configuration, sourced from environment variables.
type Config struct {
	Addr               string // HTTP listen address, e.g. ":8080"
	DBPath             string // SQLite database file path
	JWTSecret          string // Secret used to sign JWTs
	SuperadminEmail    string // Email of the seeded super admin
	SuperadminPassword string // Plaintext super admin password, hashed at seed time
}

// Load reads configuration from the environment, applying sensible defaults
// for local development.
func Load() Config {
	return Config{
		Addr:               getenv("ADDR", ":8080"),
		DBPath:             getenv("DB_PATH", "study-companion.db"),
		JWTSecret:          getenv("JWT_SECRET", "dev-secret-change-me"),
		SuperadminEmail:    getenv("SUPERADMIN_EMAIL", ""),
		SuperadminPassword: getenv("SUPERADMIN_PASSWORD", ""),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
