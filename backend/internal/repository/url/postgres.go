package url

import (
	"context"
	"errors"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) Repository {
	return &PostgresRepository{db: db}
}

func (repo PostgresRepository) Store(url *domain.URL) (*domain.URL, error) {
	ctx := context.Background()

	err := repo.db.
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "original_url"}},
			DoNothing: true,
		}).
		Create(url).Error

	if err != nil {
		return nil, domain.ErrInternal
	}

	storedUrl, err := gorm.G[domain.URL](repo.db).Where("original_url = ?", url.OriginalURL).First(ctx)
	if err != nil {
		return nil, domain.ErrInternal
	}

	return &storedUrl, nil
}

func (repo PostgresRepository) FindByOriginalURL(originalURL string) (*domain.URL, error) {
	ctx := context.Background()

	url, err := gorm.G[domain.URL](repo.db).Where("original_url = ?", originalURL).First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrURLNotFound
		}
		return nil, domain.ErrInternal
	}

	return &url, nil
}

func (repo PostgresRepository) FindByShortCode(shortCode string) (*domain.URL, error) {
	ctx := context.Background()

	url, err := gorm.G[domain.URL](repo.db).Where("short_code = ?", shortCode).First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &url, nil
}

func (repo PostgresRepository) FindByID(id uuid.UUID) (*domain.URL, error) {
	ctx := context.Background()
	url, err := gorm.G[domain.URL](repo.db).Where("id = ?", id).First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &url, nil
}

func (repo PostgresRepository) DeleteByOriginalURL(originalURL string) error {
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

func (repo PostgresRepository) DeleteByShortCode(shortCode string) error {
	ctx := context.Background()

	rowsAffected, err := gorm.G[domain.URL](repo.db).Where("short_code = ?", shortCode).Delete(ctx)

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

func (repo PostgresRepository) DeleteByID(id uuid.UUID) error {
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
