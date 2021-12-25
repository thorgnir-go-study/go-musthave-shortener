package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/cookieauth"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/storage"
	"log"
	"net/http"
)

type responseEntity struct {
	ShortUrl    string `json:"short_url"`
	OriginalUrl string `json:"original_url"`
}

func LoadByUserHandler(s storage.URLStorage, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId, err := ca.GetUserId(r)
		if err != nil {
			if errors.Is(err, cookieauth.ErrNoTokenFound) {
				userId = uuid.NewString()
				ca.SetUserIdCookie(w, userId)
			} else {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		urlEntities, err := s.LoadByUserID(userId)
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
				ShortUrl:    fmt.Sprintf("%s/%s", baseURL, urlEntities[idx].ID),
				OriginalUrl: urlEntities[idx].OriginalURL,
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
