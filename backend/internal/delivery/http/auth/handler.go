// Package auth provides HTTP handlers for authentication endpoints.
// @title           URL Shortener API
// @version         1.0
// @description     API for shortening URLs
// @host            localhost:8080
// @BasePath        /api/v1/auth
package auth

import (
	"errors"
	"net/http"

	authcookie "github.com/AbelHaro/url-shortener/backend/internal/delivery/http/authcookie"
	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/AbelHaro/url-shortener/backend/internal/dtos"
	"github.com/AbelHaro/url-shortener/backend/internal/service/auth"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service      *auth.Service
	secureCookie bool
}

func NewHandler(svc *auth.Service, secureCookie bool) *Handler {
	return &Handler{service: svc, secureCookie: secureCookie}
}

// Register Create a new user account
// @Summary Register a new user
// @Description Create a new user account with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dtos.RegisterRequest true "Registration details"
// @Success 201 {object} dtos.AuthResponse
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 409 {object} dtos.ErrorResponse
// @Router /auth/register [post]
// @ID postAuthRegister
func (h *Handler) Register(c *gin.Context) {
	var req dtos.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: "invalid request body"})
		return
	}

	authResult, err := h.service.Register(req.Email, req.Password)
	if err != nil {
		h.handleError(c, err)
		return
	}

	authcookie.SetAccessToken(c, authResult.Tokens.AccessToken, h.service.AccessTTL(), h.secureCookie)
	authcookie.SetRefreshToken(c, authResult.Tokens.RefreshToken, h.service.RefreshTTL(), h.secureCookie)
	c.JSON(http.StatusCreated, dtos.AuthResponse{
		User: dtos.UserResponse{
			ID:    authResult.User.ID.String(),
			Email: authResult.User.Email,
			Name:  authResult.User.Name,
		},
		Tokens: dtos.TokenResponse{
			AccessToken:  authResult.Tokens.AccessToken,
			RefreshToken: authResult.Tokens.RefreshToken,
		},
	})
}

// AnonymousRegister Create an anonymous account
// @Summary Register anonymous user
// @Description Create a new anonymous account with a random name
// @Tags Auth
// @Accept json
// @Produce json
// @Success 201 {object} dtos.AuthResponse
// @Failure 500 {object} dtos.ErrorResponse
// @Router /auth/anonymous [post]
// @ID postAuthAnonymousRegister
func (h *Handler) AnonymousRegister(c *gin.Context) {
	authResult, err := h.service.RegisterAnonymous()
	if err != nil {
		h.handleError(c, err)
		return
	}

	authcookie.SetAccessToken(c, authResult.Tokens.AccessToken, h.service.AccessTTL(), h.secureCookie)
	authcookie.SetRefreshToken(c, authResult.Tokens.RefreshToken, h.service.RefreshTTL(), h.secureCookie)
	c.JSON(http.StatusCreated, dtos.AuthResponse{
		User: dtos.UserResponse{
			ID:    authResult.User.ID.String(),
			Email: authResult.User.Email,
			Name:  authResult.User.Name,
		},
		Tokens: dtos.TokenResponse{
			AccessToken:  authResult.Tokens.AccessToken,
			RefreshToken: authResult.Tokens.RefreshToken,
		},
	})
}

// Login Authenticate user and get tokens
// @Summary Login user
// @Description Authenticate with email and password, returns access and refresh tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dtos.LoginRequest true "Login credentials"
// @Success 200 {object} dtos.AuthResponse
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Router /auth/login [post]
// @ID postAuthLogin
func (h *Handler) Login(c *gin.Context) {
	var req dtos.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: "invalid request body"})
		return
	}

	authResult, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		h.handleError(c, err)
		return
	}

	authcookie.SetAccessToken(c, authResult.Tokens.AccessToken, h.service.AccessTTL(), h.secureCookie)
	authcookie.SetRefreshToken(c, authResult.Tokens.RefreshToken, h.service.RefreshTTL(), h.secureCookie)
	c.JSON(http.StatusOK, dtos.AuthResponse{
		User: dtos.UserResponse{
			ID:    authResult.User.ID.String(),
			Email: authResult.User.Email,
			Name:  authResult.User.Name,
		},
		Tokens: dtos.TokenResponse{
			AccessToken:  authResult.Tokens.AccessToken,
			RefreshToken: authResult.Tokens.RefreshToken,
		},
	})
}

// RefreshToken Get new access token
// @Summary Refresh access token
// @Description Use refresh token to get a new access token and refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dtos.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} dtos.TokenResponse
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Router /auth/refresh [post]
// @ID postAuthRefresh
func (h *Handler) RefreshToken(c *gin.Context) {
	var req dtos.RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: "invalid request body"})
		return
	}

	tokens, err := h.service.RefreshToken(req.RefreshToken)
	if err != nil {
		h.handleError(c, err)
		return
	}

	authcookie.SetAccessToken(c, tokens.AccessToken, h.service.AccessTTL(), h.secureCookie)
	authcookie.SetRefreshToken(c, tokens.RefreshToken, h.service.RefreshTTL(), h.secureCookie)
	c.JSON(http.StatusOK, dtos.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

// Logout Invalidate refresh tokens
// @Summary Logout user
// @Description Invalidate all refresh tokens for the authenticated user
// @Tags Auth
// @Security BearerAuth
// @Success 204
// @Failure 401 {object} dtos.ErrorResponse
// @Router /auth/logout [post]
// @ID postAuthLogout
func (h *Handler) Logout(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{Error: "unauthorized"})
		return
	}

	err := h.service.Logout(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: "logout failed"})
		return
	}
	authcookie.ClearAuthCookies(c, h.secureCookie)

	c.Status(http.StatusNoContent)
}

// Session returns the authenticated user from the current access token cookie.
// @Summary Current session
// @Description Returns the current authenticated user
// @Tags Auth
// @Security BearerAuth
// @Success 200 {object} dtos.SessionResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Router /auth/session [get]
// @ID getAuthSession
func (h *Handler) Session(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{Error: "unauthorized"})
		return
	}

	user, err := h.service.Session(userID.(string))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dtos.SessionResponse{
		User: dtos.UserResponse{
			ID:    user.ID.String(),
			Email: user.Email,
			Name:  user.Name,
		},
	})
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrUserExists):
		c.JSON(http.StatusConflict, dtos.ErrorResponse{Error: "email already in use"})
	case errors.Is(err, domain.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrInvalidToken):
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrTokenExpired):
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: "internal server error"})
	}
}
