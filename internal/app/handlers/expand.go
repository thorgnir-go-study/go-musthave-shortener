package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/storage"
	"net/http"
)

func ExpandURLHandler(storage storage.URLStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urlID := chi.URLParam(r, "urlID")
		u, found, err := storage.Load(urlID)
		if err != nil {
			http.Error(w, "Could not read from url storage", http.StatusInternalServerError)
			return
		}
		if found {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Header().Set("Location", u)
			w.WriteHeader(http.StatusTemporaryRedirect)
		} else {
			http.NotFound(w, r)
		}
	}
}
