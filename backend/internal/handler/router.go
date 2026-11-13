package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/mhdfadlyhasan/study-companion/internal/auth"
	"github.com/mhdfadlyhasan/study-companion/internal/store"
)

type handler struct {
	store     *store.Store
	jwtSecret string
}

// New returns the root HTTP router. It wires up middleware and mounts all
// feature routes. Feature routes are added incrementally as they are built.
func New(st *store.Store, jwtSecret string) http.Handler {
	h := &handler{store: st, jwtSecret: jwtSecret}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/api", func(r chi.Router) {
		r.Post("/auth/login", h.login)

		r.Group(func(r chi.Router) {
			r.Use(auth.Authenticate(jwtSecret))
			r.Get("/me", h.me)
		})
	})

	return r
}
