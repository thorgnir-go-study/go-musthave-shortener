package handlers

import (
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/storage"
	"net/http"
	"strings"
)

func ExpandURLHandler(storage storage.URLStorage) http.HandlerFunc {
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
