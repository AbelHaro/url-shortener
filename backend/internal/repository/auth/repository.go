package auth

import (
	"github.com/AbelHaro/url-shortener/backend/internal/domain"
)

type Repository interface {
	CreateUser(user *domain.User) error
	FindByEmail(email string) (*domain.User, error)
	FindByID(id string) (*domain.User, error)
	StoreRefreshToken(userID, token string) error
	ValidateRefreshToken(token string) error
	InvalidateRefreshToken(token string) error
	Logout(userID string) error
	DeleteUser(userID string) error
}
