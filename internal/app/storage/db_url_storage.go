package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"math/rand"
	"time"
)

type dbURLStorage struct {
	DB       *sqlx.DB
	keysRand rand.Rand
}

var (
	insertStmt         *sqlx.NamedStmt
	getByUrlIdStmt     *sqlx.Stmt
	selectByUserIdStmt *sqlx.Stmt
)

func NewDBURLStorage(connectionString string) (*dbURLStorage, error) {
	db, err := sqlx.Open("pgx", connectionString)
	if err != nil {
		return nil, err
	}

	if err = createTables(db); err != nil {
		return nil, err
	}

	if err = prepareStatements(db); err != nil {
		return nil, err
	}

	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	return &dbURLStorage{DB: db, keysRand: *r}, nil
}

//goland:noinspection SqlNoDataSourceInspection,SqlResolve
func prepareStatements(db *sqlx.DB) error {
	var err error
	if insertStmt, err = db.PrepareNamed(`insert into urls(url_id, original_url, user_id) values(:url_id, :original_url, :user_id)`); err != nil {
		return err
	}

	if getByUrlIdStmt, err = db.Preparex(`select url_id, original_url, user_id  from urls where url_id = $1`); err != nil {
		return err
	}

	if selectByUserIdStmt, err = db.Preparex(`select url_id, original_url, user_id  from urls where user_id=$1`); err != nil {
		return err
	}

	return nil
}

func (s *dbURLStorage) Store(urlEntity URLEntity) error {
	if _, err := insertStmt.Exec(&urlEntity); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

//func (s *dbURLStorage) Store(originalURL string, userID string) (string, error) {
//	urlEntity := URLEntity{
//		OriginalURL: originalURL,
//		UserID:      userID,
//	}
//	// Тут может быть много разных тактик генерации уникального идентификатора ссылки.
//	// Можно генерировать извне стораджа, и при нарушении уникальности возвращать из стораджа ошибку, пусть клиент генерирует новый и пытается сохранить еще раз
//	// Можно перед запросом на инсерт проверять, нет ли такого уже в БД (но тут заморочки с транзакциями и вообще дорого, считаем что ситуация слишком редкая, чтоб такое делать)
//	// Здесь я пошел по варианту, что айди для ссылки генерируется на стороне стораджа (как и в ин-мемори), и при нарушении констрейнта - генерируем новый айди и пытаемся сохранить снова
//	for retry := true; retry; {
//		retry = false
//		urlEntity.ID = strconv.FormatUint(s.keysRand.Uint64(), 36)
//		_, err := insertStmt.Exec(&urlEntity)
//		if err != nil {
//			fmt.Println(err)
//			var pgErr *pgconn.PgError
//			if !errors.As(err, &pgErr) {
//				return "", err
//			}
//			fmt.Println(err)
//			if pgErr.ConstraintName != "url_id_unique" {
//				return "", err
//			}
//			retry = true
//		}
//	}
//	return urlEntity.ID, nil
//}

func (s *dbURLStorage) Load(key string) (URLEntity, error) {
	var entity URLEntity
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err := getByUrlIdStmt.GetContext(ctx, &entity, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return URLEntity{}, ErrURLNotFound
		}
		return URLEntity{}, err
	}
	return entity, nil
}

func (s *dbURLStorage) LoadByUserID(userID string) ([]URLEntity, error) {
	//query := `select url_id, original_url, user_id  from urls where user_id=$1`
	var result []URLEntity
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := selectByUserIdStmt.SelectContext(ctx, &result, userID); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *dbURLStorage) Ping() error {
	return s.DB.Ping()
}

func createTables(db *sqlx.DB) error {
	//goland:noinspection SqlNoDataSourceInspection
	createScript := `
	CREATE TABLE IF NOT EXISTS urls
	(
		id bigint NOT NULL GENERATED ALWAYS AS IDENTITY,
		url_id character varying NOT NULL,
		original_url character varying  NOT NULL,
		user_id character varying NOT NULL,
		CONSTRAINT urls_pkey PRIMARY KEY (id),
		CONSTRAINT url_id_unique UNIQUE (url_id)
	)
	`
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	_, err := db.ExecContext(ctx, createScript)
	return err
}
