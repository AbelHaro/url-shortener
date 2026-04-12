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

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/AbelHaro/url-shortener/backend/internal/dtos"
	"github.com/AbelHaro/url-shortener/backend/internal/service/auth"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *auth.Service
}

func NewHandler(svc *auth.Service) *Handler {
	return &Handler{service: svc}
}

// Register Create a new user account
// @Summary Register a new user
// @Description Create a new user account with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dtos.V1RegisterRequest true "Registration details"
// @Success 201 {object} dtos.V1UserResponse
// @Failure 400 {object} dtos.V1ErrorResponse
// @Router /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req dtos.V1RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.V1ErrorResponse{Error: "invalid request body"})
		return
	}

	user, err := h.service.Register(req.Email, req.Password)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dtos.V1UserResponse{
		ID:    user.ID.String(),
		Email: user.Email,
	})
}

// Login Authenticate user and get tokens
// @Summary Login user
// @Description Authenticate with email and password, returns access and refresh tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dtos.V1LoginRequest true "Login credentials"
// @Success 200 {object} dtos.V1TokenResponse
// @Failure 400 {object} dtos.V1ErrorResponse
// @Failure 401 {object} dtos.V1ErrorResponse
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req dtos.V1LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.V1ErrorResponse{Error: "invalid request body"})
		return
	}

	tokens, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dtos.V1TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

// RefreshToken Get new access token
// @Summary Refresh access token
// @Description Use refresh token to get a new access token and refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dtos.V1RefreshTokenRequest true "Refresh token"
// @Success 200 {object} dtos.V1TokenResponse
// @Failure 400 {object} dtos.V1ErrorResponse
// @Failure 401 {object} dtos.V1ErrorResponse
// @Router /auth/refresh [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	var req dtos.V1RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.V1ErrorResponse{Error: "invalid request body"})
		return
	}

	tokens, err := h.service.RefreshToken(req.RefreshToken)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dtos.V1TokenResponse{
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
// @Failure 401 {object} dtos.V1ErrorResponse
// @Router /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.V1ErrorResponse{Error: "unauthorized"})
		return
	}

	err := h.service.Logout(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.V1ErrorResponse{Error: "logout failed"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrUserExists):
		c.JSON(http.StatusConflict, dtos.V1ErrorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, dtos.V1ErrorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrInvalidToken):
		c.JSON(http.StatusUnauthorized, dtos.V1ErrorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrTokenExpired):
		c.JSON(http.StatusUnauthorized, dtos.V1ErrorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, dtos.V1ErrorResponse{Error: "internal server error"})
	}
}
