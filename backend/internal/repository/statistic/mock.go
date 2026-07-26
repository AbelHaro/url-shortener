package statistic

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/google/uuid"
)

var _ Repository = (*MockRepository)(nil)

// MockRepository stores statistics in memory for unit tests.
type MockRepository struct {
	mu    sync.RWMutex
	stats map[uuid.UUID][]domain.URLStatistics
}

// NewMockRepository creates an empty in-memory statistics repository.
func NewMockRepository() *MockRepository {
	return &MockRepository{stats: make(map[uuid.UUID][]domain.URLStatistics)}
}

// RecordClick stores a click event.
func (m *MockRepository) RecordClick(_ context.Context, stat *domain.URLStatistics) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stats[stat.URLID] = append(m.stats[stat.URLID], *stat)
	return nil
}

// GetClickCount returns the all-time number of clicks for a URL.
func (m *MockRepository) GetClickCount(_ context.Context, urlID uuid.UUID) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return int64(len(m.stats[urlID])), nil
}

// GetLastAccessAt returns the latest click time, if one exists.
func (m *MockRepository) GetLastAccessAt(_ context.Context, urlID uuid.UUID) (*time.Time, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var latest *time.Time
	for _, stat := range m.stats[urlID] {
		if latest == nil || stat.ClickedAt.After(*latest) {
			value := stat.ClickedAt
			latest = &value
		}
	}
	return latest, nil
}

// GetClicksByDay returns UTC daily click counts since the supplied day.
func (m *MockRepository) GetClicksByDay(_ context.Context, urlID uuid.UUID, since time.Time) ([]domain.DailyClick, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	counts := make(map[time.Time]int64)
	for _, stat := range m.stats[urlID] {
		day := stat.ClickedAt.UTC().Truncate(24 * time.Hour)
		if day.Before(since) {
			continue
		}
		counts[day]++
	}

	days := make([]domain.DailyClick, 0, len(counts))
	for day, clicks := range counts {
		days = append(days, domain.DailyClick{Date: day, Clicks: clicks})
	}
	sort.Slice(days, func(i, j int) bool {
		return days[i].Date.Before(days[j].Date)
	})
	return days, nil
}

// GetTopReferrers returns the most frequent referrer hostnames.
func (m *MockRepository) GetTopReferrers(_ context.Context, urlID uuid.UUID, limit int) ([]domain.ReferrerCount, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	counts := make(map[string]int64)
	for _, stat := range m.stats[urlID] {
		counts[stat.Referer]++
	}

	referrers := make([]domain.ReferrerCount, 0, len(counts))
	for referrer, clicks := range counts {
		referrers = append(referrers, domain.ReferrerCount{Referrer: referrer, Clicks: clicks})
	}
	sort.Slice(referrers, func(i, j int) bool {
		if referrers[i].Clicks == referrers[j].Clicks {
			return referrers[i].Referrer < referrers[j].Referrer
		}
		return referrers[i].Clicks > referrers[j].Clicks
	})
	if len(referrers) > limit {
		referrers = referrers[:limit]
	}
	return referrers, nil
}

// GetRecentClicks returns the newest click events first.
func (m *MockRepository) GetRecentClicks(_ context.Context, urlID uuid.UUID, limit int) ([]domain.URLStatistics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clicks := append([]domain.URLStatistics(nil), m.stats[urlID]...)
	sort.Slice(clicks, func(i, j int) bool {
		return clicks[i].ClickedAt.After(clicks[j].ClickedAt)
	})
	if len(clicks) > limit {
		clicks = clicks[:limit]
	}
	return clicks, nil
}
