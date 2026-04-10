package domain

import "github.com/google/uuid"

type Range struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	Start         uint64    `json:"start" gorm:"not null"` // The start is inclusive, so the first ID allocated will be Start ended in 0
	Last          uint64    `json:"last" gorm:"not null"`  // The end is not inclusive, so the last ID allocated will be End - 1. This allows us to allocate the next range starting from End
	CurrentOffset uint64    `json:"current_offset" gorm:"not null"`
}
