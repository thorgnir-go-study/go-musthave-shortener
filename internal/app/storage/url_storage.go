package storage

// URLStorage представляет интерфейс работы с хранилищем ссылок
type URLStorage interface {
	// Store сохраненяет ссылку в хранилище, возвращает идентификатор сохраненной ссылки
	Store(originalURL string, userID string) (string, error)
	// Load возвращает сохраненную ссылку по идентификатору. Возвращает ссылку, если она найдена, в противном случае ErrURLNotFound
	Load(userID string) (URLEntity, error)
	// LoadByUserID возвращает все ссылки созданные юзером
	LoadByUserID(userID string) ([]URLEntity, error)
}
