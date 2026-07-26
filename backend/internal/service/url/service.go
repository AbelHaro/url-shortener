package url

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	urlRepo "github.com/AbelHaro/url-shortener/backend/internal/repository/url"
	"github.com/AbelHaro/url-shortener/backend/internal/service/counter"
	"github.com/google/uuid"
)

type Service struct {
	repo           urlRepo.Repository
	counterService *counter.Service
}

func NewService(repo urlRepo.Repository, counterService *counter.Service) *Service {
	return &Service{
		repo:           repo,
		counterService: counterService,
	}
}

func (svc *Service) Store(originalURL string, ownerID uuid.UUID) (*domain.URL, error) {
	if err := svc.ValidateURL(originalURL); err != nil {
		return nil, err
	}

	// Note: We don't check for existing URL here to avoid race conditions.
	// Beacause the short code is generated based on a counter so it is the responsibility of the counter service to ensure that the same short code is not generated twice.

	shortCode, err := svc.counterService.GenerateShortHash()
	if err != nil {
		return nil, domain.ErrInternal
	}

	urlToInsert := &domain.URL{
		OriginalURL: originalURL,
		ShortCode:   shortCode,
		UserID:      ownerID,
	}

	urlInserted, err := svc.repo.Store(urlToInsert)

	if err != nil {
		return nil, domain.ErrInternal
	}
	return urlInserted, nil
}

func (svc *Service) FindByShortCode(shortCode string) (*domain.URL, error) {
	urlFound, err := svc.repo.FindByShortCode(shortCode)
	if err != nil {
		return nil, err
	}
	if urlFound == nil {
		return nil, domain.ErrURLNotFound
	}
	return urlFound, nil
}

func (svc *Service) FindByID(id string) (*domain.URL, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("parse URL ID: %w", domain.ErrInvalidID)
	}
	urlFound, err := svc.repo.FindByID(parsedID)
	if err != nil {
		return nil, err
	}
	if urlFound == nil {
		return nil, domain.ErrURLNotFound
	}
	return urlFound, nil
}

// FindByIDForUser returns a URL only when it belongs to the supplied user.
func (svc *Service) FindByIDForUser(ctx context.Context, id string, userID uuid.UUID) (*domain.URL, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("parse URL ID: %w", domain.ErrInvalidID)
	}
	urlFound, err := svc.repo.FindByIDAndUserID(ctx, parsedID, userID)
	if err != nil {
		return nil, fmt.Errorf("find URL for user: %w", err)
	}
	if urlFound == nil {
		return nil, domain.ErrURLNotFound
	}
	return urlFound, nil
}

// UpdateOriginalURLForUser changes a URL destination without changing its short code.
func (svc *Service) UpdateOriginalURLForUser(ctx context.Context, id string, userID uuid.UUID, originalURL string) (*domain.URL, error) {
	if err := svc.ValidateURL(originalURL); err != nil {
		return nil, err
	}
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("parse URL ID: %w", domain.ErrInvalidID)
	}
	updatedURL, err := svc.repo.UpdateOriginalURL(ctx, parsedID, userID, originalURL)
	if err != nil {
		if errors.Is(err, domain.ErrURLNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("update URL for user: %w", domain.ErrInternal)
	}
	return updatedURL, nil
}

func (svc *Service) FindByOriginalURL(originalURL string) (*domain.URL, error) {
	urlFound, err := svc.repo.FindByOriginalURL(originalURL)
	if err != nil {
		return nil, domain.ErrInternal
	}
	if urlFound == nil {
		return nil, domain.ErrURLNotFound
	}
	return urlFound, nil
}

func (svc *Service) DeleteByID(id string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("parse URL ID: %w", domain.ErrInvalidID)
	}
	_, err = svc.repo.FindByID(parsedID)
	if err != nil {
		return domain.ErrURLNotFound
	}

	err = svc.repo.DeleteByID(uuid.MustParse(id))
	if errors.Is(err, domain.ErrURLNotFound) {
		return err
	}
	if err != nil {
		return domain.ErrInternal
	}
	return nil
}

// DeleteByIDForUser deletes a URL only when it belongs to the supplied user.
func (svc *Service) DeleteByIDForUser(ctx context.Context, id string, userID uuid.UUID) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("parse URL ID: %w", domain.ErrInvalidID)
	}
	if err := svc.repo.DeleteByIDAndUserID(ctx, parsedID, userID); err != nil {
		if errors.Is(err, domain.ErrURLNotFound) {
			return err
		}
		return fmt.Errorf("delete URL for user: %w", domain.ErrInternal)
	}
	return nil
}

func (svc *Service) DeleteByOriginalURL(originalURL string) error {
	_, err := svc.repo.FindByOriginalURL(originalURL)
	if err != nil {
		return domain.ErrURLNotFound
	}

	err = svc.repo.DeleteByOriginalURL(originalURL)
	if err != nil {
		return domain.ErrInternal
	}
	return nil
}

func (svc *Service) DeleteByShortCode(shortCode string) error {
	_, err := svc.repo.FindByShortCode(shortCode)
	if err != nil {
		return domain.ErrURLNotFound
	}

	err = svc.repo.DeleteByShortCode(shortCode)
	if err != nil {
		return domain.ErrInternal
	}
	return nil
}

func (svc *Service) ValidateURL(rawURL string) error {
	_, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return domain.ErrInvalidURL
	}
	return nil
}

func (svc *Service) FindAllByUserID(userID uuid.UUID) ([]domain.URL, error) {
	urlsFound, err := svc.repo.FindAllByUserID(userID)
	if err != nil {
		return nil, domain.ErrInternal
	}
	return urlsFound, nil
}

// func (svc *Service) GenerateDevData() error {
// 	urls := []string{
// 		"https://google.com",
// 		"https://github.com",
// 		"https://stackoverflow.com",
// 		"https://golang.org",
// 		"https://gin-gonic.com",
// 	}

// 	for _, u := range urls {
// 		log.Println("Storing url", u)
// 		_, err := svc.Store(u)
// 		if err != nil {
// 			return err
// 		}
// 		log.Println("Stored url", u)
// 	}

// 	return nil
// }
