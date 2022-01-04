package handlers

import (
	"net/http"
)

// PingHandler проверяет статус хранилища ссылок
func (s *Service) PingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := s.Repository.Ping(r.Context())
		if err != nil {
			http.Error(w, "url storage is not accessible", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
