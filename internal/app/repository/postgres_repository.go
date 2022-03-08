package repository

import (
	"context"
	"database/sql"
	"errors"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"time"
)

type postgresURLRepository struct {
	DB *sqlx.DB
}

var (
	insertStmt         *sqlx.NamedStmt
	batchInsertStmt    *sqlx.NamedStmt
	getByURLIDStmt     *sqlx.Stmt
	selectByUserIDStmt *sqlx.Stmt
	batchDeleteStmt    *sqlx.Stmt
)

func NewPostgresURLRepository(ctx context.Context, connectionString string) (*postgresURLRepository, error) {
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

	return &postgresURLRepository{DB: db}, nil
}

//goland:noinspection SqlNoDataSourceInspection,SqlResolve
func prepareStatements(db *sqlx.DB) error {
	var err error
	if insertStmt, err = db.PrepareNamed(`
WITH new_link AS (
    INSERT INTO urls(url_id, original_url, user_id, deleted) VALUES (:url_id, :original_url, :user_id, :deleted)
    ON CONFLICT(original_url) DO NOTHING
    RETURNING url_id
) SELECT COALESCE(
    (SELECT url_id FROM new_link),
    (SELECT url_id FROM urls WHERE original_url = :original_url)
)
`); err != nil {
		return err
	}

	if getByURLIDStmt, err = db.Preparex(`select url_id, original_url, user_id, deleted  from urls where url_id = $1`); err != nil {
		return err
	}

	if selectByUserIDStmt, err = db.Preparex(`select url_id, original_url, user_id, deleted  from urls where user_id=$1`); err != nil {
		return err
	}

	if batchInsertStmt, err = db.PrepareNamed(`INSERT INTO urls(url_id, original_url, user_id, deleted) VALUES (:url_id, :original_url, :user_id, :deleted)`); err != nil {
		return err
	}

	if batchDeleteStmt, err = db.Preparex(`update urls set deleted=true where user_id=$1 and url_id = any($2)`); err != nil {
		return err
	}

	return nil
}

func (s *postgresURLRepository) Store(ctx context.Context, urlEntity URLEntity) error {
	row := insertStmt.QueryRowContext(ctx, &urlEntity)
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

func (s *postgresURLRepository) StoreBatch(ctx context.Context, entitiesBatch []URLEntity) error {
	tx, err := s.DB.BeginTxx(ctx, nil)
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

func (s *postgresURLRepository) Load(ctx context.Context, key string) (URLEntity, error) {
	var entity URLEntity
	err := getByURLIDStmt.GetContext(ctx, &entity, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return URLEntity{}, ErrURLNotFound
		}
		return URLEntity{}, err
	}
	return entity, nil
}

func (s *postgresURLRepository) LoadByUserID(ctx context.Context, userID string) ([]URLEntity, error) {
	var result []URLEntity
	if err := selectByUserIDStmt.SelectContext(ctx, &result, userID); err != nil {
		return nil, err
	}
	return result, nil
}

// DeleteURLs implements URLRepository.DeleteURLs
func (s *postgresURLRepository) DeleteURLs(ctx context.Context, userID string, ids []string) error {
	if _, err := batchDeleteStmt.ExecContext(ctx, userID, ids); err != nil {
		return err
	}
	return nil
}

func (s *postgresURLRepository) Ping(ctx context.Context) error {
	return s.DB.PingContext(ctx)
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
		deleted boolean NOT NULL DEFAULT false,
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
