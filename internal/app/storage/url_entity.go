package storage

type URLEntity struct {
	ID          string `db:"url_id"`
	OriginalURL string `db:"original_url"`
	UserID      string `db:"user_id"`
}
