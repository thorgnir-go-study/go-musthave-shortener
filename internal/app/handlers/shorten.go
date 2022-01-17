package handlers

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/cookieauth"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/repository"
	"io"
	"log"
	"net/http"
	"net/url"
)

//ShortenURLHandler обрабатывает запросы на развертывание сокращенных ссылок
func (s *Service) ShortenURLHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyContent, err := io.ReadAll(r.Body)

		defer func() {
			err := r.Body.Close()
			if err != nil {
				log.Fatal(err)
			}
		}()

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

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(status)
		_, err = w.Write([]byte(fmt.Sprintf("%s/%s", s.Config.BaseURL, id)))
		if err != nil {
			log.Printf("Write failed: %v", err)
		}
	}
}
