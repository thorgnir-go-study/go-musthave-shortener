package storage

import "context"

// URLStorager представляет интерфейс работы с хранилищем ссылок
type URLStorager interface {
	// Store сохраняет ссылку в хранилище
	Store(urlEntity URLEntity) error
	// StoreBatch сохраняет список сокращенных ссылок
	StoreBatch(ctx context.Context, entitiesBatch []URLEntity) error
	// Load возвращает сохраненную ссылку по идентификатору. Возвращает сущность ссылки, если она найдена, в противном случае ErrURLNotFound
	Load(key string) (URLEntity, error)
	// LoadByUserID возвращает все ссылки созданные юзером
	LoadByUserID(userID string) ([]URLEntity, error)
	// Ping возвращает статус хранилища
	Ping() error
}
