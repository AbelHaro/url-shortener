package service

import (
	"errors"
	"net/url"

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
	if err := svc.validateURL(originalURL); err != nil {
		return nil, err
	}

	existing, err := svc.repo.FindByOriginalURL(originalURL)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	shortURL, err := utils.GenerateShortURL(originalURL)
	if err != nil {
		return nil, err
	}

	urlToInsert := &domain.URL{
		ID:          uuid.New(),
		OriginalURL: originalURL,
		ShortURL:    shortURL,
	}

	if err := svc.repo.Store(urlToInsert); err != nil {
		return nil, err
	}

	return urlToInsert, nil
}

func (svc *URLService) FindByShortURL(shortURL string) (*domain.URL, error) {
	urlFound, err := svc.repo.FindByShortURL(shortURL)
	if err != nil {
		return nil, err
	}
	if urlFound == nil {
		return nil, domain.ErrURLNotFound
	}
	return urlFound, nil
}

func (svc *URLService) FindByID(id string) (*domain.URL, error) {
	urlFound, err := svc.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if urlFound != nil {
		return nil, domain.ErrURLNotFound
	}
	return urlFound, nil
}

func (svc *URLService) FindByOriginalURL(originalURL string) (*domain.URL, error) {
	return svc.repo.FindByOriginalURL(originalURL)
}

func (svc *URLService) Delete(id string) error {
	_, err := svc.repo.FindByID(id)
	if err != nil {
		return domain.ErrURLNotFound
	}

	return svc.repo.DeleteByID(id)
}

func (svc *URLService) validateURL(rawURL string) error {
	_, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return errors.New("invalid url format")
	}
	return nil
}
