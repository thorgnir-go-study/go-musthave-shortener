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
func (s *mapURLStorage) Store(_ context.Context, urlEntity URLEntity) error {
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
			log.Println(err)
			return err
		}
	}

	return nil
}

func (s *mapURLStorage) StoreBatch(_ context.Context, entitiesBatch []URLEntity) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	for _, urlEntity := range entitiesBatch {
		s.m[urlEntity.ID] = urlEntity
	}
	return nil
}

// Load implements URLStorager.Load
func (s *mapURLStorage) Load(_ context.Context, key string) (urlEntity URLEntity, err error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	value, ok := s.m[key]
	if !ok {
		return URLEntity{}, ErrURLNotFound
	}
	return value, nil
}

// LoadByUserID implements URLStorager.LoadByUserID
func (s *mapURLStorage) LoadByUserID(_ context.Context, userID string) ([]URLEntity, error) {
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
func (s *mapURLStorage) Ping(_ context.Context) error {
	return nil
}
