package auth

import (
	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/AbelHaro/url-shortener/backend/internal/repository/auth"
	"github.com/AbelHaro/url-shortener/backend/internal/service/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo auth.Repository
	jwt  *jwt.Service
}

func NewService(repo auth.Repository, jwtSvc *jwt.Service) *Service {
	return &Service{
		repo: repo,
		jwt:  jwtSvc,
	}
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (s *Service) Register(email, password string) (*domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, domain.ErrInternal
	}

	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	err = s.repo.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) Login(email, password string) (*TokenPair, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	accessToken, err := s.jwt.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, domain.ErrInternal
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, domain.ErrInternal
	}

	err = s.repo.StoreRefreshToken(user.ID.String(), refreshToken)
	if err != nil {
		return nil, domain.ErrInternal
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *Service) RefreshToken(refreshToken string) (*TokenPair, error) {
	userID, err := s.jwt.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	err = s.repo.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.FindByID(userID.String())
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	accessToken, err := s.jwt.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, domain.ErrInternal
	}

	newRefreshToken, err := s.jwt.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, domain.ErrInternal
	}

	err = s.repo.InvalidateRefreshToken(refreshToken)
	if err != nil {
		return nil, domain.ErrInternal
	}

	err = s.repo.StoreRefreshToken(user.ID.String(), newRefreshToken)
	if err != nil {
		return nil, domain.ErrInternal
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *Service) ValidateAccessToken(token string) (uuid.UUID, error) {
	claims, err := s.jwt.ValidateAccessToken(token)
	if err != nil {
		return uuid.Nil, err
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, domain.ErrInvalidToken
	}

	userID, err := uuid.Parse(sub)
	if err != nil {
		return uuid.Nil, domain.ErrInvalidToken
	}

	return userID, nil
}

func (s *Service) Logout(userID string) error {
	return s.repo.Logout(userID)
}

func (s *Service) DeleteUser(userID string) error {
	return s.repo.DeleteUser(userID)
}
