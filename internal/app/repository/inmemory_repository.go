package repository

import (
	"context"
	"github.com/rs/zerolog/log"
	"sync"
)

type inMemoryRepo struct {
	mx        sync.RWMutex
	m         map[string]URLEntity
	persister inMemoryRepoFilePersister
}

type InMemoryRepositoryOption func(*inMemoryRepo) error

// NewInMemoryRepository создает реализацию хранилища ссылок в памяти, на основе map
func NewInMemoryRepository(opts ...InMemoryRepositoryOption) (*inMemoryRepo, error) {
	storage := &inMemoryRepo{
		m: make(map[string]URLEntity),
	}

	for _, opt := range opts {
		err := opt(storage)
		if err != nil {
			return nil, err
		}
	}

	return storage, nil
}

// WithFilePersistance позволяет сохранять в файле состояние хранилища, и при создании хранилища восстанавливать состояние из файла.
func WithFilePersistance(filename string) InMemoryRepositoryOption {
	return func(storage *inMemoryRepo) error {
		persister := createNewInMemoryRepoFilePersisterPlain(filename)
		storage.persister = persister
		err := storage.persister.Load(storage.m)
		if err != nil {
			return err
		}
		return nil
	}
}

// Store implements URLRepository.Store
func (s *inMemoryRepo) Store(_ context.Context, urlEntity URLEntity) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	// По заданию было добавить уникальный индекс по оригинальной ссылке только в хранилище БД
	// Поэтому тут проверка уникальности нереализована.
	// Можно реализовать, но будет крайне неэффективно при данной модели хранения - придется перебирать все записи
	s.m[urlEntity.ID] = urlEntity

	// по поводу "задачи со звездочкой" (писать в файл через middleware)
	// я не очень понял, как это можно тут реализовать малой кровью.
	// обычные http middleware тут не подходят очевидно (они на другом уровне вообще)
	// если делать нечто похожее на уровне хранилища - это потребует серьезной переделки интерфейса хранилища
	// (добавление поддержки построения цепочек вызовов, причем либо только для метода Store, либо придумывать какой-то обобщенный интерфейс для всех методов (что уже звучит сомнительно)
	// В общем, выглядит как очень сомнительная доработка, требующая внушительных усилий и ухудшения интерфейсов (но если я не догадался до какого-то очевидного решения - буду рад услышать)
	if s.persister != nil {
		if err := s.persister.Store(urlEntity); err != nil {
			log.Error().Err(err).Msg("error while writing to file")
			return err
		}
	}

	return nil
}

func (s *inMemoryRepo) StoreBatch(_ context.Context, entitiesBatch []URLEntity) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	for _, urlEntity := range entitiesBatch {
		s.m[urlEntity.ID] = urlEntity
	}
	return nil
}

// Load implements URLRepository.Load
func (s *inMemoryRepo) Load(_ context.Context, key string) (urlEntity URLEntity, err error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	value, ok := s.m[key]
	if !ok {
		return URLEntity{}, ErrURLNotFound
	}
	return value, nil
}

// DeleteURLs implements URLRepository.DeleteURLs
func (s *inMemoryRepo) DeleteURLs(_ context.Context, userID string, ids []string) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	for _, id := range ids {
		if entity, ok := s.m[id]; ok && entity.UserID == userID {
			entity.Deleted = true
			s.m[id] = entity

			if s.persister != nil {
				if err := s.persister.Store(entity); err != nil {
					log.Error().Err(err).Msg("error while writing to file")
					return err
				}
			}
		}
	}
	return nil
}

// LoadByUserID implements URLRepository.LoadByUserID
func (s *inMemoryRepo) LoadByUserID(_ context.Context, userID string) ([]URLEntity, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	entities := make([]URLEntity, 0)
	for _, entity := range s.m {
		if entity.UserID == userID {
			entities = append(entities, entity)
		}
	}
	return entities, nil
}

// Ping implements URLRepository.Ping
func (s *inMemoryRepo) Ping(_ context.Context) error {
	return nil
}
