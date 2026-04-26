package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

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

func randomAccountName() string {
	parts := []string{"swift", "calm", "bright", "mellow", "kind", "wild", "quiet", "brisk"}
	animals := []string{"fox", "otter", "sparrow", "panda", "wolf", "heron", "lynx", "koala"}
	partIdx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(parts))))
	animalIdx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(animals))))
	return fmt.Sprintf("%s-%s-%d", parts[partIdx.Int64()], animals[animalIdx.Int64()], time.Now().UnixNano()%10000)
}

type AuthResult struct {
	User   *domain.User
	Tokens *TokenPair
}

func (s *Service) issueTokensForUser(user *domain.User) (*TokenPair, error) {
	accessToken, err := s.jwt.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, domain.ErrInternal
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, domain.ErrInternal
	}

	if err := s.repo.StoreRefreshToken(user.ID.String(), refreshToken); err != nil {
		return nil, domain.ErrInternal
	}

	return &TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *Service) IssueAccessToken(userID uuid.UUID, email string) (string, error) {
	accessToken, err := s.jwt.GenerateAccessToken(userID, email)
	if err != nil {
		return "", domain.ErrInternal
	}

	return accessToken, nil
}

func (s *Service) AccessTTL() time.Duration {
	return s.jwt.AccessTTL()
}

func (s *Service) RefreshTTL() time.Duration {
	return s.jwt.RefreshTTL()
}

func (s *Service) Register(email, password string) (*AuthResult, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, domain.ErrInternal
	}

	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		Name:         randomAccountName(),
		PasswordHash: string(hashedPassword),
	}

	err = s.repo.CreateUser(user)
	if err != nil {
		return nil, err
	}

	tokens, err := s.issueTokensForUser(user)
	if err != nil {
		return nil, err
	}

	return &AuthResult{User: user, Tokens: tokens}, nil
}

func (s *Service) RegisterAnonymous() (*AuthResult, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(uuid.NewString()), bcrypt.DefaultCost)
	if err != nil {
		return nil, domain.ErrInternal
	}

	user := &domain.User{
		ID:           uuid.New(),
		Email:        fmt.Sprintf("anon-%s@local", uuid.NewString()),
		Name:         randomAccountName(),
		PasswordHash: string(hashedPassword),
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}

	tokens, err := s.issueTokensForUser(user)
	if err != nil {
		return nil, err
	}

	return &AuthResult{User: user, Tokens: tokens}, nil
}

func (s *Service) Login(email, password string) (*AuthResult, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	tokens, err := s.issueTokensForUser(user)
	if err != nil {
		return nil, err
	}

	return &AuthResult{User: user, Tokens: tokens}, nil
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

func (s *Service) Session(userID string) (*domain.User, error) {
	return s.repo.FindByID(userID)
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

func (s *Service) ValidateAccessTokenClaims(token string) (map[string]any, error) {
	claims, err := s.jwt.ValidateAccessToken(token)
	if err != nil {
		return nil, err
	}

	result := make(map[string]any, len(claims))
	for k, v := range claims {
		result[k] = v
	}

	return result, nil
}

func (s *Service) Logout(userID string) error {
	return s.repo.Logout(userID)
}

func (s *Service) DeleteUser(userID string) error {
	return s.repo.DeleteUser(userID)
}

func (s *Service) UpdateRefreshTokenExpiration(refreshToken string) error {
	return s.repo.UpdateRefreshTokenExpiration(refreshToken)
}
