package main

import (
	"log"
	"net/http"

	"github.com/mhdfadlyhasan/study-companion/internal/config"
	"github.com/mhdfadlyhasan/study-companion/internal/handler"
	"github.com/mhdfadlyhasan/study-companion/internal/store"
)

func main() {
	cfg := config.Load()

	st, err := store.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer st.Close()

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: handler.New(),
	}

	log.Printf("study-companion listening on %s (db=%s)", cfg.Addr, cfg.DBPath)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
}
