package statistic

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	statisticRepo "github.com/AbelHaro/url-shortener/backend/internal/repository/statistic"
	"github.com/google/uuid"
)

func TestService_RecordClick(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.July, 26, 12, 30, 0, 0, time.UTC)
	urlID := uuid.New()

	tests := []struct {
		name        string
		urlID       string
		referrer    string
		wantReferer string
		wantErr     bool
	}{
		{name: "hostname", urlID: urlID.String(), referrer: "https://News.Example.com/article?id=1", wantReferer: "news.example.com"},
		{name: "direct", urlID: urlID.String(), referrer: "", wantReferer: "Direct"},
		{name: "invalid referrer", urlID: urlID.String(), referrer: "not a URL", wantReferer: "Direct"},
		{name: "invalid URL ID", urlID: "invalid", referrer: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := statisticRepo.NewMockRepository()
			service := NewService(repo, WithClock(func() time.Time { return now }))

			err := service.RecordClick(context.Background(), tt.urlID, tt.referrer)
			if (err != nil) != tt.wantErr {
				t.Fatalf("RecordClick() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			clicks, err := repo.GetRecentClicks(context.Background(), urlID, 1)
			if err != nil {
				t.Fatalf("GetRecentClicks() error = %v", err)
			}
			if len(clicks) != 1 {
				t.Fatalf("GetRecentClicks() length = %d, want 1", len(clicks))
			}
			if clicks[0].Referer != tt.wantReferer {
				t.Errorf("Referer = %q, want %q", clicks[0].Referer, tt.wantReferer)
			}
			if !clicks[0].ClickedAt.Equal(now) {
				t.Errorf("ClickedAt = %v, want %v", clicks[0].ClickedAt, now)
			}
		})
	}
}

func TestService_GetDashboard(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	now := time.Date(2026, time.July, 26, 17, 0, 0, 0, time.UTC)
	urlID := uuid.New()
	repo := statisticRepo.NewMockRepository()
	service := NewService(repo, WithClock(func() time.Time { return now }))

	clicks := []struct {
		at       time.Time
		referrer string
	}{
		{at: now.AddDate(0, 0, -29), referrer: "alpha.example"},
		{at: now.AddDate(0, 0, -1), referrer: "alpha.example"},
		{at: now, referrer: "beta.example"},
	}
	for _, click := range clicks {
		err := repo.RecordClick(ctx, &domain.URLStatistics{
			ID:        uuid.New(),
			URLID:     urlID,
			ClickedAt: click.at,
			Referer:   click.referrer,
		})
		if err != nil {
			t.Fatalf("RecordClick() error = %v", err)
		}
	}

	dashboard, err := service.GetDashboard(ctx, urlID.String())
	if err != nil {
		t.Fatalf("GetDashboard() error = %v", err)
	}

	if dashboard.TotalClicks != 3 {
		t.Errorf("TotalClicks = %d, want 3", dashboard.TotalClicks)
	}
	if dashboard.LastClickedAt == nil || !dashboard.LastClickedAt.Equal(now) {
		t.Errorf("LastClickedAt = %v, want %v", dashboard.LastClickedAt, now)
	}
	if len(dashboard.ClicksByDay) != 30 {
		t.Fatalf("ClicksByDay length = %d, want 30", len(dashboard.ClicksByDay))
	}
	if dashboard.ClicksByDay[0].Clicks != 1 || dashboard.ClicksByDay[28].Clicks != 1 || dashboard.ClicksByDay[29].Clicks != 1 {
		t.Errorf("ClicksByDay did not contain expected zero-filled series: %+v", dashboard.ClicksByDay)
	}
	if len(dashboard.TopReferrers) != 2 || dashboard.TopReferrers[0].Referrer != "alpha.example" || dashboard.TopReferrers[0].Clicks != 2 {
		t.Errorf("TopReferrers = %+v, want alpha.example first with 2 clicks", dashboard.TopReferrers)
	}
	if len(dashboard.RecentClicks) != 3 || !dashboard.RecentClicks[0].ClickedAt.Equal(now) {
		t.Errorf("RecentClicks = %+v, want newest click first", dashboard.RecentClicks)
	}
}

func TestService_GetDashboardErrors(t *testing.T) {
	t.Parallel()

	service := NewService(statisticRepo.NewMockRepository())
	_, err := service.GetDashboard(context.Background(), "invalid")
	if !errors.Is(err, domain.ErrInvalidID) {
		t.Fatalf("GetDashboard() error = %v, want ErrInvalidID", err)
	}
}

func TestService_GetDashboardEmpty(t *testing.T) {
	t.Parallel()

	service := NewService(
		statisticRepo.NewMockRepository(),
		WithClock(func() time.Time {
			return time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)
		}),
	)
	dashboard, err := service.GetDashboard(context.Background(), uuid.New().String())
	if err != nil {
		t.Fatalf("GetDashboard() error = %v", err)
	}
	if dashboard.TotalClicks != 0 || dashboard.LastClickedAt != nil {
		t.Errorf("empty summary = total %d, last %v", dashboard.TotalClicks, dashboard.LastClickedAt)
	}
	if len(dashboard.ClicksByDay) != 30 {
		t.Fatalf("ClicksByDay length = %d, want 30", len(dashboard.ClicksByDay))
	}
	for _, day := range dashboard.ClicksByDay {
		if day.Clicks != 0 {
			t.Errorf("ClicksByDay contains non-zero item: %+v", day)
		}
	}
	if len(dashboard.TopReferrers) != 0 || len(dashboard.RecentClicks) != 0 {
		t.Errorf("empty dashboard contains activity: %+v", dashboard)
	}
}
