package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/middlewares/cookieauth"

	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/repository"
	"io"
	"log"
	"net/http"
	"net/url"
)

type jsonShortenRequest struct {
	URL string `json:"url"`
}

type jsonShortenResponse struct {
	Result string `json:"result"`
}

func (s *Service) JSONShortenURLHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyContent, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, "Could not read request body", http.StatusInternalServerError)
			return
		}

		var req jsonShortenRequest

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

		userID, err := cookieauth.FromContext(r.Context())
		if err != nil {
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
				http.Error(w, "Could not write url to repository", http.StatusInternalServerError)
				return
			}
			id = errExists.ID
			status = http.StatusConflict
		}

		resp := &jsonShortenResponse{Result: fmt.Sprintf("%s/%s", s.Config.BaseURL, id)}
		respJSON, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "Can't serialize response", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(status)

		_, err = w.Write(respJSON)
		if err != nil {
			log.Printf("Write failed: %v", err)
		}
	}
}
