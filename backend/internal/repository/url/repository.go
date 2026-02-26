package url

import (
	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/google/uuid"
)

type Repository interface {
	Store(url *domain.URL) error
	FindByOriginalURL(originalURL string) (*domain.URL, error)
	FindByShortURL(shortURL string) (*domain.URL, error)
	FindByID(id uuid.UUID) (*domain.URL, error)
	DeleteByOriginalURL(originalURL string) error
	DeleteByShortURL(shortURL string) error
	DeleteByID(id uuid.UUID) error
}
