package domain

import (
	"time"

	"github.com/google/uuid"
)

type URLStatistics struct {
	ID        uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UrlID     uuid.UUID `json:"url_id" gorm:"not null;index"`
	Url       URL       `json:"url" gorm:"foreignKey:UrlID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	ClickedAt time.Time `json:"clicked_at" gorm:"not null"`
	Referer   string    `json:"referer"`
	UserAgent string    `json:"user_agent"`
	Ip        string    `json:"ip"`
}
