package idsranges

import (
	"context"
	"errors"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const RANGE_SIZE = 1000
const RANGE_OFFSET = 100

var _ Repository = (*PostgresRepository)(nil)

const rangeLockKey int64 = 9102026

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) Repository {
	return &PostgresRepository{db: db}
}

/*
AllocateRange allocates a new range of IDs for the service. When called during normal operation (counter exhaustion), it always allocates a new range. The service must ensure that if calls to AllocateRange, its done because the current range is exhausted and not because of a transient error, to avoid unnecessary range allocations.
*/
func (p *PostgresRepository) AllocateNewRange() (*domain.IDsRange, error) {
	ctx := context.Background()

	var rangeAllocated *domain.IDsRange

	err := p.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", rangeLockKey).Error; err != nil {
			return domain.ErrInternal
		}

		lastRange, err := gorm.G[domain.IDsRange](tx).Last(ctx)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// This is the first range being allocated so the start is 0
			rangeAllocated = &domain.IDsRange{
				ID:            uuid.New(),
				Start:         0,
				Last:          RANGE_SIZE,
				CurrentOffset: 0,
			}
		} else if err != nil {
			return domain.ErrInternal
		} else {
			// Allocate a new range starting from the last range's last value
			rangeAllocated = &domain.IDsRange{
				ID:            uuid.New(),
				Start:         lastRange.Last,
				Last:          lastRange.Last + RANGE_SIZE,
				CurrentOffset: 0,
			}
		}

		err = gorm.G[domain.IDsRange](tx).Create(ctx, rangeAllocated)
		if err != nil {
			return domain.ErrInternal
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return rangeAllocated, nil
}

/*
UpdateRangeOffset updates the current offset of the specified range by adding RANGE_OFFSET to it. It first checks if the range exists and if it has not been fully consumed. If the range is already consumed or if the new offset would exceed the last ID in the range, it returns an domain.ErrRangeConsumed or domain.ErrInvalidRange error, respectively. Otherwise, it updates the offset in the database. This method is designed to be idempotent, so if the client is recovering from a failure, it can call this method again without risking ID duplication, as long as the range has not been fully consumed.
Returns:
@param rangeID: The UUID of the range to update.
@return error: {domain.ErrRangeNotFound | domain.ErrRangeConsumed | domain.ErrRangeInvalid | domain.ErrInternal}
*/
func (p *PostgresRepository) UpdateRangeOffset(rangeID uuid.UUID) error {
	ctx := context.Background()

	err := p.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", rangeLockKey).Error; err != nil {
			return domain.ErrInternal
		}

		rangeToBeUpdated, err := gorm.G[domain.IDsRange](tx).Where("id = ?", rangeID).First(ctx)
		if err != nil {
			return domain.ErrRangeNotFound
		}

		if rangeToBeUpdated.Start+rangeToBeUpdated.CurrentOffset >= rangeToBeUpdated.Last {
			return domain.ErrRangeConsumed
		}

		rangeToBeUpdated.CurrentOffset += RANGE_OFFSET

		if rangeToBeUpdated.Start+rangeToBeUpdated.CurrentOffset > rangeToBeUpdated.Last {
			return domain.ErrRangeInvalid
		}

		rowsAffected, err := gorm.G[domain.IDsRange](tx).Where("id = ?", rangeID).Update(ctx, "current_offset", rangeToBeUpdated.CurrentOffset)
		if err != nil || rowsAffected == 0 {
			return domain.ErrRangeNotFound
		}
		return nil
	})
	return err
}

func (p *PostgresRepository) GetActiveRange() (*domain.IDsRange, error) {
	ctx := context.Background()

	activeRange, err := gorm.G[domain.IDsRange](p.db).Order("start DESC").First(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, domain.ErrInternal
	}

	return &activeRange, nil
}
