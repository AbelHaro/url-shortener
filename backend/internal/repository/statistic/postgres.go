package statistic

import (
	"context"
	"fmt"
	"time"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var _ Repository = (*PostgresRepository)(nil)

// PostgresRepository persists and aggregates URL click statistics.
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository creates a PostgreSQL statistics repository.
func NewPostgresRepository(db *gorm.DB) Repository {
	return &PostgresRepository{db: db}
}

// RecordClick persists a click event.
func (repo PostgresRepository) RecordClick(ctx context.Context, stat *domain.URLStatistics) error {
	if err := repo.db.WithContext(ctx).Create(stat).Error; err != nil {
		return fmt.Errorf("record click: %w", domain.ErrInternal)
	}
	return nil
}

// GetClickCount returns the all-time click total for a URL.
func (repo PostgresRepository) GetClickCount(ctx context.Context, urlID uuid.UUID) (int64, error) {
	var count int64
	if err := repo.db.WithContext(ctx).
		Model(&domain.URLStatistics{}).
		Where("url_id = ?", urlID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count clicks: %w", domain.ErrInternal)
	}
	return count, nil
}

// GetLastAccessAt returns the latest click timestamp, if any.
func (repo PostgresRepository) GetLastAccessAt(ctx context.Context, urlID uuid.UUID) (*time.Time, error) {
	var result struct {
		LastClickedAt *time.Time
	}
	if err := repo.db.WithContext(ctx).
		Model(&domain.URLStatistics{}).
		Select("MAX(clicked_at) AS last_clicked_at").
		Where("url_id = ?", urlID).
		Scan(&result).Error; err != nil {
		return nil, fmt.Errorf("get last click: %w", domain.ErrInternal)
	}
	return result.LastClickedAt, nil
}

// GetClicksByDay returns UTC daily click totals from the supplied day onward.
func (repo PostgresRepository) GetClicksByDay(ctx context.Context, urlID uuid.UUID, since time.Time) ([]domain.DailyClick, error) {
	var clicks []domain.DailyClick
	if err := repo.db.WithContext(ctx).
		Model(&domain.URLStatistics{}).
		Select("DATE(clicked_at AT TIME ZONE 'UTC') AS date, COUNT(*) AS clicks").
		Where("url_id = ? AND clicked_at >= ?", urlID, since).
		Group("DATE(clicked_at AT TIME ZONE 'UTC')").
		Order("date ASC").
		Scan(&clicks).Error; err != nil {
		return nil, fmt.Errorf("get daily clicks: %w", domain.ErrInternal)
	}
	return clicks, nil
}

// GetTopReferrers returns the most frequent referrers in descending order.
func (repo PostgresRepository) GetTopReferrers(ctx context.Context, urlID uuid.UUID, limit int) ([]domain.ReferrerCount, error) {
	var referrers []domain.ReferrerCount
	if err := repo.db.WithContext(ctx).
		Model(&domain.URLStatistics{}).
		Select("referer AS referrer, COUNT(*) AS clicks").
		Where("url_id = ?", urlID).
		Group("referer").
		Order("clicks DESC, referrer ASC").
		Limit(limit).
		Scan(&referrers).Error; err != nil {
		return nil, fmt.Errorf("get top referrers: %w", domain.ErrInternal)
	}
	return referrers, nil
}

// GetRecentClicks returns the latest click events in descending order.
func (repo PostgresRepository) GetRecentClicks(ctx context.Context, urlID uuid.UUID, limit int) ([]domain.URLStatistics, error) {
	var clicks []domain.URLStatistics
	if err := repo.db.WithContext(ctx).
		Select("id", "url_id", "clicked_at", "referer").
		Where("url_id = ?", urlID).
		Order("clicked_at DESC").
		Limit(limit).
		Find(&clicks).Error; err != nil {
		return nil, fmt.Errorf("get recent clicks: %w", domain.ErrInternal)
	}
	return clicks, nil
}
