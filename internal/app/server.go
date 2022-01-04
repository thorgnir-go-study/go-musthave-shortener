package app

import (
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/config"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/handlers"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/shortener"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/storage"
	"log"
	"net/http"
)

//StartURLShortenerServer старт нового сервера сокращения ссылок
func StartURLShortenerServer(cfg config.Config, repo storage.URLStorager, idGenerator shortener.URLIDGenerator) {
	service := handlers.NewService(repo, idGenerator, cfg)
	r := handlers.NewRouter(service)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, r))
}
