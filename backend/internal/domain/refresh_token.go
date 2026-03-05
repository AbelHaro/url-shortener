package domain

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID         uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Token      string    `json:"token" gorm:"not null;uniqueIndex"`
	UserID     uuid.UUID `json:"user_id" gorm:"not null;uniqueIndex:idx_user_id;foreignKey:UserID"`
	User       User      `json:"user" gorm:"foreignKey:UserID"`
	ValidUntil time.Time `json:"valid_until" gorm:"not null;default:now()"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
