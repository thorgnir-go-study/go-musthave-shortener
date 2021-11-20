package handlers

import (
	"fmt"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/storage"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func RootHanlder(storage storage.URLStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			shortenURLHandler(storage)(w, r)
		case http.MethodGet:
			getURLHandler(storage)(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}

	}
}

func shortenURLHandler(storage storage.URLStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyContent, err := io.ReadAll(r.Body)
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

		key, err := storage.Store(u.String())
		if err != nil {
			http.Error(w, "Could not write url to storage", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(fmt.Sprintf("http://localhost:8080/%s", key)))
		if err != nil {
			log.Printf("Write failed: %v", err)
		}
	}
}

func getURLHandler(storage storage.URLStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.Trim(r.URL.Path, "/")

		if len(path) == 0 {
			http.Error(w, "URL id missing", http.StatusBadRequest)
			return
		}

		pathSegments := strings.Split(path, "/")
		if len(pathSegments) > 1 {
			http.NotFound(w, r)
			return
		}

		urlID := pathSegments[0]

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
