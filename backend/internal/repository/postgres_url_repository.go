package repository

import (
	"context"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"gorm.io/gorm"
)

type PostgresURLRepository struct {
	db *gorm.DB
}

func (repo PostgresURLRepository) Store(url *domain.URL) error {
	ctx := context.Background()
	return gorm.G[domain.URL](repo.db).Create(ctx, url)

}

func (repo PostgresURLRepository) FindByOriginalURL(originalURL string) (*domain.URL, error) {
	ctx := context.Background()

	url, err := gorm.G[domain.URL](repo.db).Where("original_url = ?", originalURL).First(ctx)

	if err != nil {
		return nil, err
	}

	return &url, nil
}

func (repo PostgresURLRepository) FindByShortURL(shortURL string) (*domain.URL, error) {
	ctx := context.Background()

	url, err := gorm.G[domain.URL](repo.db).Where("short_url = ?", shortURL).First(ctx)

	if err != nil {
		return nil, err
	}

	return &url, nil
}

func (repo PostgresURLRepository) FindByID(id string) (*domain.URL, error) {
	//TODO implement me
	panic("implement me")
}

func (repo PostgresURLRepository) DeleteByOriginalURL(originalURL string) error {
	//TODO implement me
	panic("implement me")
}

func (repo PostgresURLRepository) DeleteByShortURL(shortURL string) error {
	//TODO implement me
	panic("implement me")
}

func (repo PostgresURLRepository) DeleteByID(id string) error {
	//TODO implement me
	panic("implement me")
}

func NewPostgresURLRepository(db *gorm.DB) URLRepository {
	return &PostgresURLRepository{db: db}
}
