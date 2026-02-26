package counter

import (
	"context"
	"errors"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"gorm.io/gorm"
)

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) Repository {
	return &PostgresRepository{db: db}
}

func (repo *PostgresRepository) GetCounter() (*domain.Counter, error) {
	ctx := context.Background()

	counter, err := gorm.G[domain.Counter](repo.db).First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newCounter := &domain.Counter{
				Counter: 0,
			}
			if err := gorm.G[domain.Counter](repo.db).Create(ctx, newCounter); err != nil {
				return nil, err
			}
			return newCounter, nil
		}
		return nil, err
	}

	return &counter, nil
}

func (repo *PostgresRepository) UpdateCounter(counter int64) error {
	ctx := context.Background()

	hashCounter, err := gorm.G[domain.Counter](repo.db).First(ctx)

	if err != nil {
		return err
	}

	rowsAffected, err := gorm.G[domain.Counter](repo.db).Where("id = ?", hashCounter.ID).Update(ctx, "counter", counter)

	if err != nil {
		return err
	}

	if rowsAffected != 1 {

	}

	return err
}
