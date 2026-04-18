package idsranges

import (
	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/google/uuid"
)

type Repository interface {
	AllocateNewRange() (*domain.IDsRange, error)
	UpdateRangeOffset(rangeID uuid.UUID) error
	GetActiveRange() (*domain.IDsRange, error)
}
