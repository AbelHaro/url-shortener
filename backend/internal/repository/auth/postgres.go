package auth

import (
	"context"
	"time"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

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

func (repo *PostgresRepository) Authenticate(email string, hashedPassword string) (*domain.RefreshToken, error) {
	ctx := context.Background()

	user, err := gorm.G[domain.User](repo.db).Where("email = ? AND password = ?", email, hashedPassword).First(ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, domain.ErrInternal
	}

	refreshToken := &domain.RefreshToken{
		ID:         uuid.New(),
		Token:      uuid.NewString(),
		UserID:     user.ID,
		ValidUntil: time.Now().Add(7 * 24 * time.Hour),
	}

	_, err = gorm.G[domain.RefreshToken](repo.db).Where("user_id = ?", user.ID).Update(ctx, "valid_until", time.Now())
	if err != nil {
		return nil, domain.ErrInternal
	}

	err = gorm.G[domain.RefreshToken](repo.db).Create(ctx, refreshToken)
	if err != nil {
		return nil, domain.ErrInternal
	}

	return refreshToken, nil
}

func (repo *PostgresRepository) ValidateToken(token string) error {
	ctx := context.Background()

	refreshToken, err := gorm.G[domain.RefreshToken](repo.db).Where("token = ?", token).First(ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return domain.ErrInvalidToken
		}
		return domain.ErrInternal
	}

	if time.Now().After(refreshToken.ValidUntil) {
		return domain.ErrInvalidToken
	}

	return nil
}

func (repo *PostgresRepository) GetUserByToken(token string) (uuid.UUID, error) {
	ctx := context.Background()

	refreshToken, err := gorm.G[domain.RefreshToken](repo.db).Where("token = ?", token).First(ctx)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return uuid.Nil, domain.ErrInvalidToken
		}
		return uuid.Nil, domain.ErrInternal
	}

	if time.Now().After(refreshToken.ValidUntil) {
		return uuid.Nil, domain.ErrInvalidToken
	}

	return refreshToken.UserID, nil
}

func (repo *PostgresRepository) DeleteUser(userID string) error {
	ctx := context.Background()

	userIDInUUID, err := uuid.Parse(userID)
	if err != nil {
		return domain.ErrInvalidCredentials
	}

	rowsAffected, err := gorm.G[domain.User](repo.db).Where("id = ?", userIDInUUID).Delete(ctx)
	if err != nil {
		return domain.ErrInternal
	}

	if rowsAffected == 0 {
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

	_, err = gorm.G[domain.RefreshToken](repo.db).Where("user_id = ?", userIDInUUID).Delete(ctx)
	if err != nil {
		return domain.ErrInternal
	}

	return nil
}
