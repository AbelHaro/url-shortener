package domain

import "github.com/google/uuid"

type URL struct {
	ID          uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	OriginalURL string    `json:"original_url" gorm:"not null"`
	ShortURL    string    `json:"short_url" gorm:"not null;unique"`
}
