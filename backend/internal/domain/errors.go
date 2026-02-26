package domain

import "errors"

var (
	ErrURLNotFound = errors.New("url not found")
	ErrURLExists   = errors.New("url already exists")
	ErrInvalidURL  = errors.New("invalid url")
	ErrInternal    = errors.New("internal error")
)
