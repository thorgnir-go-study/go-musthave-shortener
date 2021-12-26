package storage

import (
	"database/sql"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type dbURLStorage struct {
	DB *sql.DB
}

func NewDBURLStorage(connectionString string) (*dbURLStorage, error) {
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, err
	}
	return &dbURLStorage{DB: db}, nil
}

func (s *dbURLStorage) Store(originalURL string, userID string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (s *dbURLStorage) Load(userID string) (URLEntity, error) {
	//TODO implement me
	panic("implement me")
}

func (s *dbURLStorage) LoadByUserID(userID string) ([]URLEntity, error) {
	//TODO implement me
	panic("implement me")
}

func (s *dbURLStorage) Ping() error {
	return s.DB.Ping()
}
