package idsranges

import (
	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/AbelHaro/url-shortener/backend/internal/repository/idsranges"
	"github.com/google/uuid"
)

type Service struct {
	repo idsranges.Repository
}

// NewService creates a new IDs range service.
func NewService(repo idsranges.Repository) *Service {
	return &Service{repo: repo}
}

// AllocateRange resolves the range to use at startup.
// If no active range exists, it allocates the first one.
// If an active range exists and still has capacity, it advances the offset to resume safely.
// If the active range is exhausted, it allocates the next range.
func (s *Service) AllocateRange() (*domain.IDsRange, error) {
	activeRange, err := s.repo.GetActiveRange()
	if err != nil {
		return nil, err
	}

	if activeRange == nil {
		return s.repo.AllocateNewRange()
	}

	if activeRange.Start+activeRange.CurrentOffset >= activeRange.Last {
		return s.repo.AllocateNewRange()
	}

	if err := s.repo.UpdateRangeOffset(activeRange.ID); err != nil {
		return nil, err
	}

	return s.repo.GetActiveRange()
}

// AllocateNewRange allocates the next range from the repository.
func (s *Service) AllocateNewRange() (*domain.IDsRange, error) {
	return s.repo.AllocateNewRange()
}

// UpdateRangeOffset advances the persisted offset for the given range.
func (s *Service) UpdateRangeOffset(rangeID uuid.UUID) error {
	return s.repo.UpdateRangeOffset(rangeID)
}

// GetActiveRange returns the current active range, if one exists.
func (s *Service) GetActiveRange() (*domain.IDsRange, error) {
	return s.repo.GetActiveRange()
}
