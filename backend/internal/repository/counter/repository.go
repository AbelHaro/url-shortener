package counter

import "github.com/AbelHaro/url-shortener/backend/internal/domain"

type Repository interface {
	GetCounter() (*domain.Counter, error)
	UpdateCounter(counter int64) error
}
