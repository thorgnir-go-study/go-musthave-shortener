package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4/stdlib"
	"math/rand"
	"strconv"
	"time"
)

type dbURLStorage struct {
	DB       *sql.DB
	keysRand rand.Rand
}

func NewDBURLStorage(connectionString string) (*dbURLStorage, error) {
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, err
	}
	err = createTables(db)
	if err != nil {
		return nil, err
	}
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	return &dbURLStorage{DB: db, keysRand: *r}, nil
}

//goland:noinspection SqlNoDataSourceInspection,SqlResolve
func (s *dbURLStorage) Store(originalURL string, userID string) (string, error) {
	var newUrlId string
	query := `insert into urls(url_id, original_url, user_id) values($1, $2, $3)`
	// Тут может быть много разных тактик генерации уникального идентификатора ссылки.
	// Можно генерировать извне стораджа, и при нарушении уникальности возвращать из стораджа ошибку, пусть клиент генерирует новый и пытается сохранить еще раз
	// Можно перед запросом на инсерт проверять, нет ли такого уже в БД (но тут заморочки с транзакциями и вообще дорого, считаем что ситуация слишком редкая, чтоб такое делать)
	// Здесь я пошел по варианту, что айди для ссылки генерируется на стороне стораджа (как и в ин-мемори), и при нарушении констрейнта - генерируем новый айди и пытаемся сохранить снова
	for retry := true; retry; {
		retry = false
		newUrlId = strconv.FormatUint(s.keysRand.Uint64(), 36)
		_, err := s.DB.Exec(query, newUrlId, originalURL, userID)
		if err != nil {
			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) {
				return "", err
			}
			fmt.Println(err)
			if pgErr.ConstraintName != "url_id_unique" {
				return "", err
			}
			retry = true
		}
	}
	return newUrlId, nil
}

//goland:noinspection SqlNoDataSourceInspection,SqlResolve
func (s *dbURLStorage) Load(key string) (URLEntity, error) {
	query := `select url_id, original_url, user_id  from urls where url_id = $1`
	var entity URLEntity
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	result := s.DB.QueryRowContext(ctx, query, key)
	err := result.Scan(&entity.ID, &entity.OriginalURL, &entity.UserID)
	if err != nil {
		fmt.Println(err)
		if errors.Is(err, sql.ErrNoRows) {
			return URLEntity{}, ErrURLNotFound
		}
		return URLEntity{}, err
	}
	return entity, nil
}

//goland:noinspection SqlNoDataSourceInspection,SqlResolve
func (s *dbURLStorage) LoadByUserID(userID string) ([]URLEntity, error) {
	query := `select url_id, original_url, user_id  from urls where user_id=$1`
	var result []URLEntity
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	rows, err := s.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var entity URLEntity
		err = rows.Scan(&entity.ID, &entity.OriginalURL, &entity.UserID)
		if err != nil {
			return nil, err
		}
		result = append(result, entity)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *dbURLStorage) Ping() error {
	return s.DB.Ping()
}

func createTables(db *sql.DB) error {
	//goland:noinspection SqlNoDataSourceInspection
	createScript := `
	CREATE TABLE IF NOT EXISTS urls
	(
		id bigint NOT NULL GENERATED ALWAYS AS IDENTITY,
		url_id character varying NOT NULL,
		original_url character varying  NOT NULL,
		user_id character varying,
		CONSTRAINT urls_pkey PRIMARY KEY (id),
		CONSTRAINT url_id_unique UNIQUE (url_id)
	)
	`
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	_, err := db.ExecContext(ctx, createScript)
	return err
}
