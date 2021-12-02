package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/storage"
	"time"
)

//NewRouter возращает настроенный для сокращения ссылок chi.Router
func NewRouter(storage storage.URLStorage) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(10 * time.Second))

	r.Post("/", ShortenURLHandler(storage))
	r.Get("/{urlID}", ExpandURLHandler(storage))
	r.Post("/api/shorten", JSONShortenURLHandler(storage))

	return r
}
