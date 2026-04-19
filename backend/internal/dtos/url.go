package dtos

import (
	"time"

	"github.com/google/uuid"
)

// CreateShortenRequest is the request to create a shortened URL
// @name CreateShortenRequest
type CreateShortenRequest struct {
	OriginalUrl string `json:"original_url" binding:"required"`
}

// URLResponse is the response containing URL details
// @name URLResponse
type URLResponse struct {
	ID          uuid.UUID `json:"id" binding:"required"`
	OriginalURL string    `json:"original_url" binding:"required"`
	ShortCode   string    `json:"short_code" binding:"required"`
	UserID      uuid.UUID `json:"user_id" binding:"required"`
	CreatedAt   time.Time `json:"created_at" binding:"required"`
	UpdatedAt   time.Time `json:"updated_at" binding:"required"`
}

// SearchByOriginalURLRequest is the request to search for a URL by original URL
// @name SearchByOriginalURLRequest
type SearchByOriginalURLRequest struct {
	OriginalURL string `json:"original_url" binding:"required"`
}
