package dtos

// V1ErrorResponse is a shared error response format used across all APIs
type V1ErrorResponse struct {
	Error string `json:"error"`
}

// Common HTTP status codes (for documentation/reference)
const (
	StatusBadRequest          = 400
	StatusUnauthorized        = 401
	StatusForbidden           = 403
	StatusNotFound            = 404
	StatusConflict            = 409
	StatusInternalServerError = 500
)
