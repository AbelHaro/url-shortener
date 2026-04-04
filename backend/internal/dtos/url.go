package dtos

import (
	"time"

	"github.com/google/uuid"
)

// V1CreateShortenRequest is the request to create a shortened URL
type V1CreateShortenRequest struct {
	OriginalUrl string `json:"original_url" binding:"required"`
}

// V1URLResponse is the response containing URL details
type V1URLResponse struct {
	ID          uuid.UUID `json:"id"`
	OriginalURL string    `json:"original_url"`
	ShortCode   string    `json:"short_code"`
	UserID      uuid.UUID `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// V1SearchByOriginalURLRequest is the request to search for a URL by original URL
type V1SearchByOriginalURLRequest struct {
	URL string `json:"url" binding:"required"`
}
