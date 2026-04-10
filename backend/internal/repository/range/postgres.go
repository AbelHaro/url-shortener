package rangerepository

import (
	"context"
	"errors"
	"fmt"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const RANGE_SIZE = 1000
const RANGE_OFFSET = 100

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) Repository {
	return &PostgresRepository{db: db}
}
func (p *PostgresRepository) AllocateRange() (*domain.Range, error) {
	ctx := context.Background()

	var rangeAllocated *domain.Range

	err := p.db.Transaction(func(tx *gorm.DB) error {

		lastRange, err := p.GetNextRangeAvailable()

		// This case happens when there are no ranges allocated yet, so we start from 0
		if errors.Is(err, domain.ErrRangeNotFound) {
			lastRange = 0
		} else if err != nil {
			return err
		}

		rangeToAllocate := &domain.Range{
			ID:            uuid.New(),
			Start:         lastRange,
			Last:          lastRange + RANGE_SIZE, // The end is not inclusive, so we can allocate the next range starting from lastRange + RANGE_SIZE
			CurrentOffset: 0,
		}

		err = gorm.G[domain.Range](p.db).Create(ctx, rangeToAllocate)

		if err != nil {
			return domain.ErrRangeAllocFailed
		}

		record, err := gorm.G[domain.Range](p.db).Where("id = ?", rangeToAllocate.ID).First(ctx)

		if err != nil {
			return domain.ErrRangeNotFound
		}

		rangeAllocated = &record
		return nil
	})

	if err != nil {
		return nil, err
	}

	return rangeAllocated, nil
}

func (p *PostgresRepository) UpdateRangeOffset(rangeID uuid.UUID) error {
	ctx := context.Background()

	err := p.db.Transaction(func(tx *gorm.DB) error {

		rangeToBeUpdated, err := gorm.G[domain.Range](p.db).Where("id = ?", rangeID).First(ctx)
		if err != nil {
			return domain.ErrRangeNotFound
		}

		fmt.Printf("Range to be updated: %+v\n", rangeToBeUpdated)

		if rangeToBeUpdated.Start+rangeToBeUpdated.CurrentOffset >= rangeToBeUpdated.Last {
			return domain.ErrRangeConsumed
		}

		// I the start is 2000, last is 3000, and the current offset is 900, it means that we are announcing that the ids between 2900 and 2999 are already used, so we need to update the offset to 3000 to avoid duplicating IDs. In this case, the range is already consumed, so we need to allocate a new range.
		if rangeToBeUpdated.Start+rangeToBeUpdated.CurrentOffset+RANGE_OFFSET > rangeToBeUpdated.Last {
			return domain.ErrInvalidRange
		}

		rowsAffected, err := gorm.G[domain.Range](p.db).Where("id = ?", rangeID).Update(ctx, "current_offset", rangeToBeUpdated.CurrentOffset+RANGE_OFFSET)
		if err != nil {
			return domain.ErrRangeNotFound
		}
		if rowsAffected == 0 {
			return domain.ErrRangeNotFound
		}
		return nil
	})

	return err
}

func (p *PostgresRepository) GetNextRangeAvailable() (lastRange uint64, err error) {
	ctx := context.Background()

	rangeRecord, err := gorm.G[domain.Range](p.db).Last(ctx)
	if err != nil {
		fmt.Printf("Error fetching last range: %v", err)
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, domain.ErrInternal
	}

	return rangeRecord.Last, nil
}

func (p *PostgresRepository) GetActiveRange() (*domain.Range, error) {
	ctx := context.Background()

	activeRange, err := gorm.G[domain.Range](p.db).Where(`start + current_offset < last`).Take(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, domain.ErrInternal
	}

	return &activeRange, nil
}
