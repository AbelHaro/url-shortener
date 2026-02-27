package domain

import (
	"time"

	"github.com/google/uuid"
)

type URL struct {
	ID          uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OriginalURL string    `json:"original_url" gorm:"not null;uniqueIndex"`
	ShortURL    string    `json:"short_url" gorm:"not null;uniqueIndex:idx_short_url"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
