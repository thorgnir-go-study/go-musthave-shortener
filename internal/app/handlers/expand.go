package handlers

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/storage"
	"net/http"
)

// ExpandURLHandler обрабатывает запросы на сокращение ссылок
func (s *Service) ExpandURLHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urlID := chi.URLParam(r, "urlID")

		u, err := s.Repository.Load(r.Context(), urlID)
		if errors.Is(err, storage.ErrURLNotFound) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Location", u.OriginalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)

	}
}
