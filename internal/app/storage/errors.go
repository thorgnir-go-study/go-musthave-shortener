package storage

import "errors"

// ErrURLNotFound ошибка "ссылка не найдена в хранилище"
var ErrURLNotFound = errors.New("url not found in storage")
