package handlers

import (
	"fmt"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/storage"
	"io"
	"log"
	"net/http"
	"net/url"
)

//ShortenURLHandler обрабатывает запросы на развертывание сокращенных ссылок
func ShortenURLHandler(s storage.URLStorage, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyContent, err := io.ReadAll(r.Body)

		defer func() {
			err := r.Body.Close()
			if err != nil {
				log.Fatal(err)
			}
		}()

		if err != nil {
			http.Error(w, "Could not read request body", http.StatusInternalServerError)
			return
		}
		u, err := url.ParseRequestURI(string(bodyContent))
		if err != nil {
			http.Error(w, "Not a valid url", http.StatusBadRequest)
			return
		}

		if !u.IsAbs() {
			http.Error(w, "Only absolute urls allowed", http.StatusBadRequest)
			return
		}

		key, err := s.Store(u.String())
		if err != nil {
			http.Error(w, "Could not write url to storage", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(fmt.Sprintf("%s/%s", baseURL, key)))
		if err != nil {
			log.Printf("Write failed: %v", err)
		}
	}
}
