package handlers

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/repository"
	"net/http"
)

// ExpandURLHandler обрабатывает запросы на сокращение ссылок
func (s *Service) ExpandURLHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urlID := chi.URLParam(r, "urlID")

		u, err := s.Repository.Load(r.Context(), urlID)
		if errors.Is(err, repository.ErrURLNotFound) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			log.Error().Err(err).Msg("error while loading shortened link")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if u.Deleted {
			w.WriteHeader(http.StatusGone)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Location", u.OriginalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)

	}
}
