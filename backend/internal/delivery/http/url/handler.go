// Package url provides HTTP handlers for URL-related endpoints.
// @title           URL Shortener API
// @version         1.0
// @description     API for shortening URLs
// @host            localhost:8080
// @BasePath        /api/v1
package url

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/AbelHaro/url-shortener/backend/internal/dtos"
	"github.com/AbelHaro/url-shortener/backend/internal/service/url"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	Service *url.Service
}

func NewHandler(svc *url.Service) *Handler {
	return &Handler{Service: svc}
}

// Create shorten URL
// @Summary Shorten a URL
// @Description Create a shortened URL from a long URL
// @Tags URLs
// @Accept json
// @Produce json
// @Param request body dtos.V1CreateShortenRequest true "Request body"
// @Success 201 {object} dtos.V1URLResponse
// @Failure 400 {object} dtos.V1ErrorResponse
// @Router /shorten [post]
func (h *Handler) Create(c *gin.Context) {

	ownerIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.V1ErrorResponse{Error: "user not authenticated"})
		return
	}

	ownerID := uuid.MustParse(fmt.Sprintf("%v", ownerIDRaw))

	if ownerID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, dtos.V1ErrorResponse{Error: "invalid user ID"})
		return
	}

	var req dtos.V1CreateShortenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, dtos.V1ErrorResponse{Error: "invalid request body"})
		return
	}

	urlCreated, err := h.Service.Store(req.OriginalUrl, ownerID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	urlCreatedResponse := dtos.V1URLResponse{
		ID:          urlCreated.ID,
		OriginalURL: urlCreated.OriginalURL,
		ShortCode:   urlCreated.ShortCode,
		UserID:      urlCreated.UserID,
		CreatedAt:   urlCreated.CreatedAt,
		UpdatedAt:   urlCreated.UpdatedAt,
	}

	c.JSON(http.StatusCreated, urlCreatedResponse)
}

// Redirect to original URL
// @Summary Redirect to original URL
// @Description Redirects a shortened URL to its original URL
// @Tags URLs
// @Param shortURL path string true "Short URL"
// @Success 301
// @Router /{shortURL} [get]
func (h *Handler) Redirect(c *gin.Context) {
	shortURL := c.Param("shortURL")

	urlFound, err := h.Service.FindByShortCode(shortURL)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.Redirect(http.StatusMovedPermanently, urlFound.OriginalURL)
}

// FindByID Find URL by ID
// @Summary Get URL by ID
// @Description Retrieve a URL by its ID
// @Tags URLs
// @Produce json
// @Param id path string true "URL ID"
// @Success 200 {object} domain.URL
// @Failure 404 {object} dtos.V1ErrorResponse
// @Router /urls/{id} [get]
func (h *Handler) FindByID(c *gin.Context) {
	id := c.Param("id")

	urlFound, err := h.Service.FindByID(id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, urlFound)
}

// FindByShortCode Find URL by short code
// @Summary Get URL by short code
// @Description Retrieve a URL by its short code
// @Tags URLs
// @Produce json
// @Param shortCode path string true "Short Code"
// @Success 200 {object} domain.URL
// @Failure 404 {object} dtos.V1ErrorResponse
// @Router /urls/short/{shortCode} [get]
func (h *Handler) FindByShortCode(c *gin.Context) {
	shortCode := c.Param("shortCode")

	urlFound, err := h.Service.FindByShortCode(shortCode)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, urlFound)
}

// DeleteByID Delete URL by ID
// @Summary Delete URL
// @Description Delete a URL by its ID
// @Tags URLs
// @Param id path string true "URL ID"
// @Success 204
// @Failure 404 {object} dtos.V1ErrorResponse
// @Router /urls/{id} [delete]
func (h *Handler) DeleteByID(c *gin.Context) {
	id := c.Param("id")

	err := h.Service.DeleteByID(id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// FindByOriginalURL Find URL by original URL
// @Summary Search URL by original URL
// @Description Find a shortened URL by its original URL
// @Tags URLs
// @Accept json
// @Produce json
// @Param request body dtos.V1SearchByOriginalURLRequest true "Request body"
// @Success 200 {object} domain.URL
// @Failure 404 {object} dtos.V1ErrorResponse
// @Router /urls/search [post]
func (h *Handler) FindByOriginalURL(c *gin.Context) {
	var req dtos.V1SearchByOriginalURLRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.V1ErrorResponse{Error: "invalid request body"})
		return
	}

	urlFound, err := h.Service.FindByOriginalURL(req.URL)
	if err != nil {
		h.handleError(c, err)
		return
	}

	if urlFound == nil {
		c.JSON(http.StatusNotFound, dtos.V1ErrorResponse{Error: "url not found"})
		return
	}

	c.JSON(http.StatusOK, urlFound)
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrURLNotFound):
		c.JSON(http.StatusNotFound, dtos.V1ErrorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrInvalidURL):
		c.JSON(http.StatusBadRequest, dtos.V1ErrorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, dtos.V1ErrorResponse{Error: "internal server error"})
	}
}
