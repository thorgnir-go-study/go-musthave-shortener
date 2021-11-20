package app

import (
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/handlers"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/storage"
	"net/http"
	"time"
)

func StartURLShortenerServer(port uint16, storage storage.URLStorage) {
	r := handlers.NewRouter(storage)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(10 * time.Second))
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), r)
	if err != nil {
		panic(err)
	}
}
