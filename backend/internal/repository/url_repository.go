package repository

import "github.com/AbelHaro/url-shortener/backend/internal/domain"

type URLRepository interface {
	Store(url *domain.URL) error
	FindByOriginalURL(originalURL string) (*domain.URL, error)
	FindByShortURL(shortURL string) (*domain.URL, error)
	FindByID(id string) (*domain.URL, error)
	DeleteByOriginalURL(originalURL string) error
	DeleteByShortURL(shortURL string) error
	DeleteByID(id string) error
}
