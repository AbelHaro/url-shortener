package idsranges

import (
	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/google/uuid"
)

type Repository interface {
	AllocateRange() (*domain.IDsRange, error)
	UpdateRangeOffset(rangeID uuid.UUID) error
	GetNextRangeAvailable() (start uint64, err error)
	GetActiveRange() (*domain.IDsRange, error)
}
