package storage

import (
	"context"
	"log"
	"sync"
)

type mapURLStorage struct {
	mx        sync.RWMutex
	m         map[string]URLEntity
	persister URLStoragePersister
}

type MapURLStorageOption func(*mapURLStorage) error

// NewMapURLStorage создает реализацию хранилища ссылок в памяти, на основе map
func NewMapURLStorage(opts ...MapURLStorageOption) (*mapURLStorage, error) {
	storage := &mapURLStorage{
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
func WithFilePersistance(filename string) MapURLStorageOption {
	return func(storage *mapURLStorage) error {
		persister := createNewPlainTextFileURLStoragePersister(filename)
		storage.persister = persister
		err := storage.persister.Load(storage.m)
		if err != nil {
			return err
		}
		return nil
	}
}

// Store implements URLStorager.Store
func (s *mapURLStorage) Store(urlEntity URLEntity) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.m[urlEntity.ID] = urlEntity

	if s.persister != nil {
		if err := s.persister.Store(urlEntity); err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func (s *mapURLStorage) StoreBatch(ctx context.Context, entitiesBatch []URLEntity) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	for _, urlEntity := range entitiesBatch {
		s.m[urlEntity.ID] = urlEntity
	}
	return nil
}

// Load implements URLStorager.Load
func (s *mapURLStorage) Load(key string) (urlEntity URLEntity, err error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	value, ok := s.m[key]
	if !ok {
		return URLEntity{}, ErrURLNotFound
	}
	return value, nil
}

// LoadByUserID implements URLStorager.LoadByUserID
func (s *mapURLStorage) LoadByUserID(userID string) ([]URLEntity, error) {
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

// Ping implements URLStorager.Ping
func (s *mapURLStorage) Ping() error {
	return nil
}
