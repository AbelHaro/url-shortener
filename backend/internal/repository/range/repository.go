package rangerepository

import (
	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/google/uuid"
)

type Repository interface {
	AllocateRange(ownerID uuid.UUID) (*domain.Range, error)
	UpdateRangeOffset(rangeID uuid.UUID, ownerID uuid.UUID) error
	GetNextRangeAvailable() (start uint64, err error)
	GetActiveRange(ownerID uuid.UUID) (*domain.Range, error)
}
