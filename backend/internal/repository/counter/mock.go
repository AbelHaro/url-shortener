package counter

import (
	"time"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/google/uuid"
)

type MockRepository struct {
	counter domain.Counter
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		counter: domain.Counter{
			ID:        uuid.New(),
			Counter:   0,
			UpdatedAt: time.Now(),
		},
	}
}
func (m MockRepository) GetCounter() (*domain.Counter, error) {
	return &m.counter, nil
}

func (m MockRepository) UpdateCounter(counter int64) error {
	m.counter.Counter += counter
	return nil
}
