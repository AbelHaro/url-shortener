package url

import (
	"context"
	"errors"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Ensure PostgresRepository implements the Repository interface
var _ Repository = (*PostgresRepository)(nil)

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) Repository {
	return &PostgresRepository{db: db}
}

func (repo PostgresRepository) Store(url *domain.URL) (*domain.URL, error) {
	if err := repo.db.Create(url).Error; err != nil {
		return nil, domain.ErrInternal
	}
	return url, nil
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

func (repo PostgresRepository) FindByIDAndUserID(ctx context.Context, id, userID uuid.UUID) (*domain.URL, error) {
	storedURL, err := gorm.G[domain.URL](repo.db).
		Where("id = ? AND user_id = ?", id, userID).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, domain.ErrInternal
	}
	return &storedURL, nil
}

func (repo PostgresRepository) UpdateOriginalURL(ctx context.Context, id, userID uuid.UUID, originalURL string) (*domain.URL, error) {
	result := repo.db.WithContext(ctx).
		Model(&domain.URL{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("original_url", originalURL)
	if result.Error != nil {
		return nil, domain.ErrInternal
	}
	if result.RowsAffected == 0 {
		return nil, domain.ErrURLNotFound
	}
	return repo.FindByIDAndUserID(ctx, id, userID)
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

func (repo PostgresRepository) DeleteByIDAndUserID(ctx context.Context, id, userID uuid.UUID) error {
	rowsAffected, err := gorm.G[domain.URL](repo.db).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(ctx)
	if err != nil {
		return domain.ErrInternal
	}
	if rowsAffected == 0 {
		return domain.ErrURLNotFound
	}
	return nil
}

func (repo PostgresRepository) FindAllByUserID(userID uuid.UUID) ([]domain.URL, error) {
	ctx := context.Background()

	urls, err := gorm.G[domain.URL](repo.db).Where("user_id = ?", userID).Find(ctx)

	if err != nil {
		return nil, domain.ErrInternal
	}

	return urls, nil
}
