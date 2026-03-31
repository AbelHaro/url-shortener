package rangerepository

import (
	"context"
	"errors"

	"github.com/AbelHaro/url-shortener/rangealloc/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const RANGE_SIZE = 1000

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) Repository {
	return &PostgresRepository{db: db}
}
func (p *PostgresRepository) AllocateRange(ownerID uuid.UUID) (*domain.Range, error) {
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
			ID:      uuid.New(),
			Start:   lastRange,
			End:     lastRange + RANGE_SIZE, // The end is not inclusive, so we can allocate the next range starting from lastRange + RANGE_SIZE
			OwnerID: ownerID,
			Offset:  0,
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

func (p *PostgresRepository) UpdateRangeOffset(rangeID uuid.UUID, ownerID uuid.UUID, offset uint64) error {
	ctx := context.Background()

	err := p.db.Transaction(func(tx *gorm.DB) error {

		rangeToBeUpdated, err := gorm.G[domain.Range](p.db).Where("id = ? AND owner_id = ?", rangeID, ownerID).First(ctx)
		if err != nil {
			return domain.ErrRangeNotFound
		}

		if rangeToBeUpdated.Start+rangeToBeUpdated.Offset+offset > rangeToBeUpdated.End {
			return domain.ErrInvalidRange
		}

		rowsAffected, err := gorm.G[domain.Range](p.db).Where("id = ?", rangeID).Update(ctx, "offset", rangeToBeUpdated.Offset+offset)
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

	rangeRecord, err := gorm.G[domain.Range](p.db).Order("end desc").First(ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, domain.InternalError
	}

	return rangeRecord.End, nil
}
