package rangerepository

import (
	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/google/uuid"
)

type Repository interface {
	AllocateRange() (*domain.Range, error)
	UpdateRangeOffset(rangeID uuid.UUID) error
	GetNextRangeAvailable() (start uint64, err error)
	GetActiveRange() (*domain.Range, error)
}
