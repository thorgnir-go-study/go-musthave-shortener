package handlers

import (
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/config"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/shortener"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/storage"
)

type Service struct {
	Repository  storage.URLStorager
	IDGenerator shortener.URLIDGenerator
	Config      config.Config
}

func NewService(repository storage.URLStorager, IDGenerator shortener.URLIDGenerator, config config.Config) *Service {
	return &Service{Repository: repository, IDGenerator: IDGenerator, Config: config}
}
