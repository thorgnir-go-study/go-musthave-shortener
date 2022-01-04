package storage

import (
	"context"
	"database/sql"
	"errors"
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
	getByURLIDStmt     *sqlx.Stmt
	selectByUserIDStmt *sqlx.Stmt
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
	if insertStmt, err = db.PrepareNamed(`
WITH new_link AS (
    INSERT INTO urls(url_id, original_url, user_id) VALUES (:url_id, :original_url, :user_id)
    ON CONFLICT(original_url_unique) DO NOTHING
    RETURNING url_id
) SELECT COALESCE(
    (SELECT url_id FROM new_link),
    (SELECT url_id FROM urls WHERE original_url = :original_url)
)
`); err != nil {
		return err
	}

	if getByURLIDStmt, err = db.Preparex(`select url_id, original_url, user_id  from urls where url_id = $1`); err != nil {
		return err
	}

	if selectByUserIDStmt, err = db.Preparex(`select url_id, original_url, user_id  from urls where user_id=$1`); err != nil {
		return err
	}

	return nil
}

func (s *dbURLStorage) Store(urlEntity URLEntity) error {
	row := insertStmt.QueryRow(&urlEntity)

	var urlID string
	err := row.Scan(&urlID)
	if err != nil {
		return err
	}
	if urlID != urlEntity.ID {
		return NewErrURLExists(urlID)
	}
	return nil
}

func (s *dbURLStorage) StoreBatch(ctx context.Context, entitiesBatch []URLEntity) error {
	// длительность таймаутов возможно нужно вынести в настройки. или в какие-то константы с понятными названиями и собранные плюс-минус в одном месте.
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	tx, err := s.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback() //nolint:errcheck

	txInsertStmt := tx.NamedStmtContext(ctx, insertStmt)
	for _, entity := range entitiesBatch {
		if _, err = txInsertStmt.Exec(&entity); err != nil {
			return err
		}
	}
	if err = tx.Commit(); err != nil {
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
	err := getByURLIDStmt.GetContext(ctx, &entity, key)
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
	if err := selectByUserIDStmt.SelectContext(ctx, &result, userID); err != nil {
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
		CONSTRAINT url_id_unique UNIQUE (url_id),
		CONSTRAINT original_url_unique UNIQUE (original_url)
	)
	`
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	_, err := db.ExecContext(ctx, createScript)
	return err
}
