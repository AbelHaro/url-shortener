package repository

import (
	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/google/uuid"
)

type MockURLRepository struct {
	urls map[string]*domain.URL
}

func NewMockURLRepository() *MockURLRepository {
	return &MockURLRepository{urls: make(map[string]*domain.URL)}
}

func (m *MockURLRepository) Store(url *domain.URL) error {
	m.urls[url.ShortURL] = url
	return nil
}
func (m *MockURLRepository) FindByOriginalURL(originalURL string) (*domain.URL, error) {
	for _, url := range m.urls {
		if url.OriginalURL == originalURL {
			return url, nil
		}
	}
	return nil, nil
}
func (m *MockURLRepository) FindByShortURL(shortURL string) (*domain.URL, error) {
	if url, ok := m.urls[shortURL]; ok {
		return url, nil
	}
	return nil, nil
}
func (m *MockURLRepository) FindByID(id uuid.UUID) (*domain.URL, error) {
	for _, url := range m.urls {
		if url.ID == id {
			return url, nil
		}
	}
	return nil, nil
}
func (m *MockURLRepository) DeleteByOriginalURL(originalURL string) error {
	for _, url := range m.urls {
		if url.OriginalURL == originalURL {
			delete(m.urls, url.ShortURL)
			return nil
		}
	}
	return domain.ErrURLNotFound
}
func (m *MockURLRepository) DeleteByShortURL(shortURL string) error {
	_, ok := m.urls[shortURL]
	if !ok {
		return domain.ErrURLNotFound
	}
	delete(m.urls, shortURL)
	return nil
}
func (m *MockURLRepository) DeleteByID(id uuid.UUID) error {
	for _, url := range m.urls {
		if url.ID == id {
			delete(m.urls, url.ShortURL)
			return nil
		}
	}
	return domain.ErrURLNotFound
}
