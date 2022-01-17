package handlers

import (
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/config"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/repository"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/shortener"
)

type Service struct {
	Repository  repository.URLRepository
	IDGenerator shortener.URLIDGenerator
	Config      config.Config
}

func NewService(repository repository.URLRepository, IDGenerator shortener.URLIDGenerator, config config.Config) *Service {
	return &Service{Repository: repository, IDGenerator: IDGenerator, Config: config}
}
