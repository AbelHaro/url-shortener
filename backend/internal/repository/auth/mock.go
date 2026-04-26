package auth

import (
	"time"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/google/uuid"
)

var _ Repository = (*MockRepository)(nil)

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

func (m *MockRepository) UpdateUser(user *domain.User) error {
	m.users[user.ID.String()] = user
	return nil
}

func (m *MockRepository) FindByEmail(email string) (*domain.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (m *MockRepository) FindByID(id string) (*domain.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (m *MockRepository) StoreRefreshToken(userID, token string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return domain.ErrInternal
	}

	refreshToken := &domain.RefreshToken{
		ID:         uuid.New(),
		Token:      token,
		UserID:     userUUID,
		ValidUntil: time.Now().Add(7 * 24 * time.Hour),
	}
	m.refreshTokens[token] = refreshToken
	return nil
}

func (m *MockRepository) ValidateRefreshToken(token string) error {
	rt, ok := m.refreshTokens[token]
	if !ok {
		return domain.ErrInvalidToken
	}
	if time.Now().After(rt.ValidUntil) {
		return domain.ErrTokenExpired
	}
	return nil
}

func (m *MockRepository) InvalidateRefreshToken(token string) error {
	delete(m.refreshTokens, token)
	return nil
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

func (m *MockRepository) UpdateRefreshTokenExpiration(token string) error {
	rt, ok := m.refreshTokens[token]
	if !ok {
		return domain.ErrInvalidToken
	}

	rt.ValidUntil = time.Now().Add(7 * 24 * time.Hour)
	rt.UpdatedAt = time.Now()
	m.refreshTokens[token] = rt
	return nil
}
