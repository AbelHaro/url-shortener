package service

import (
	"errors"
	"log"
	"net/url"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/AbelHaro/url-shortener/backend/internal/repository"
	"github.com/google/uuid"
)

type URLService struct {
	repo           repository.URLRepository
	counterService *CounterService
}

func NewURLService(repo repository.URLRepository, counterService *CounterService) *URLService {
	return &URLService{
		repo:           repo,
		counterService: counterService,
	}
}

func (svc *URLService) Store(originalURL string) (*domain.URL, error) {
	if err := svc.validateURL(originalURL); err != nil {
		return nil, err
	}

	existing, err := svc.repo.FindByOriginalURL(originalURL)
	if err != nil && errors.Is(err, domain.ErrInternal) {
		return nil, err
	}

	if existing != nil {
		return existing, nil
	}

	shortURL, err := svc.counterService.GenerateShortHash()
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
	urlFound, err := svc.repo.FindByID(uuid.MustParse(id))
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

func (svc *URLService) DeleteByID(id string) error {
	_, err := svc.repo.FindByID(uuid.MustParse(id))
	if err != nil {
		return domain.ErrURLNotFound
	}

	return svc.repo.DeleteByID(uuid.MustParse(id))
}

func (svc *URLService) DeleteByOriginalURL(originalURL string) error {
	_, err := svc.repo.FindByOriginalURL(originalURL)
	if err != nil {
		return domain.ErrURLNotFound
	}

	return svc.repo.DeleteByOriginalURL(originalURL)
}

func (svc *URLService) DeleteByShortURL(shortURL string) error {
	_, err := svc.repo.FindByShortURL(shortURL)
	if err != nil {
		return domain.ErrURLNotFound
	}

	return svc.repo.DeleteByShortURL(shortURL)
}

func (svc *URLService) validateURL(rawURL string) error {
	_, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return errors.New("invalid url format")
	}
	return nil
}

func (svc *URLService) GenerateDevData() error {
	urls := []string{
		"https://google.com",
		"https://github.com",
		"https://stackoverflow.com",
		"https://golang.org",
		"https://gin-gonic.com",
	}

	for _, u := range urls {
		log.Println("Storing url", u)
		_, err := svc.Store(u)
		if err != nil {
			return err
		}
		log.Println("Stored url", u)
	}

	return nil
}
