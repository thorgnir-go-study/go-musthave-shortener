package repository

import (
	"errors"
	"fmt"
)

// ErrURLNotFound ошибка "ссылка не найдена в хранилище"
var ErrURLNotFound = errors.New("url not found in repository")

// ErrURLExists ошибка "оригинальная ссылка уже существует в хранилище"
type ErrURLExists struct {
	ID  string
	Err error
}

func NewErrURLExists(id string) *ErrURLExists {
	return &ErrURLExists{ID: id}
}

func (e *ErrURLExists) Error() string {
	return fmt.Sprintf("url already exists in repository. short url id: %s", e.ID)
}

func (e *ErrURLExists) Unwrap() error {
	return e.Err
}
