package domain

import (
	"time"

	"github.com/google/uuid"
)

type Counter struct {
	ID        uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Counter   int64     `json:"counter" gorm:"not null;default:0;type:bigint"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"not null;default:now()"`
}
