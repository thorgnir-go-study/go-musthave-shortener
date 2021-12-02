package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/storage"
	"io"
	"log"
	"net/http"
	"net/url"
)

type request struct {
	URL string `json:"url"`
}

type response struct {
	Result string `json:"result"`
}

func JSONShortenURLHandler(s storage.URLStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyContent, err := io.ReadAll(r.Body)

		defer r.Body.Close()

		if err != nil {
			http.Error(w, "Could not read request body", http.StatusInternalServerError)
			return
		}

		var req request

		if err := json.Unmarshal(bodyContent, &req); err != nil {
			http.Error(w, "Invalid json", http.StatusBadRequest)
		}

		u, err := url.ParseRequestURI(req.URL)
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

		responseObj := &response{Result: fmt.Sprintf("http://localhost:8080/%s", key)}
		serializedResp, err := json.Marshal(responseObj)
		if err != nil {
			http.Error(w, "Can't serialize response", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusCreated)

		_, err = w.Write(serializedResp)
		if err != nil {
			log.Printf("Write failed: %v", err)
		}
	}
}
