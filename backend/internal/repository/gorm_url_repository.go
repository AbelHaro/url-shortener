package repository

import (
	"errors"
	"fmt"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"gorm.io/gorm"
)

type GormURLRepository struct {
	db *gorm.DB
}

func NewGormURLRepository(db *gorm.DB) URLRepository {
	return &GormURLRepository{db: db}
}

func (repo *GormURLRepository) Store(url *domain.URL) error {
	result := repo.db.Create(url)
	if result.Error != nil {
		return fmt.Errorf("failed to store url: %w", result.Error)
	}
	return nil
}

func (repo *GormURLRepository) FindByOriginalURL(originalURL string) (*domain.URL, error) {
	var url domain.URL
	result := repo.db.Where("original_url = ?", originalURL).First(&url)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find url by original: %w", result.Error)
	}
	return &url, nil
}

func (repo *GormURLRepository) FindByShortURL(shortURL string) (*domain.URL, error) {
	var url domain.URL
	result := repo.db.Where("short_url = ?", shortURL).First(&url)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find url by short: %w", result.Error)
	}
	return &url, nil
}

func (repo *GormURLRepository) FindByID(id string) (*domain.URL, error) {
	var url domain.URL
	result := repo.db.First(&url, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find url by id: %w", result.Error)
	}
	return &url, nil
}

func (repo *GormURLRepository) DeleteByOriginalURL(originalURL string) error {
	result := repo.db.Where("original_url = ?", originalURL).Delete(&domain.URL{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete url: %w", result.Error)
	}
	return nil
}

func (repo *GormURLRepository) DeleteByShortURL(shortURL string) error {
	result := repo.db.Where("short_url = ?", shortURL).Delete(&domain.URL{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete url: %w", result.Error)
	}
	return nil
}

func (repo *GormURLRepository) DeleteByID(id string) error {
	result := repo.db.Delete(&domain.URL{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete url: %w", result.Error)
	}
	return nil
}
