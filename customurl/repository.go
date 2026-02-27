package customurl

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

type Repository interface {
	Create(ctx context.Context, originalUrl, customUrl string) error
	GetUrl(ctx context.Context, customUrl string) (string, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(url string) (Repository, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &postgresRepository{db}, nil
}

func (r *postgresRepository) Create(ctx context.Context, originalUrl, customUrl string) error {
	query := `
		INSERT INTO urls (original_url, custom_url)
		VALUES ($1, $2)
		`
	if _, err := r.db.ExecContext(ctx, query, originalUrl, customUrl); err != nil {
		return err
	}

	return nil
}

func (r *postgresRepository) GetUrl(ctx context.Context, customUrl string) (string, error) {
	query := `
		SELECT original_url
		FROM urls
		WHERE custom_url = $1
	`

	var originalUrl string

	if err := r.db.QueryRowContext(ctx, query, customUrl).Scan(&originalUrl); err != nil {
		if err == sql.ErrNoRows {
			return "", sql.ErrNoRows
		}
		return "", err
	}

	return originalUrl, nil
}
