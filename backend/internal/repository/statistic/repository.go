package statistic

import (
	"time"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
)

type Repository interface {
	RecordClick(stat *domain.URLStatistics) error
	GetStatistics(urlID string) ([]*domain.URLStatistics, error)
	GetClickCount(urlID string) (int64, error)
	GetLastAccessAt(urlID string) (time.Time, error)
}
