package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/middlewares/cookieauth"
	"log"
	"net/http"
)

type responseEntity struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func (s *Service) LoadByUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := cookieauth.FromContext(r.Context())
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		urlEntities, err := s.Repository.LoadByUserID(r.Context(), userID)
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
