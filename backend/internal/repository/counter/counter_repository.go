package counter

import "github.com/AbelHaro/url-shortener/backend/internal/domain"

type CounterRepository interface {
	GetCounter() (*domain.Counter, error)
	UpdateCounter(counter int64) error
}
