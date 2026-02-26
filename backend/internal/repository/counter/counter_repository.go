package counter

import "github.com/AbelHaro/url-shortener/backend/internal/domain"

type CounterRepository interface {
	GetCounter() (*domain.HashCounter, error)
	UpdateCounter(counter int64) error
}
