package shorturl

import "context"

type Service interface {
	CreateShortUrl(ctx context.Context, originalUrl, customUrl string) (string, error)
	GetOriginalUrl(ctx context.Context, customUrl string) (string, error)
}

type service struct {
	repository Repository
}

func NewService(r Repository) Service {
	return &service{repository: r}
}

func (s *service) CreateShortUrl(ctx context.Context, originalUrl, customUrl string) (string, error) {

	if err := s.repository.Create(ctx, originalUrl, customUrl); err != nil {
		return "", err
	}

	return customUrl, nil
}

func (s *service) GetOriginalUrl(ctx context.Context, customUrl string) (string, error) {
	return s.repository.GetUrl(ctx, customUrl)
}
