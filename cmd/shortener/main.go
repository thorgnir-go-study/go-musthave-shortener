package main

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/config"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/repository"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/shortener"
	"os"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to get configuration")
	}
	configureLogger(*cfg)

	urlStorage, err := repository.NewRepository(context.Background(), *cfg)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to create repository")
	}

	idGenerator := shortener.NewRandomStringURLIDGenerator(cfg.ShortURLIdentifierLength)

	app.StartURLShortenerServer(*cfg, urlStorage, idGenerator)
}

func configureLogger(_ config.Config) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// в дальнейшем можно добавить в конфиг требуемый уровень логирования, аутпут (файл или еще чего) и т.д.
	// пока пишем в консоль красивенько
	log.Logger = log.With().Caller().Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})
}
