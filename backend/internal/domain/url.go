package domain

import (
	"time"

	"github.com/google/uuid"
)

type URL struct {
	ID          uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OriginalURL string    `json:"original_url" gorm:"not null"`
	ShortURL    string    `json:"short_url" gorm:"not null;unique"`
	CreatedAt   time.Time `json:"createdAt" gorm:"not null;default:now()"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"not null;default:now()"`
}
