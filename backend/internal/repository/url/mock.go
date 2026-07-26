package url

import (
	"context"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/google/uuid"
)

var _ Repository = (*MockRepository)(nil)

type MockRepository struct {
	urls map[string]*domain.URL
}

func NewMockRepository() Repository {
	return &MockRepository{urls: make(map[string]*domain.URL)}
}

func (m *MockRepository) Store(url *domain.URL) (*domain.URL, error) {
	if url.ID == uuid.Nil {
		url.ID = uuid.New()
	}
	m.urls[url.ShortCode] = url
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
func (m *MockRepository) UpdateOriginalURL(_ context.Context, id, userID uuid.UUID, originalURL string) (*domain.URL, error) {
	for _, storedURL := range m.urls {
		if storedURL.ID == id && storedURL.UserID == userID {
			storedURL.OriginalURL = originalURL
			return storedURL, nil
		}
	}
	return nil, domain.ErrURLNotFound
}
func (m *MockRepository) FindByShortCode(shortCode string) (*domain.URL, error) {
	if url, ok := m.urls[shortCode]; ok {
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
func (m *MockRepository) FindByIDAndUserID(_ context.Context, id, userID uuid.UUID) (*domain.URL, error) {
	for _, storedURL := range m.urls {
		if storedURL.ID == id && storedURL.UserID == userID {
			return storedURL, nil
		}
	}
	return nil, nil
}
func (m *MockRepository) DeleteByOriginalURL(originalURL string) error {
	for _, url := range m.urls {
		if url.OriginalURL == originalURL {
			delete(m.urls, url.ShortCode)
			return nil
		}
	}
	return domain.ErrURLNotFound
}
func (m *MockRepository) DeleteByShortCode(shortCode string) error {
	_, ok := m.urls[shortCode]
	if !ok {
		return domain.ErrURLNotFound
	}
	delete(m.urls, shortCode)
	return nil
}
func (m *MockRepository) DeleteByID(id uuid.UUID) error {
	for _, url := range m.urls {
		if url.ID == id {
			delete(m.urls, url.ShortCode)
			return nil
		}
	}
	return domain.ErrURLNotFound
}
func (m *MockRepository) DeleteByIDAndUserID(_ context.Context, id, userID uuid.UUID) error {
	for _, storedURL := range m.urls {
		if storedURL.ID == id && storedURL.UserID == userID {
			delete(m.urls, storedURL.ShortCode)
			return nil
		}
	}
	return domain.ErrURLNotFound
}

func (m *MockRepository) FindAllByUserID(userID uuid.UUID) ([]domain.URL, error) {
	var urls []domain.URL
	for _, url := range m.urls {
		if url.UserID == userID {
			urls = append(urls, *url)
		}
	}
	return urls, nil
}
