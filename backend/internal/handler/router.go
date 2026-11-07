package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// New returns the root HTTP router. It wires up middleware and mounts all
// feature routes. Feature routes are added incrementally as they are built.
func New() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Feature routes are mounted here as they are implemented:
	// r.Route("/api", func(r chi.Router) {
	//     r.Mount("/auth", auth.Routes(...))
	//     r.Mount("/classes", classes.Routes(...))
	//     r.Mount("/sessions", sessions.Routes(...))
	// })

	return r
}
