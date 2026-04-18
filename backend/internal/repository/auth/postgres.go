package auth

import (
	"context"
	"time"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var _ Repository = (*PostgresRepository)(nil)

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) Repository {
	return &PostgresRepository{db: db}
}

func (repo *PostgresRepository) CreateUser(user *domain.User) error {
	ctx := context.Background()

	err := gorm.G[domain.User](repo.db).Create(ctx, user)
	if err != nil {
		return domain.ErrInternal
	}
	return nil
}

func (repo *PostgresRepository) FindByEmail(email string) (*domain.User, error) {
	ctx := context.Background()

	var user domain.User
	result := repo.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, domain.ErrUserNotFound
	}

	return &user, nil
}

func (repo *PostgresRepository) FindByID(id string) (*domain.User, error) {
	ctx := context.Background()

	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	var user domain.User
	result := repo.db.WithContext(ctx).Where("id = ?", userID).First(&user)
	if result.Error != nil {
		return nil, domain.ErrUserNotFound
	}

	return &user, nil
}

func (repo *PostgresRepository) StoreRefreshToken(userID, token string) error {
	ctx := context.Background()

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

	result := repo.db.WithContext(ctx).Create(refreshToken)
	if result.Error != nil {
		return domain.ErrInternal
	}

	return nil
}

func (repo *PostgresRepository) ValidateRefreshToken(token string) error {
	ctx := context.Background()

	var refreshToken domain.RefreshToken
	result := repo.db.WithContext(ctx).Where("token = ?", token).First(&refreshToken)
	if result.Error != nil {
		return domain.ErrInvalidToken
	}

	if time.Now().After(refreshToken.ValidUntil) {
		return domain.ErrTokenExpired
	}

	return nil
}

func (repo *PostgresRepository) InvalidateRefreshToken(token string) error {
	ctx := context.Background()

	result := repo.db.WithContext(ctx).Where("token = ?", token).Delete(&domain.RefreshToken{})
	if result.Error != nil {
		return domain.ErrInternal
	}

	return nil
}

func (repo *PostgresRepository) DeleteUser(userID string) error {
	ctx := context.Background()

	userIDInUUID, err := uuid.Parse(userID)
	if err != nil {
		return domain.ErrInvalidCredentials
	}

	result := repo.db.WithContext(ctx).Where("id = ?", userIDInUUID).Delete(&domain.User{})
	if result.Error != nil {
		return domain.ErrInternal
	}

	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (repo *PostgresRepository) Logout(userID string) error {
	ctx := context.Background()

	userIDInUUID, err := uuid.Parse(userID)
	if err != nil {
		return domain.ErrInvalidCredentials
	}

	result := repo.db.WithContext(ctx).Where("user_id = ?", userIDInUUID).Delete(&domain.RefreshToken{})
	if result.Error != nil {
		return domain.ErrInternal
	}

	return nil
}
