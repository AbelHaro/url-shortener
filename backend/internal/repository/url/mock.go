package url

import (
	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/google/uuid"
)

type MockRepository struct {
	urls map[string]*domain.URL
}

func NewMockRepository() Repository {
	return &MockRepository{urls: make(map[string]*domain.URL)}
}

func (m *MockRepository) Store(url *domain.URL) (*domain.URL, error) {
	for _, existing := range m.urls {
		if existing.OriginalURL == url.OriginalURL {
			return existing, nil
		}
	}
	if url.ID == uuid.Nil {
		url.ID = uuid.New()
	}
	m.urls[url.ShortURL] = url
	return url, nil
}
func (m *MockRepository) FindByOriginalURL(originalURL string) (*domain.URL, error) {
	for _, url := range m.urls {
		if url.OriginalURL == originalURL {
			return url, nil
		}
	}
	return nil, nil
}
func (m *MockRepository) FindByShortURL(shortURL string) (*domain.URL, error) {
	if url, ok := m.urls[shortURL]; ok {
		return url, nil
	}
	return nil, nil
}
func (m *MockRepository) FindByID(id uuid.UUID) (*domain.URL, error) {
	for _, url := range m.urls {
		if url.ID == id {
			return url, nil
		}
	}
	return nil, nil
}
func (m *MockRepository) DeleteByOriginalURL(originalURL string) error {
	for _, url := range m.urls {
		if url.OriginalURL == originalURL {
			delete(m.urls, url.ShortURL)
			return nil
		}
	}
	return domain.ErrURLNotFound
}
func (m *MockRepository) DeleteByShortURL(shortURL string) error {
	_, ok := m.urls[shortURL]
	if !ok {
		return domain.ErrURLNotFound
	}
	delete(m.urls, shortURL)
	return nil
}
func (m *MockRepository) DeleteByID(id uuid.UUID) error {
	for _, url := range m.urls {
		if url.ID == id {
			delete(m.urls, url.ShortURL)
			return nil
		}
	}
	return domain.ErrURLNotFound
}
