package auth

import (
	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/AbelHaro/url-shortener/backend/internal/repository/auth"
	"github.com/google/uuid"
)

type Service struct {
	repo auth.Repository
}

func NewService(repo auth.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Register(email, password string) (*domain.User, error) {
	user := &domain.User{
		ID:       uuid.New(),
		Email:    email,
		Password: password,
	}

	err := s.repo.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) Login(email, password string) (*domain.RefreshToken, error) {
	return s.repo.Authenticate(email, password)
}

func (s *Service) ValidateToken(token string) error {
	return s.repo.ValidateToken(token)
}

func (s *Service) GetUserByToken(token string) (uuid.UUID, error) {
	return s.repo.GetUserByToken(token)
}

func (s *Service) Logout(userID string) error {
	return s.repo.Logout(userID)
}

func (s *Service) DeleteUser(userID string) error {
	return s.repo.DeleteUser(userID)
}
