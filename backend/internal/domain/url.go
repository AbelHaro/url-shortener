package domain

import (
	"time"

	"github.com/google/uuid"
)

type URL struct {
	ID          uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OriginalURL string    `json:"original_url" gorm:"not null;uniqueIndex"`
	ShortCode   string    `json:"short_code" gorm:"not null;uniqueIndex:idx_short_code"`
	UserID      uuid.UUID `json:"user_id" gorm:"not null;uniqueIndex:idx_user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
