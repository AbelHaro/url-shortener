package dtos

// RegisterRequest is the request to register a new user
// @name RegisterRequest
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// AnonymousRegisterRequest is the request to create an anonymous account
// @name AnonymousRegisterRequest
type AnonymousRegisterRequest struct{}

// LoginRequest is the request to login a user
// @name LoginRequest
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// RefreshTokenRequest is the request to refresh an access token
// @name RefreshTokenRequest
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// TokenResponse contains access and refresh tokens
// @name TokenResponse
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// UserResponse contains user information
// @name UserResponse
type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// AuthResponse contains the user and token pair after authentication
// @name AuthResponse
type AuthResponse struct {
	User   UserResponse `json:"user"`
	Tokens TokenResponse `json:"tokens"`
}

// SessionResponse contains the authenticated user
// @name SessionResponse
type SessionResponse struct {
	User UserResponse `json:"user"`
}

// LogoutRequest is the request to logout a user
// @name LogoutRequest
type LogoutRequest struct {
	// No fields needed for logout
}
