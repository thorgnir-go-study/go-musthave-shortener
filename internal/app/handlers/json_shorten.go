package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/cookieauth"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/storage"
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

		userID, err := ca.GetUserID(r)
		if err != nil {
			if errors.Is(err, cookieauth.ErrNoTokenFound) || errors.Is(err, cookieauth.ErrInvalidToken) {
				userID = uuid.NewString()
				ca.SetUserIDCookie(w, userID)
			} else {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}
		id := s.IDGenerator.GenerateURLID(u.String())
		urlEntity := storage.URLEntity{
			ID:          id,
			OriginalURL: u.String(),
			UserID:      userID,
		}
		status := http.StatusCreated
		err = s.Repository.Store(urlEntity)
		if err != nil {
			var errExists *storage.ErrURLExists
			if !errors.As(err, &errExists) {
				http.Error(w, "Could not write url to storage", http.StatusInternalServerError)
				return
			}
			id = errExists.ID
			status = http.StatusConflict
		}

		responseObj := &jsonShortenResponse{Result: fmt.Sprintf("%s/%s", s.Config.BaseURL, id)}
		serializedResp, err := json.Marshal(responseObj)
		if err != nil {
			http.Error(w, "Can't serialize response", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(status)

		_, err = w.Write(serializedResp)
		if err != nil {
			log.Printf("Write failed: %v", err)
		}
	}
}
