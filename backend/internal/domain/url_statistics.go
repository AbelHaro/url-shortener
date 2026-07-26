package domain

import (
	"time"

	"github.com/google/uuid"
)

// URLStatistics is a privacy-preserving record of one shortened-link click.
type URLStatistics struct {
	ID        uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	URLID     uuid.UUID `json:"url_id" gorm:"not null;index:idx_url_statistics_url_clicked_at,priority:1"`
	URL       URL       `json:"-" gorm:"foreignKey:URLID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	ClickedAt time.Time `json:"clicked_at" gorm:"not null;index:idx_url_statistics_url_clicked_at,priority:2"`
	Referer   string    `json:"referer" gorm:"not null"`
}

// DailyClick contains the number of clicks recorded on a UTC calendar day.
type DailyClick struct {
	Date   time.Time `json:"date"`
	Clicks int64     `json:"clicks"`
}

// ReferrerCount contains an aggregated click count for a referrer hostname.
type ReferrerCount struct {
	Referrer string `json:"referrer"`
	Clicks   int64  `json:"clicks"`
}
