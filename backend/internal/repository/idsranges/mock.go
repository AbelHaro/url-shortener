package idsranges

import (
	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/google/uuid"
)

type MockRepository struct {
	idsRanges []*domain.IDsRange
}

func NewMockRepository() Repository {
	return &MockRepository{
		idsRanges: []*domain.IDsRange{
			{ID: uuid.New(), Start: 0, Last: 1000, CurrentOffset: 0},
		},
	}
}

func (m *MockRepository) AllocateRange() (*domain.IDsRange, error) {
	lastRange := m.idsRanges[len(m.idsRanges)-1]
	if lastRange.Start+lastRange.CurrentOffset+RANGE_OFFSET >= lastRange.Last {
		newRange := &domain.IDsRange{
			ID:            uuid.New(),
			Start:         lastRange.Last,
			Last:          lastRange.Last + RANGE_SIZE,
			CurrentOffset: 0,
		}
		m.idsRanges = append(m.idsRanges, newRange)
		return newRange, nil
	}

	return lastRange, nil
}

func (m *MockRepository) UpdateRangeOffset(rangeID uuid.UUID) error {
	for _, r := range m.idsRanges {
		if r.ID == rangeID {
			r.CurrentOffset += RANGE_OFFSET
			return nil
		}
	}
	return domain.ErrRangeNotFound
}

func (m *MockRepository) GetNextRangeAvailable() (start uint64, err error) {
	lastRange := m.idsRanges[len(m.idsRanges)-1]
	return lastRange.Last, nil
}

func (m *MockRepository) GetActiveRange() (*domain.IDsRange, error) {
	lastRange := m.idsRanges[len(m.idsRanges)-1]
	if lastRange.CurrentOffset >= RANGE_SIZE {
		return nil, domain.ErrRangeNotFound
	}
	return lastRange, nil
}
