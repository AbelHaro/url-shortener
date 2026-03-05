package statistic

import (
	"time"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	statisticRepo "github.com/AbelHaro/url-shortener/backend/internal/repository/statistic"
	"github.com/google/uuid"
)

type Service struct {
	repo statisticRepo.Repository
}

func NewService(repo statisticRepo.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) RecordClick(urlID, referer, userAgent, ip string) error {
	stat := &domain.URLStatistics{
		ID:        uuid.New(),
		UrlID:     uuid.MustParse(urlID),
		ClickedAt: time.Now(),
		Referer:   referer,
		UserAgent: userAgent,
		Ip:        ip,
	}
	return s.repo.RecordClick(stat)
}

func (s *Service) GetStatistics(urlID string) ([]*domain.URLStatistics, error) {
	return s.repo.GetStatistics(urlID)
}

func (s *Service) GetClickCount(urlID string) (int64, error) {
	return s.repo.GetClickCount(urlID)
}

func (s *Service) GetLastAccessAt(urlID string) (time.Time, error) {
	return s.repo.GetLastAccessAt(urlID)
}
