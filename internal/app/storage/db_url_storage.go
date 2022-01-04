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
	batchInsertStmt    *sqlx.NamedStmt
	getByURLIDStmt     *sqlx.Stmt
	selectByUserIDStmt *sqlx.Stmt
)

func NewDBURLStorage(ctx context.Context, connectionString string) (*dbURLStorage, error) {
	db, err := sqlx.Open("pgx", connectionString)
	if err != nil {
		return nil, err
	}

	if err = createTables(ctx, db); err != nil {
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
    ON CONFLICT(original_url) DO NOTHING
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

	if batchInsertStmt, err = db.PrepareNamed(`INSERT INTO urls(url_id, original_url, user_id) VALUES (:url_id, :original_url, :user_id)`); err != nil {
		return err
	}

	return nil
}

func (s *dbURLStorage) Store(ctx context.Context, urlEntity URLEntity) error {
	// длительность таймаутов возможно нужно вынести в настройки. или в какие-то константы с понятными названиями и собранные плюс-минус в одном месте.
	innerCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	row := insertStmt.QueryRowContext(innerCtx, &urlEntity)
	var urlID string
	err := row.Scan(&urlID)
	if err != nil {
		return err
	}
	// Можно было бы использовать более простой запрос на вставку, ловить ошибку, анализировать ее на нарушение конкретного констрейнта
	// Но тогда нужно было бы делать дополнительный запрос в БД для получения идентификатора конфликтующей записи
	// Как лучше - большой вопрос, зависит от частоты возникновения конфликтов в реальном мире
	if urlID != urlEntity.ID {
		return NewErrURLExists(urlID)
	}
	return nil
}

func (s *dbURLStorage) StoreBatch(ctx context.Context, entitiesBatch []URLEntity) error {
	// длительность таймаутов возможно нужно вынести в настройки. или в какие-то константы с понятными названиями и собранные плюс-минус в одном месте.
	innerCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	tx, err := s.DB.BeginTxx(innerCtx, nil)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback() //nolint:errcheck

	txInsertStmt := tx.NamedStmtContext(ctx, batchInsertStmt)
	for _, entity := range entitiesBatch {
		// TODO: тут нет проверки ошибки на нарушение уникальности оригинального урла и соответствующей обработки (выброса ErrURLExists)
		// Поэтому при попытке сохранить дубликат ссылки весь батч просто упадет с ошибкой (но по ТЗ и не надо было обрабатывать).
		// Если будет время - надо допилить.
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

func (s *dbURLStorage) Load(ctx context.Context, key string) (URLEntity, error) {
	innerCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	var entity URLEntity
	err := getByURLIDStmt.GetContext(innerCtx, &entity, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return URLEntity{}, ErrURLNotFound
		}
		return URLEntity{}, err
	}
	return entity, nil
}

func (s *dbURLStorage) LoadByUserID(ctx context.Context, userID string) ([]URLEntity, error) {
	innerCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	var result []URLEntity
	if err := selectByUserIDStmt.SelectContext(innerCtx, &result, userID); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *dbURLStorage) Ping(ctx context.Context) error {
	innerCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	return s.DB.PingContext(innerCtx)
}

func createTables(ctx context.Context, db *sqlx.DB) error {
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
	innerCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	_, err := db.ExecContext(innerCtx, createScript)
	return err
}
