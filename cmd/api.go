package main

import (
	"fmt"
	"github.com/AdamElHassanLeb/VOD-Downloader/API/cmd/internal/Controllers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"log"
	"net/http"
	"time"
)

type server struct {
	config config
}

type config struct {
	port int
}

func (s *server) mount() http.Handler {

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS middleware settings
	r.Use(cors.Handler(cors.Options{
		//AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Route("/api", func(r chi.Router) {
		r.Route("/Media", func(r chi.Router) {
			r.Route("/Ingest", func(r chi.Router) {
				r.Get("/VOD", Controllers.IngestVOD)
			})
		})
	})
	return r
}

func (s *server) run(mux http.Handler) error {

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.port),
		Handler:      mux,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Starting server on port %d\n", s.config.port)

	return server.ListenAndServe()
}
