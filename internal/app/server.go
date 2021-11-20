package app

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/handlers"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/storage"
	"net/http"
)

func StartURLShortenerServer(port uint16, storage storage.URLStorage) {
	r := chi.NewRouter()

	r.Post("/", handlers.ShortenURLHandler(storage))
	r.Get("/{urlID}", handlers.ExpandURLHandler(storage))

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), r)
	if err != nil {
		panic(err)
	}
}
