package statistic

import (
	"context"
	"time"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var _ Repository = (*PostgresRepository)(nil)

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) Repository {
	return &PostgresRepository{db: db}
}

func (repo PostgresRepository) RecordClick(stat *domain.URLStatistics) error {
	ctx := context.Background()

	result := repo.db.WithContext(ctx).Create(stat)
	if result.Error != nil {
		return domain.ErrInternal
	}
	return nil
}

func (repo PostgresRepository) GetStatistics(urlID string) ([]*domain.URLStatistics, error) {
	ctx := context.Background()

	urlIDParsed, err := uuid.Parse(urlID)
	if err != nil {
		return nil, domain.ErrInvalidURL
	}

	var statistics []*domain.URLStatistics
	result := repo.db.WithContext(ctx).Where("url_id = ?", urlIDParsed).Order("clicked_at DESC").Find(&statistics)
	if result.Error != nil {
		return nil, domain.ErrInternal
	}

	return statistics, nil
}

func (repo PostgresRepository) GetClickCount(urlID string) (int64, error) {
	ctx := context.Background()

	urlIDParsed, err := uuid.Parse(urlID)
	if err != nil {
		return 0, domain.ErrInvalidURL
	}

	var count int64
	result := repo.db.WithContext(ctx).Model(&domain.URLStatistics{}).Where("url_id = ?", urlIDParsed).Count(&count)
	if result.Error != nil {
		return 0, domain.ErrInternal
	}
	return count, nil
}

func (repo PostgresRepository) GetLastAccessAt(urlID string) (time.Time, error) {
	ctx := context.Background()

	urlIDParsed, err := uuid.Parse(urlID)
	if err != nil {
		return time.Time{}, domain.ErrInvalidURL
	}

	var lastAccess time.Time
	result := repo.db.WithContext(ctx).Model(&domain.URLStatistics{}).Select("MAX(clicked_at)").Where("url_id = ?", urlIDParsed).Scan(&lastAccess)
	if result.Error != nil {
		return time.Time{}, domain.ErrInternal
	}

	return lastAccess, nil
}
