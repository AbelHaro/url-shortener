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

func (s *Service) AllocateRange() (*domain.Range, error) {
	rangeFound, err := s.repo.GetActiveRange()

	fmt.Printf("Error when getting active range: %v\n", err)

	if err != nil {
		return nil, err
	}
	if rangeFound != nil {
		fmt.Printf("Range was found: %+v, so need to update offset\n", rangeFound)

		//Update the current offset because if the client is recovering from a failure, it can be that some IDs in the range are already used, so we need to update the current offset to avoid duplicating IDs. For example, if the range is from 0 to 1000 and the offset is 200, it means that between 200 and 299 could be already used, so we need to update the offset to 300 to avoid duplicating IDs.
		err = s.repo.UpdateRangeOffset(rangeFound.ID)
		if err != nil {
			return nil, err
		}
		return rangeFound, nil
	}

	return s.repo.AllocateRange()
}

func (s *Service) UpdateRangeOffset(rangeID uuid.UUID) error {
	return s.repo.UpdateRangeOffset(rangeID)
}
