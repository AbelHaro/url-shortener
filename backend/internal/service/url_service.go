package service

import (
	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/AbelHaro/url-shortener/backend/internal/repository"
	"github.com/AbelHaro/url-shortener/backend/internal/utils"
	"github.com/google/uuid"
)

type URLService struct {
	repo repository.URLRepository
}

func NewURLService(repo repository.URLRepository) *URLService {
	return &URLService{
		repo: repo,
	}
}

func (svc *URLService) Store(originalURL string) (*domain.URL, error) {
	shortURL, err := utils.GenerateShortURL(originalURL)

	if err != nil {
		return nil, err
	}

	url := &domain.URL{
		ID:          uuid.New(),
		OriginalURL: originalURL,
		ShortURL:    shortURL,
	}
	err = svc.repo.Store(url)
	if err != nil {
		return nil, err
	}

	return url, nil
}

func (svc *URLService) FindByShortURL(shortURL string) (*domain.URL, error) {
	return svc.repo.FindByShortURL(shortURL)
}

func (svc *URLService) FindByID(id string) (*domain.URL, error) {
	return svc.repo.FindByID(id)
}

func (svc *URLService) FindByOriginalURL(originalURL string) (*domain.URL, error) {
	return svc.repo.FindByOriginalURL(originalURL)
}

func (svc *URLService) Delete(id string) error {
	return svc.repo.DeleteByID(id)
}
