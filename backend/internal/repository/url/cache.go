package url

import (
	"context"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
)

// Cache stores URL lookups by short code. Implementations must make deletes
// idempotent and return domain.ErrCacheMiss for absent keys.
type Cache interface {
	// GetByShortCode returns ErrCacheMiss when the key does not exist.
	GetByShortCode(ctx context.Context, shortCode string) (*domain.URL, error)
	SetByShortCode(ctx context.Context, shortCode string, url *domain.URL) error
	DeleteByShortCode(ctx context.Context, shortCode string) error
}
