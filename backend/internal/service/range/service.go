package rangeservice

import (
	"fmt"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	rangerepository "github.com/AbelHaro/url-shortener/backend/internal/repository/range"
	"github.com/google/uuid"
)

type Service struct {
	repo rangerepository.Repository
}

func NewService(repo rangerepository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) AllocateRange(ownerID uuid.UUID) (*domain.Range, error) {
	rangeFound, err := s.repo.GetActiveRange(ownerID)

	fmt.Printf("Error when getting active range for owner %s: %v\n", ownerID, err)

	if err != nil {
		return nil, err
	}
	if rangeFound != nil {
		fmt.Printf("Range was found for owner %s: %+v, so need to update offset\n", ownerID, rangeFound)

		//Update the current offset because if the client is recovering from a failure, it can be that some IDs in the range are already used, so we need to update the current offset to avoid duplicating IDs. For example, if the range is from 0 to 1000 and the offset is 200, it means that between 200 and 299 could be already used, so we need to update the offset to 300 to avoid duplicating IDs.
		err = s.repo.UpdateRangeOffset(rangeFound.ID, ownerID)
		if err != nil {
			return nil, err
		}
		return rangeFound, nil
	}

	return s.repo.AllocateRange(ownerID)
}

func (s *Service) UpdateRangeOffset(rangeID uuid.UUID, ownerID uuid.UUID) error {
	return s.repo.UpdateRangeOffset(rangeID, ownerID)
}
