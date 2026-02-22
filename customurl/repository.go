package customurl

import (
	"context"
	"database/sql"
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

}

func (r *postgresRepository) GetUrl(ctx context.Context, customUrl string) (string, error) {

}
