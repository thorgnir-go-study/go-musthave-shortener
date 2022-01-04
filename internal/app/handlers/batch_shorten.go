package handlers

import (
	"context"
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
	"time"
)

type batchShortenRequestEntity struct {
	OriginalURL   string `json:"original_url"`
	CorrelationID string `json:"correlation_id"`
}
type batchShortenResponseEntity struct {
	CorrelationID string `json:"correlation_id"`
	ShortUrl      string `json:"short_url"`
}

type batchShortenRequest []batchShortenRequestEntity

func (s *Service) BatchShortenURLHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		bodyContent, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, "Could not read request body", http.StatusInternalServerError)
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

		var req batchShortenRequest

		if err := json.Unmarshal(bodyContent, &req); err != nil {
			http.Error(w, "Invalid json", http.StatusBadRequest)
		}

		if isValid, invalidUrl := isValidRequest(req); !isValid {
			http.Error(w, "invalid url"+invalidUrl, http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), time.Minute)
		defer cancel()
		batch := storage.NewBatchService(100, s.Repository)
		//goland:noinspection GoUnhandledErrorResult
		defer batch.Flush(ctx)
		resp := make([]batchShortenResponseEntity, len(req))

		for _, reqEntity := range req {
			id := s.IDGenerator.GenerateURLID(reqEntity.OriginalURL)
			entity := storage.URLEntity{
				ID:          id,
				OriginalURL: reqEntity.OriginalURL,
				UserID:      userID,
			}
			err = batch.Add(r.Context(), entity)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			resp = append(resp, batchShortenResponseEntity{
				CorrelationID: reqEntity.CorrelationID,
				ShortUrl:      fmt.Sprintf("%s/%s", s.Config.BaseURL, id),
			})
		}

		serializedResp, err := json.Marshal(resp)
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

func isValidRequest(req batchShortenRequest) (isValid bool, firstInvalidUrl string) {
	for _, entity := range req {
		if !isValidUrl(entity.OriginalURL) {
			return false, entity.OriginalURL
		}
	}

	return true, ""
}

func isValidUrl(input string) bool {
	u, err := url.ParseRequestURI(input)
	if err != nil {
		return false
	}

	return u.IsAbs()
}
