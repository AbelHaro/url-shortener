package statistic

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	idsrangesRepo "github.com/AbelHaro/url-shortener/backend/internal/repository/idsranges"
	statisticRepo "github.com/AbelHaro/url-shortener/backend/internal/repository/statistic"
	urlRepo "github.com/AbelHaro/url-shortener/backend/internal/repository/url"
	counterSvc "github.com/AbelHaro/url-shortener/backend/internal/service/counter"
	idsrangesSvc "github.com/AbelHaro/url-shortener/backend/internal/service/idsranges"
	statisticSvc "github.com/AbelHaro/url-shortener/backend/internal/service/statistic"
	urlSvc "github.com/AbelHaro/url-shortener/backend/internal/service/url"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type failingRecordRepository struct {
	statisticRepo.Repository
}

func (repository failingRecordRepository) RecordClick(context.Context, *domain.URLStatistics) error {
	return errors.New("statistics unavailable")
}

func newTestHandler(t *testing.T, statisticsRepository statisticRepo.Repository) (*Handler, *urlSvc.Service, *statisticSvc.Service) {
	t.Helper()

	rangeService := idsrangesSvc.NewService(idsrangesRepo.NewMockRepository())
	counterService, err := counterSvc.NewService(rangeService)
	if err != nil {
		t.Fatalf("NewService(counter) error = %v", err)
	}
	urlService := urlSvc.NewService(urlRepo.NewMockRepository(), nil, counterService)
	statisticsService := statisticSvc.NewService(statisticsRepository)
	return NewHandler(statisticsService, urlService), urlService, statisticsService
}

func TestHandlerResolve(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name              string
		shortCode         string
		failStatistics    bool
		wantStatus        int
		wantRecordedCount int64
	}{
		{name: "records click", wantStatus: http.StatusOK, wantRecordedCount: 1},
		{name: "statistics failure is non-blocking", failStatistics: true, wantStatus: http.StatusOK},
		{name: "unknown short code", shortCode: "missing", wantStatus: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memoryRepository := statisticRepo.NewMockRepository()
			var repository statisticRepo.Repository = memoryRepository
			if tt.failStatistics {
				repository = failingRecordRepository{Repository: memoryRepository}
			}
			handler, urlService, _ := newTestHandler(t, repository)
			storedURL, err := urlService.Store("https://example.com", uuid.New())
			if err != nil {
				t.Fatalf("Store() error = %v", err)
			}
			shortCode := tt.shortCode
			if shortCode == "" {
				shortCode = storedURL.ShortCode
			}

			router := gin.New()
			router.POST("/urls/short/:shortCode/resolve", handler.Resolve)
			request := httptest.NewRequest(http.MethodPost, "/urls/short/"+shortCode+"/resolve", bytes.NewBufferString(`{"referrer":"https://news.example/article"}`))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Errorf("Resolve() status = %d, want %d; body = %s", response.Code, tt.wantStatus, response.Body)
			}
			if tt.wantRecordedCount > 0 {
				count, err := memoryRepository.GetClickCount(context.Background(), storedURL.ID)
				if err != nil {
					t.Fatalf("GetClickCount() error = %v", err)
				}
				if count != tt.wantRecordedCount {
					t.Errorf("click count = %d, want %d", count, tt.wantRecordedCount)
				}
			}
		})
	}
}

func TestHandlerGetDashboardAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	statisticsRepository := statisticRepo.NewMockRepository()
	handler, urlService, _ := newTestHandler(t, statisticsRepository)
	ownerID := uuid.New()
	storedURL, err := urlService.Store("https://example.com", ownerID)
	if err != nil {
		t.Fatalf("Store() error = %v", err)
	}

	tests := []struct {
		name       string
		userID     any
		wantStatus int
	}{
		{name: "owner", userID: ownerID, wantStatus: http.StatusOK},
		{name: "different user", userID: uuid.New(), wantStatus: http.StatusNotFound},
		{name: "unauthenticated", wantStatus: http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/urls/:id/statistics", func(c *gin.Context) {
				if tt.userID != nil {
					c.Set("userID", tt.userID)
				}
				c.Next()
			}, handler.GetDashboard)

			request := httptest.NewRequest(http.MethodGet, "/urls/"+storedURL.ID.String()+"/statistics", nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Errorf("GetDashboard() status = %d, want %d; body = %s", response.Code, tt.wantStatus, response.Body)
			}
		})
	}
}
