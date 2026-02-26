package counter

import (
	"time"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/google/uuid"
)

type MockCounterRepository struct {
	counter domain.Counter
}

func NewMockCounterRepository() *MockCounterRepository {
	return &MockCounterRepository{
		counter: domain.Counter{
			ID:        uuid.New(),
			Counter:   0,
			UpdatedAt: time.Now(),
		},
	}
}
func (m MockCounterRepository) GetCounter() (*domain.Counter, error) {
	return &m.counter, nil
}

func (m MockCounterRepository) UpdateCounter(counter int64) error {
	m.counter.Counter += counter
	return nil
}
