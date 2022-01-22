package handlers

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/middlewares/cookieauth"
	"io"
	"net/http"
)

// DeleteURLsHandler обрабатывает запросы на удаление ссылок
func (s *Service) DeleteURLsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyContent, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			log.Error().Err(err).Str("handler", "DeleteURLsHandler").Msg("could not read request body")
			http.Error(w, "Could not read request body", http.StatusInternalServerError)
			return
		}
		var ids []string
		if err = json.Unmarshal(bodyContent, &ids); err != nil {
			log.Info().Err(err).Str("handler", "DeleteURLsHandler").Msg("invalid json")
			http.Error(w, "Invalid json", http.StatusBadRequest)
		}

		userID, err := cookieauth.FromContext(r.Context())
		if err != nil {
			log.Info().Err(err).Str("handler", "DeleteURLsHandler").Msg("unauthorized")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		s.deleteURLsReqCh <- deleteURLsRequest{
			ids:    ids,
			userID: userID,
		}

		w.WriteHeader(http.StatusAccepted)

	}
}
