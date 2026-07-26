package statistic

import (
	"context"
	"time"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/google/uuid"
)

type Repository interface {
	RecordClick(ctx context.Context, stat *domain.URLStatistics) error
	GetClickCount(ctx context.Context, urlID uuid.UUID) (int64, error)
	GetLastAccessAt(ctx context.Context, urlID uuid.UUID) (*time.Time, error)
	GetClicksByDay(ctx context.Context, urlID uuid.UUID, since time.Time) ([]domain.DailyClick, error)
	GetTopReferrers(ctx context.Context, urlID uuid.UUID, limit int) ([]domain.ReferrerCount, error)
	GetRecentClicks(ctx context.Context, urlID uuid.UUID, limit int) ([]domain.URLStatistics, error)
}
