package app

import (
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/config"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/handlers"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/storage"
	"log"
	"net/http"
)

//StartURLShortenerServer старт нового сервера сокращения ссылок
func StartURLShortenerServer(cfg config.Config, storage storage.URLStorage) {

	r := handlers.NewRouter(storage, cfg)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, r))
}
