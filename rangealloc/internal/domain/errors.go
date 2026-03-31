package domain

import "errors"

var (
	ErrRangeAllocFailed = errors.New("failed to allocate range")
	ErrRangeNotFound    = errors.New("range not found")
	ErrInvalidRange     = errors.New("invalid range")
	InternalError       = errors.New("internal error")
)
