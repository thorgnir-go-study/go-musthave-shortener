package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/cookieauth"
	"log"
	"net/http"
)

type responseEntity struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func (s *Service) LoadByUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		urlEntities, err := s.Repository.LoadByUserID(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(urlEntities) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		respEntities := make([]responseEntity, len(urlEntities))
		for idx := range urlEntities {
			respEntities[idx] = responseEntity{
				ShortURL:    fmt.Sprintf("%s/%s", s.Config.BaseURL, urlEntities[idx].ID),
				OriginalURL: urlEntities[idx].OriginalURL,
			}
		}

		serializedResp, err := json.Marshal(respEntities)
		if err != nil {
			http.Error(w, "Can't serialize response", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, err = w.Write(serializedResp)
		if err != nil {
			log.Printf("Write failed: %v", err)
		}

	}
}
