package repository

import (
	"errors"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
)

type InMemoryURLRepository struct {
	urls map[string]*domain.URL
}

func NewInMemoryURLRepository() URLRepository {
	return &InMemoryURLRepository{
		urls: make(map[string]*domain.URL),
	}
}

func (repo *InMemoryURLRepository) Store(url *domain.URL) error {
	repo.urls[url.ShortURL] = url
	return nil
}

func (repo *InMemoryURLRepository) FindByOriginalURL(originalURL string) (*domain.URL, error) {
	// TODO: Not implemented yet
	return nil, nil
}

func (repo *InMemoryURLRepository) FindByShortURL(shortURL string) (*domain.URL, error) {
	url, ok := repo.urls[shortURL]
	if !ok {
		return nil, errors.New("url not found")
	}
	return url, nil
}
func (repo *InMemoryURLRepository) FindByID(id string) (*domain.URL, error) {
	// TODO: Not implemented yet
	return nil, nil
}

func (repo *InMemoryURLRepository) DeleteByOriginalURL(originalURL string) error {
	// TODO: Not implemented yet
	return nil
}

func (repo *InMemoryURLRepository) DeleteByShortURL(shortURL string) error {
	repo.urls[shortURL] = nil
	return nil
}

func (repo *InMemoryURLRepository) DeleteByID(id string) error {
	// TODO: Not implemented yet
	return nil
}
