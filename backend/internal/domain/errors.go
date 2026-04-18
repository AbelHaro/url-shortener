package domain

import "errors"

var (
	ErrURLNotFound = errors.New("url not found")
	ErrURLExists   = errors.New("url already exists")
	ErrInvalidURL  = errors.New("invalid url")
	ErrInternal    = errors.New("internal error")

	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")

	ErrUrlStatisticsNotFound = errors.New("url statistics not found")

	// Range allocation errors
	ErrRangeAllocFailed = errors.New("failed to allocate range")
	ErrRangeNotFound    = errors.New("range not found")
	ErrRangeInvalid     = errors.New("invalid range")
	ErrRangeConsumed    = errors.New("range consumed")
)
