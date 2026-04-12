package idsranges

import (
	"fmt"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/AbelHaro/url-shortener/backend/internal/repository/idsranges"
	"github.com/google/uuid"
)

type Service struct {
	repo idsranges.Repository
}

func NewService(repo idsranges.Repository) *Service {
	return &Service{repo: repo}
}

/*
AllocateRange allocates a new range of IDs for the service. It first checks if there is an active range that can be used, and if so, it updates the offset of that range to avoid duplicating IDs. If there is no active range, it allocates a new range. This method is designed to be idempotent, so if the client is recovering from a failure, it can call this method again without risking ID duplication.
*/
func (s *Service) AllocateRange() (*domain.IDsRange, error) {
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
