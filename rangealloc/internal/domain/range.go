package domain

import "github.com/google/uuid"

type Range struct {
	ID      uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	Start   uint64    `json:"start" gorm:"not null"` // The start is inclusive, so the first ID allocated will be Start ended in 0
	End     uint64    `json:"end" gorm:"not null"`   // The end is not inclusive, so the last ID allocated will be End - 1. This allows us to allocate the next range starting from End
	Offset  uint64    `json:"offset" gorm:"not null"`
	OwnerID uuid.UUID `json:"owner_id" gorm:"type:uuid;not null"`
}
