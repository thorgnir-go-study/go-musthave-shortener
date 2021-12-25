package storage

import (
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

type mapURLStorage struct {
	mx        sync.RWMutex
	m         map[string]URLEntity
	keysRand  rand.Rand
	persister URLStoragePersister
}

type MapURLStorageOption func(*mapURLStorage) error

// CreateMapURLStorage создает реализацию хранилища ссылок в памяти, на основе map
func CreateMapURLStorage(opts ...MapURLStorageOption) (*mapURLStorage, error) {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	storage := &mapURLStorage{
		m:        make(map[string]URLEntity),
		keysRand: *r,
	}

	for _, opt := range opts {
		err := opt(storage)
		if err != nil {
			return nil, err
		}
	}

	return storage, nil
}

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

// Store сохраненяет ссылку в хранилище, возвращает идентификатор сохраненной ссылки
func (s *mapURLStorage) Store(url string, userID string) (key string, err error) {
	s.mx.Lock()
	defer s.mx.Unlock()
	var id string
	for used := true; used; _, used = s.m[id] {
		id = strconv.FormatUint(s.keysRand.Uint64(), 36)
	}
	entity := URLEntity{
		ID:          id,
		OriginalURL: url,
		UserID:      userID,
	}
	s.m[id] = entity

	if s.persister != nil {
		err := s.persister.Store(entity)
		if err != nil {
			log.Println(err)
			return id, err
		}
	}

	return id, nil
}

// Load возвращает сохраненную ссылку по идентификатору. Возвращает ссылку, если она найдена, в противном случае ErrURLNotFound
func (s *mapURLStorage) Load(key string) (urlEntity URLEntity, err error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	value, ok := s.m[key]
	if !ok {
		return URLEntity{}, ErrURLNotFound
	}
	return value, nil
}

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
