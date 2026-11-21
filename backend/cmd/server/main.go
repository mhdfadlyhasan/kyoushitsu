package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/mhdfadlyhasan/study-companion/internal/auth"
	"github.com/mhdfadlyhasan/study-companion/internal/config"
	"github.com/mhdfadlyhasan/study-companion/internal/handler"
	"github.com/mhdfadlyhasan/study-companion/internal/store"
)

const superAdminName = "Super Admin"

func main() {
	if len(os.Args) == 1 {
		runServer()
		return
	}

	if len(os.Args) == 2 && os.Args[1] == "seed-superadmin" {
		if err := runSeedSuperAdmin(); err != nil {
			log.Fatalf("seed super admin: %v", err)
		}
		return
	}

	log.Fatalf("usage: %s [seed-superadmin]", os.Args[0])
}

func runServer() {
	cfg := config.Load()

	st, err := openMigratedStore(cfg)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer st.Close()

	exists, err := st.HasSuperAdmin()
	if err != nil {
		log.Fatalf("check super admin: %v", err)
	}
	if !exists {
		log.Printf("warning: no super admin exists; run `%s seed-superadmin` to create one", os.Args[0])
	}

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: handler.New(st, cfg.JWTSecret),
	}

	log.Printf("study-companion listening on %s (db=postgres)", cfg.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
}

func runSeedSuperAdmin() error {
	cfg := config.Load()

	st, err := openMigratedStore(cfg)
	if err != nil {
		return err
	}
	defer st.Close()

	exists, err := st.HasSuperAdmin()
	if err != nil {
		return fmt.Errorf("check super admin: %w", err)
	}
	if exists {
		log.Printf("super admin already exists; nothing to do")
		return nil
	}

	email := store.NormalizeEmail(cfg.SuperadminEmail)
	if email == "" || cfg.SuperadminPassword == "" {
		return fmt.Errorf("SUPERADMIN_EMAIL and SUPERADMIN_PASSWORD are required")
	}

	hash, err := auth.HashPassword(cfg.SuperadminPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	created, err := st.CreateSuperAdminIfNone(superAdminName, email, hash)
	if err != nil {
		return err
	}
	if created {
		log.Printf("created super admin %s", email)
	}
	return nil
}

func openMigratedStore(cfg config.Config) (*store.Store, error) {
	st, err := store.Open(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("open store: %w", err)
	}
	if err := st.Migrate(); err != nil {
		_ = st.Close()
		return nil, fmt.Errorf("migrate store: %w", err)
	}
	return st, nil
}
