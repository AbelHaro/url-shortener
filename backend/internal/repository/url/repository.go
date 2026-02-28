package url

import (
	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/google/uuid"
)

type Repository interface {
	Store(url *domain.URL) (*domain.URL, error)
	FindByOriginalURL(originalURL string) (*domain.URL, error)
	FindByShortCode(shortCode string) (*domain.URL, error)
	FindByID(id uuid.UUID) (*domain.URL, error)
	DeleteByOriginalURL(originalURL string) error
	DeleteByShortCode(shortCode string) error
	DeleteByID(id uuid.UUID) error
}
