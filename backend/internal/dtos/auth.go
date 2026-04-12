package dtos

// V1RegisterRequest is the request to register a new user
type V1RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// V1LoginRequest is the request to login a user
type V1LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// V1RefreshTokenRequest is the request to refresh an access token
type V1RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// V1TokenResponse contains access and refresh tokens
type V1TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// V1UserResponse contains user information
type V1UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// V1LogoutRequest is the request to logout a user
type V1LogoutRequest struct {
	// No fields needed for logout
}
