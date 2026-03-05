package auth

import (
	"time"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/google/uuid"
)

type MockRepository struct {
	users         map[string]*domain.User
	refreshTokens map[string]*domain.RefreshToken
}

func NewMockRepository() Repository {
	return &MockRepository{
		users:         make(map[string]*domain.User),
		refreshTokens: make(map[string]*domain.RefreshToken),
	}
}

func (m *MockRepository) CreateUser(user *domain.User) error {
	for _, existing := range m.users {
		if existing.Email == user.Email {
			return domain.ErrUserExists
		}
	}
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	m.users[user.ID.String()] = user
	return nil
}

func (m *MockRepository) Authenticate(email string, hashedPassword string) (*domain.RefreshToken, error) {
	for _, user := range m.users {
		if user.Email == email && user.Password == hashedPassword {
			refreshToken := &domain.RefreshToken{
				ID:         uuid.New(),
				Token:      uuid.NewString(),
				UserID:     user.ID,
				ValidUntil: time.Now().Add(7 * 24 * time.Hour),
			}
			m.refreshTokens[refreshToken.Token] = refreshToken
			return refreshToken, nil
		}
	}
	return nil, domain.ErrInvalidCredentials
}

func (m *MockRepository) ValidateToken(token string) error {
	rt, ok := m.refreshTokens[token]
	if !ok {
		return domain.ErrInvalidToken
	}
	if time.Now().After(rt.ValidUntil) {
		return domain.ErrInvalidToken
	}
	return nil
}

func (m *MockRepository) GetUserByToken(token string) (uuid.UUID, error) {
	rt, ok := m.refreshTokens[token]
	if !ok {
		return uuid.Nil, domain.ErrInvalidToken
	}
	if time.Now().After(rt.ValidUntil) {
		return uuid.Nil, domain.ErrInvalidToken
	}
	return rt.UserID, nil
}

func (m *MockRepository) Logout(userID string) error {
	for token, rt := range m.refreshTokens {
		if rt.UserID.String() == userID {
			delete(m.refreshTokens, token)
		}
	}
	return nil
}

func (m *MockRepository) DeleteUser(userID string) error {
	if _, ok := m.users[userID]; !ok {
		return domain.ErrUserNotFound
	}
	delete(m.users, userID)
	return nil
}
