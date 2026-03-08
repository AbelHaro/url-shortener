// Package http
// @title           URL Shortener API
// @version         1.0
// @description     API for shortening URLs
// @host            localhost:8080
// @BasePath        /api/v1
package http

import (
	"errors"
	"log"
	"net/http"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/AbelHaro/url-shortener/backend/internal/service/url"
	"github.com/gin-gonic/gin"
)

type URLHandler struct {
	service *url.Service
}

func NewURLHandler(svc *url.Service) *URLHandler {
	return &URLHandler{service: svc}
}

// Create shorten URL
// @Summary Shorten a URL
// @Description Create a shortened URL from a long URL
// @Tags URLs
// @Accept json
// @Produce json
// @Param request body CreateShortenRequest true "Request body"
// @Success 201 {object} domain.URL
// @Failure 400 {object} ErrorResponse
// @Router /shorten [post]
func (h *URLHandler) Create(c *gin.Context) {
	var req CreateShortenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	urlCreated, err := h.service.Store(req.OriginalUrl)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, urlCreated)
}

// Redirect to original URL
// @Summary Redirect to original URL
// @Description Redirects a shortened URL to its original URL
// @Tags URLs
// @Param shortURL path string true "Short URL"
// @Success 301
// @Router /{shortURL} [get]
func (h *URLHandler) Redirect(c *gin.Context) {
	shortURL := c.Param("shortURL")

	urlFound, err := h.service.FindByShortCode(shortURL)
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
// @Failure 404 {object} ErrorResponse
// @Router /urls/{id} [get]
func (h *URLHandler) FindByID(c *gin.Context) {
	id := c.Param("id")

	urlFound, err := h.service.FindByID(id)
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
// @Failure 404 {object} ErrorResponse
// @Router /urls/short/{shortCode} [get]
func (h *URLHandler) FindByShortCode(c *gin.Context) {
	shortCode := c.Param("shortCode")

	urlFound, err := h.service.FindByShortCode(shortCode)
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
// @Failure 404 {object} ErrorResponse
// @Router /urls/{id} [delete]
func (h *URLHandler) DeleteByID(c *gin.Context) {
	id := c.Param("id")

	err := h.service.DeleteByID(id)
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
// @Param request body SearchByOriginalURLRequest true "Request body"
// @Success 200 {object} domain.URL
// @Failure 404 {object} ErrorResponse
// @Router /urls/search [post]
func (h *URLHandler) FindByOriginalURL(c *gin.Context) {
	var req SearchByOriginalURLRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	urlFound, err := h.service.FindByOriginalURL(req.URL)
	if err != nil {
		h.handleError(c, err)
		return
	}

	if urlFound == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "url not found"})
		return
	}

	c.JSON(http.StatusOK, urlFound)
}

// Health check
// @Summary Health check
// @Description Returns the health status of the API
// @Tags Health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *URLHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{Status: "ok"})
}

func (h *URLHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrURLNotFound):
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrInvalidURL):
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
}
