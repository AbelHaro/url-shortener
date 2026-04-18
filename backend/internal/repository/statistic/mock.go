package statistic

import (
	"time"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
)

var _ Repository = (*MockRepository)(nil)

type MockRepository struct {
	stats map[string][]*domain.URLStatistics
}

func NewMockRepository() Repository {
	return &MockRepository{stats: make(map[string][]*domain.URLStatistics)}
}

func (m *MockRepository) RecordClick(stat *domain.URLStatistics) error {
	m.stats[stat.UrlID.String()] = append(m.stats[stat.UrlID.String()], stat)
	return nil
}

func (m *MockRepository) GetStatistics(urlID string) ([]*domain.URLStatistics, error) {
	if stats, ok := m.stats[urlID]; ok {
		return stats, nil
	}
	return nil, domain.ErrUrlStatisticsNotFound
}

func (m *MockRepository) GetClickCount(urlID string) (int64, error) {
	return int64(len(m.stats[urlID])), nil
}

func (m *MockRepository) GetLastAccessAt(urlID string) (time.Time, error) {
	stats := m.stats[urlID]
	if len(stats) == 0 {
		return time.Time{}, nil
	}
	return stats[len(stats)-1].ClickedAt, nil
}
