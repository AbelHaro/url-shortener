package rangeservice

import (
	"github.com/AbelHaro/url-shortener/rangealloc/internal/domain"
	rangerepository "github.com/AbelHaro/url-shortener/rangealloc/internal/repository/range"
	"github.com/google/uuid"
)

type Service struct {
	repo rangerepository.Repository
}

func NewService(repo rangerepository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) AllocateRange(ownerID uuid.UUID) (*domain.Range, error) {
	return s.repo.AllocateRange(ownerID)
}

func (s *Service) UpdateRangeOffset(rangeID uuid.UUID, ownerID uuid.UUID, offset uint64) error {
	return s.repo.UpdateRangeOffset(rangeID, ownerID, offset)
}
