package storage

import (
	"math/rand"
	"strconv"
	"sync"
	"time"
)

type mapURLStorage struct {
	mx       sync.RWMutex
	m        map[string]string
	keysRand rand.Rand
}

// CreateMapURLStorage создает реализацию хранилища ссылок в памяти, на основе map
func CreateMapURLStorage() *mapURLStorage {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	return &mapURLStorage{
		m:        make(map[string]string),
		keysRand: *r,
	}
}

// Store сохраненяет ссылку в хранилище, возвращает идентификатор сохраненной ссылки
func (s *mapURLStorage) Store(url string) (key string, err error) {
	s.mx.Lock()
	defer s.mx.Unlock()
	var id string
	for used := true; used; _, used = s.m[id] {
		id = strconv.FormatUint(s.keysRand.Uint64(), 36)
	}
	s.m[id] = url

	return id, nil
}

// Load возвращает сохраненную ссылку по идентификатору. Возвращает ссылку, если она найдена, в противном случае URLNotFoundErr
func (s *mapURLStorage) Load(key string) (url string, err error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	value, ok := s.m[key]
	if !ok {
		return "", URLNotFoundErr
	}
	return value, nil
}
