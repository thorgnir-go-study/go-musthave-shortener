package repository

import (
	"context"
	"github.com/thorgnir-go-study/go-musthave-shortener/internal/app/config"
)

//goland:noinspection GoNameStartsWithPackageName
type RepositoryType int

const (
	InMemoryRepository RepositoryType = iota
	DatabaseRepository
)

// URLRepository представляет интерфейс работы с хранилищем ссылок
type URLRepository interface {
	// Store сохраняет ссылку в хранилище
	Store(ctx context.Context, urlEntity URLEntity) error
	// StoreBatch сохраняет список сокращенных ссылок
	StoreBatch(ctx context.Context, entitiesBatch []URLEntity) error
	// Load возвращает сохраненную ссылку по идентификатору. Возвращает сущность ссылки, если она найдена, в противном случае ErrURLNotFound
	Load(ctx context.Context, key string) (URLEntity, error)
	// LoadByUserID возвращает все ссылки созданные юзером
	LoadByUserID(ctx context.Context, userID string) ([]URLEntity, error)
	// DeleteURLs помечает ссылки удаленными
	DeleteURLs(ctx context.Context, userID string, ids []string) error
	// Ping возвращает статус хранилища
	Ping(ctx context.Context) error
}

func NewRepository(ctx context.Context, cfg config.Config) (URLRepository, error) {
	var repo URLRepository
	var err error
	switch getRepositoryType(cfg) {
	case InMemoryRepository:
		var options []InMemoryRepositoryOption
		if cfg.StorageFilePath != "" {
			options = append(options, WithFilePersistance(cfg.StorageFilePath))
		}
		repo, err = NewInMemoryRepository(options...)
		if err != nil {
			return nil, err
		}
	case DatabaseRepository:
		repo, err = NewPostgresURLRepository(ctx, cfg.DatabaseDSN)
		if err != nil {
			return nil, err
		}
	}

	return repo, nil
}

func getRepositoryType(cfg config.Config) RepositoryType {
	if cfg.DatabaseDSN != "" {
		return DatabaseRepository
	}
	return InMemoryRepository
}
