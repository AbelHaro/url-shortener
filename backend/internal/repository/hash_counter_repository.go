package repository

import "github.com/AbelHaro/url-shortener/backend/internal/domain"

type HashCounterRepository interface {
	GetCounter() (*domain.HashCounter, error)
	UpdateCounter(counter int64) error
}
