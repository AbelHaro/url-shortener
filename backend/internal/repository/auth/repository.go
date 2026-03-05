package auth

import (
	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/google/uuid"
)

type Repository interface {
	CreateUser(user *domain.User) error
	Authenticate(email, hashedPassword string) (*domain.RefreshToken, error)
	ValidateToken(token string) error
	GetUserByToken(token string) (uuid.UUID, error)
	Logout(userID string) error
	DeleteUser(userID string) error
}
