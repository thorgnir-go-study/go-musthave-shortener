package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/storage"
)

func NewRouter(storage storage.URLStorage) chi.Router {
	r := chi.NewRouter()

	r.Post("/", ShortenURLHandler(storage))
	r.Get("/{urlID}", ExpandURLHandler(storage))

	return r
}
