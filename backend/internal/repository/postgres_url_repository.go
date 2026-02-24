package repository

import (
	"context"
	"errors"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PostgresURLRepository struct {
	db *gorm.DB
}

func NewPostgresURLRepository(db *gorm.DB) URLRepository {
	return &PostgresURLRepository{db: db}
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

func (repo PostgresURLRepository) FindByID(id uuid.UUID) (*domain.URL, error) {
	ctx := context.Background()
	url, err := gorm.G[domain.URL](repo.db).Where("id = ?", id).First(ctx)

	if err != nil {
		return nil, err
	}

	return &url, nil
}

func (repo PostgresURLRepository) DeleteByOriginalURL(originalURL string) error {
	ctx := context.Background()

	rowsAffected, err := gorm.G[domain.URL](repo.db).Where("original_url = ?", originalURL).Delete(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrURLNotFound
	}

	return nil
}

func (repo PostgresURLRepository) DeleteByShortURL(shortURL string) error {
	ctx := context.Background()

	rowsAffected, err := gorm.G[domain.URL](repo.db).Where("short_url = ?", shortURL).Delete(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrURLNotFound
	}

	return nil
}

func (repo PostgresURLRepository) DeleteByID(id uuid.UUID) error {
	ctx := context.Background()

	rowsAffected, err := gorm.G[domain.URL](repo.db).Where("id = ?", id).Delete(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrURLNotFound
	}

	return nil
}
