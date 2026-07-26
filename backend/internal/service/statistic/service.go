// Package statistic records and aggregates URL click analytics.
package statistic

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	statisticRepo "github.com/AbelHaro/url-shortener/backend/internal/repository/statistic"
	"github.com/google/uuid"
)

const (
	dashboardDays       = 30
	topReferrerLimit    = 5
	recentClickLimit    = 10
	directReferrerLabel = "Direct"
)

// Dashboard contains the aggregated analytics shown for one shortened URL.
type Dashboard struct {
	TotalClicks   int64
	LastClickedAt *time.Time
	ClicksByDay   []domain.DailyClick
	TopReferrers  []domain.ReferrerCount
	RecentClicks  []domain.URLStatistics
}

// Option configures a statistics service.
type Option func(*Service)

// WithClock overrides the service clock, primarily for deterministic tests.
func WithClock(now func() time.Time) Option {
	return func(service *Service) {
		service.now = now
	}
}

// Service records click events and builds analytics dashboards.
type Service struct {
	repo statisticRepo.Repository
	now  func() time.Time
}

// NewService creates a statistics service.
func NewService(repo statisticRepo.Repository, options ...Option) *Service {
	service := &Service{
		repo: repo,
		now:  time.Now,
	}
	for _, option := range options {
		option(service)
	}
	return service
}

// RecordClick validates and stores one URL click.
func (service *Service) RecordClick(ctx context.Context, rawURLID, rawReferrer string) error {
	urlID, err := uuid.Parse(rawURLID)
	if err != nil {
		return fmt.Errorf("parse URL ID: %w", domain.ErrInvalidID)
	}

	click := &domain.URLStatistics{
		ID:        uuid.New(),
		URLID:     urlID,
		ClickedAt: service.now().UTC(),
		Referer:   normalizeReferrer(rawReferrer),
	}
	if err := service.repo.RecordClick(ctx, click); err != nil {
		return fmt.Errorf("store click: %w", err)
	}
	return nil
}

// GetDashboard returns all aggregates required by the analytics page.
func (service *Service) GetDashboard(ctx context.Context, rawURLID string) (*Dashboard, error) {
	urlID, err := uuid.Parse(rawURLID)
	if err != nil {
		return nil, fmt.Errorf("parse URL ID: %w", domain.ErrInvalidID)
	}

	today := service.now().UTC().Truncate(24 * time.Hour)
	since := today.AddDate(0, 0, -(dashboardDays - 1))

	totalClicks, err := service.repo.GetClickCount(ctx, urlID)
	if err != nil {
		return nil, fmt.Errorf("get click count: %w", err)
	}
	lastClickedAt, err := service.repo.GetLastAccessAt(ctx, urlID)
	if err != nil {
		return nil, fmt.Errorf("get last click: %w", err)
	}
	recordedDays, err := service.repo.GetClicksByDay(ctx, urlID, since)
	if err != nil {
		return nil, fmt.Errorf("get daily clicks: %w", err)
	}
	topReferrers, err := service.repo.GetTopReferrers(ctx, urlID, topReferrerLimit)
	if err != nil {
		return nil, fmt.Errorf("get top referrers: %w", err)
	}
	recentClicks, err := service.repo.GetRecentClicks(ctx, urlID, recentClickLimit)
	if err != nil {
		return nil, fmt.Errorf("get recent clicks: %w", err)
	}

	countsByDay := make(map[string]int64, len(recordedDays))
	for _, day := range recordedDays {
		countsByDay[day.Date.UTC().Format(time.DateOnly)] = day.Clicks
	}
	clicksByDay := make([]domain.DailyClick, 0, dashboardDays)
	for offset := 0; offset < dashboardDays; offset++ {
		day := since.AddDate(0, 0, offset)
		clicksByDay = append(clicksByDay, domain.DailyClick{
			Date:   day,
			Clicks: countsByDay[day.Format(time.DateOnly)],
		})
	}

	return &Dashboard{
		TotalClicks:   totalClicks,
		LastClickedAt: lastClickedAt,
		ClicksByDay:   clicksByDay,
		TopReferrers:  topReferrers,
		RecentClicks:  recentClicks,
	}, nil
}

func normalizeReferrer(rawReferrer string) string {
	rawReferrer = strings.TrimSpace(rawReferrer)
	if rawReferrer == "" {
		return directReferrerLabel
	}

	parsed, err := url.ParseRequestURI(rawReferrer)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" {
		return directReferrerLabel
	}
	return strings.ToLower(parsed.Hostname())
}
