package rangerepository

import (
	"github.com/AbelHaro/url-shortener/rangealloc/internal/domain"
	"github.com/google/uuid"
)

type Repository interface {
	AllocateRange(ownerID uuid.UUID) (*domain.Range, error)
	UpdateRangeOffset(rangeID uuid.UUID, ownerID uuid.UUID, offset uint64) error
	GetNextRangeAvailable() (start uint64, err error)
}
