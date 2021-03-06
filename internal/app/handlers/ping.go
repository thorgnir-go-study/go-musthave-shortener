package handlers

import (
	"github.com/rs/zerolog/log"
	"net/http"
)

// PingHandler проверяет статус хранилища ссылок
func (s *Service) PingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := s.Repository.Ping(r.Context())
		if err != nil {
			log.Error().Err(err).Msg("url repository is not accessible")
			http.Error(w, "url repository is not accessible", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
