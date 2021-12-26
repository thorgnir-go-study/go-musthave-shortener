package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/config"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/cookieauth"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/middlewares"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/storage"
	"time"
)

var ca *cookieauth.CookieAuth

//NewRouter возращает настроенный для сокращения ссылок chi.Router
func NewRouter(storage storage.URLStorage, cfg config.Config) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(10 * time.Second))
	r.Use(middleware.Compress(5))
	r.Use(middlewares.GzipRequestDecompressor)

	ca = cookieauth.New([]byte("secret"), "UserId")

	r.Post("/", ShortenURLHandler(storage, cfg.BaseURL))
	r.Get("/{urlID}", ExpandURLHandler(storage))
	r.Post("/api/shorten", JSONShortenURLHandler(storage, cfg.BaseURL))
	r.Get("/user/urls", LoadByUserHandler(storage, cfg.BaseURL))
	r.Get("/ping", PingHandler(storage))

	return r
}
