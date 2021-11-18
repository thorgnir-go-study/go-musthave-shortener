package app

import (
	"math/rand"
	"strconv"
	"sync"
	"time"
)

type mapURLStorage struct {
	mx       sync.RWMutex
	m        map[string]string
	keysRand *rand.Rand
}

func CreateMapURLStorage() *mapURLStorage {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	return &mapURLStorage{
		m:        make(map[string]string),
		keysRand: r,
	}
}

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

func (s *mapURLStorage) Load(key string) (url string, found bool, err error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	value, ok := s.m[key]
	return value, ok, nil
}
