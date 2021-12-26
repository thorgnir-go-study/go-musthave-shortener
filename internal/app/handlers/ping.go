package handlers

import (
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/storage"
	"net/http"
)

// PingHandler проверяет статус хранилища ссылок
func PingHandler(s storage.URLStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := s.Ping()
		if err != nil {
			http.Error(w, "url storage is not accessible", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
