package main

import (
	"github.com/caarlos0/env/v6"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/config"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/storage"
	"log"
)

func main() {
	var cfg config.Config

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalln("Error parsing config")
	}

	urlStorage := storage.CreateMapURLStorage()
	app.StartURLShortenerServer(cfg, urlStorage)
}
