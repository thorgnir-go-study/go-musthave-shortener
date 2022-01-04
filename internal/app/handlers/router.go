package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/cookieauth"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/middlewares"
	"time"
)

var ca *cookieauth.CookieAuth

//NewRouter возращает настроенный для сокращения ссылок chi.Router
func NewRouter(service *Service) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(10 * time.Second))
	r.Use(middleware.Compress(5))
	r.Use(middlewares.GzipRequestDecompressor)

	ca = cookieauth.New([]byte(service.Config.AuthSecretKey), "UserId")

	r.Post("/", service.ShortenURLHandler())
	r.Get("/{urlID}", service.ExpandURLHandler())
	r.Post("/api/shorten", service.JSONShortenURLHandler())
	r.Get("/user/urls", service.LoadByUserHandler())
	r.Get("/ping", service.PingHandler())

	return r
}
