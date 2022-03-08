package handlers

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/middlewares/cookieauth"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/repository"
	"io"
	"net/http"
	"net/url"
)

//ShortenURLHandler обрабатывает запросы на развертывание сокращенных ссылок
func (s *Service) ShortenURLHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyContent, err := io.ReadAll(r.Body)
		defer r.Body.Close()

		if err != nil {
			log.Info().Err(err).Msg("could not read request body")
			http.Error(w, "Could not read request body", http.StatusInternalServerError)
			return
		}
		u, err := url.ParseRequestURI(string(bodyContent))
		if err != nil {
			log.Info().Err(err).Msg("not a valid url")
			http.Error(w, "Not a valid url", http.StatusBadRequest)
			return
		}

		if !u.IsAbs() {
			log.Info().Err(err).Msg("not an absolute url")
			http.Error(w, "Only absolute urls allowed", http.StatusBadRequest)
			return
		}

		userID, err := cookieauth.FromContext(r.Context())
		if err != nil {
			log.Info().Err(err).Msg("unauthorized")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		id := s.IDGenerator.GenerateURLID(u.String())
		urlEntity := repository.URLEntity{
			ID:          id,
			OriginalURL: u.String(),
			UserID:      userID,
		}
		status := http.StatusCreated
		err = s.Repository.Store(r.Context(), urlEntity)
		if err != nil {
			var errExists *repository.ErrURLExists
			if !errors.As(err, &errExists) {
				log.Error().Err(err).Msg("could not write url to repository")
				http.Error(w, "Could not write url to repository", http.StatusInternalServerError)
				return
			}
			id = errExists.ID
			status = http.StatusConflict
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(status)
		_, err = w.Write([]byte(fmt.Sprintf("%s/%s", s.Config.BaseURL, id)))
		if err != nil {
			log.Error().Err(err).Msg("could write response")
		}
	}
}
