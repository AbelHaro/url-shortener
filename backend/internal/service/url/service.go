package url

import (
	"errors"
	"log"
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

func (svc *Service) Store(originalURL string) (*domain.URL, error) {
	if err := svc.ValidateURL(originalURL); err != nil {
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
		return nil, domain.ErrInternal
	}

	urlToInsert := &domain.URL{
		ID:          uuid.New(),
		OriginalURL: originalURL,
		ShortURL:    shortURL,
	}

	if err := svc.repo.Store(urlToInsert); err != nil {
		return nil, domain.ErrInternal
	}

	return urlToInsert, nil
}

func (svc *Service) FindByShortURL(shortURL string) (*domain.URL, error) {
	urlFound, err := svc.repo.FindByShortURL(shortURL)
	if err != nil {
		return nil, err
	}
	if urlFound == nil {
		return nil, domain.ErrURLNotFound
	}
	return urlFound, nil
}

func (svc *Service) FindByID(id string) (*domain.URL, error) {
	urlFound, err := svc.repo.FindByID(uuid.MustParse(id))
	if err != nil {
		return nil, err
	}
	if urlFound == nil {
		return nil, domain.ErrURLNotFound
	}
	return urlFound, nil
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
	_, err := svc.repo.FindByID(uuid.MustParse(id))
	if err != nil {
		return domain.ErrURLNotFound
	}

	err = svc.repo.DeleteByID(uuid.MustParse(id))
	if err != nil {
		return domain.ErrInternal
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

func (svc *Service) DeleteByShortURL(shortURL string) error {
	_, err := svc.repo.FindByShortURL(shortURL)
	if err != nil {
		return domain.ErrURLNotFound
	}

	err = svc.repo.DeleteByShortURL(shortURL)
	if err != nil {
		return domain.ErrInternal
	}
	return nil
}

func (svc *Service) ValidateURL(rawURL string) error {
	_, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return errors.New("invalid url format")
	}
	return nil
}

func (svc *Service) GenerateDevData() error {
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
