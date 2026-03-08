package jwt

import (
	"time"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Service struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewService(secret string, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (s *Service) GenerateAccessToken(userID uuid.UUID, email string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID.String(),
		"email": email,
		"type":  "access",
		"exp":   time.Now().Add(s.accessTTL).Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *Service) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID.String(),
		"type": "refresh",
		"exp":  time.Now().Add(s.refreshTTL).Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *Service) ValidateAccessToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, domain.ErrInvalidToken
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, domain.ErrInvalidToken
	}

	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "access" {
		return nil, domain.ErrInvalidToken
	}

	return claims, nil
}

func (s *Service) ValidateRefreshToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, domain.ErrInvalidToken
		}
		return s.secret, nil
	})

	if err != nil {
		return uuid.Nil, domain.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return uuid.Nil, domain.ErrInvalidToken
	}

	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "refresh" {
		return uuid.Nil, domain.ErrInvalidToken
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
