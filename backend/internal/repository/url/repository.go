package url

import (
	"context"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/google/uuid"
)

type Repository interface {
	Store(url *domain.URL) (*domain.URL, error)
	FindByOriginalURL(originalURL string) (*domain.URL, error)
	FindByShortCode(shortCode string) (*domain.URL, error)
	FindByID(id uuid.UUID) (*domain.URL, error)
	FindByIDAndUserID(ctx context.Context, id, userID uuid.UUID) (*domain.URL, error)
	UpdateOriginalURL(ctx context.Context, id, userID uuid.UUID, originalURL string) (*domain.URL, error)
	DeleteByOriginalURL(originalURL string) error
	DeleteByShortCode(shortCode string) error
	DeleteByID(id uuid.UUID) error
	DeleteByIDAndUserID(ctx context.Context, id, userID uuid.UUID) error
	FindAllByUserID(userID uuid.UUID) ([]domain.URL, error)
}
