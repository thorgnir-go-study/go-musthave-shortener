package handlers

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/storage"
	"net/http"
)

// ExpandURLHandler обрабатывает запросы на сокращение ссылок
func ExpandURLHandler(s storage.URLStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urlID := chi.URLParam(r, "urlID")

		u, err := s.Load(urlID)
		if errors.Is(err, storage.URLNotFoundErr) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Location", u)
		w.WriteHeader(http.StatusTemporaryRedirect)

	}
}
