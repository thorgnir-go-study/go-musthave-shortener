package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/middlewares/cookieauth"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/repository"
	"io"
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
	ShortURL      string `json:"short_url"`
}

type batchShortenRequest []batchShortenRequestEntity

func (s *Service) BatchShortenURLHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyContent, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			log.Error().Err(err).Msg("could not read request body")
			http.Error(w, "Could not read request body", http.StatusInternalServerError)
			return
		}
		userID, err := cookieauth.FromContext(r.Context())
		if err != nil {
			log.Info().Err(err).Msg("could not read request body")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req batchShortenRequest

		if err = json.Unmarshal(bodyContent, &req); err != nil {
			log.Info().Err(err).Msg("invalid json")
			http.Error(w, "Invalid json", http.StatusBadRequest)
		}

		if isValid, invalidURL := isValidRequest(req); !isValid {
			log.Info().Err(err).Msg("invalid url" + invalidURL)
			http.Error(w, "invalid url"+invalidURL, http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), time.Minute)
		defer cancel()
		batch := repository.NewBatchURLEntityStoreService(s.Config.ShortenBatchSize, s.Repository)

		resp := make([]batchShortenResponseEntity, len(req))

		for idx, reqEntity := range req {
			id := s.IDGenerator.GenerateURLID(reqEntity.OriginalURL)
			entity := repository.URLEntity{
				ID:          id,
				OriginalURL: reqEntity.OriginalURL,
				UserID:      userID,
			}
			err = batch.Add(ctx, entity)
			if err != nil {
				log.Error().Err(err).Msg("error in batch.add")
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			resp[idx] = batchShortenResponseEntity{
				CorrelationID: reqEntity.CorrelationID,
				ShortURL:      fmt.Sprintf("%s/%s", s.Config.BaseURL, id),
			}
		}
		// по комменту из ревью (сделай через defer func() { err := batch.Flush(ctx)}, так у тебя добавиться больше опций и если где-то ты добавишь return, то у тебя Flush все равно сработает)
		// flush-то сработает, но ошибку мы уже не поймаем, и на клиент не отдадим 500 (попробовал, тестом поймал что в таком случае при отказе репозитория - клиенту отдается 201 типа все в порядке)
		// так что оставляю так
		err = batch.Flush(ctx)
		if err != nil {
			log.Error().Err(err).Msg("error while batch.flush")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		serializedResp, err := json.Marshal(resp)
		if err != nil {
			log.Error().Err(err).Msg("can't serialize response")
			http.Error(w, "Can't serialize response", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusCreated)

		_, err = w.Write(serializedResp)
		if err != nil {
			log.Error().Err(err).Msg("write response failed")
			log.Printf("Write failed: %v", err)
		}
	}
}

func isValidRequest(req batchShortenRequest) (isValid bool, firstInvalidURL string) {
	for _, entity := range req {
		if !isValidURL(entity.OriginalURL) {
			return false, entity.OriginalURL
		}
	}

	return true, ""
}

func isValidURL(input string) bool {
	u, err := url.ParseRequestURI(input)
	if err != nil {
		return false
	}

	return u.IsAbs()
}
