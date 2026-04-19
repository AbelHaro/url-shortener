package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Email        string         `json:"email" gorm:"not null;uniqueIndex"`
	Name         string         `json:"name" gorm:"not null;uniqueIndex"`
	PasswordHash string         `json:"-" gorm:"not null"`
	Urls         []URL          `json:"urls" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Tokens       []RefreshToken `json:"tokens" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}
