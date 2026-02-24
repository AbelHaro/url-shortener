package http

import (
	"errors"
	"net/http"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/AbelHaro/url-shortener/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type URLHandler struct {
	service *service.URLService
}

func NewURLHandler(svc *service.URLService) *URLHandler {
	return &URLHandler{service: svc}
}

func (h *URLHandler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		api.POST("/shorten", h.Create)
		api.GET("/urls/:id", h.FindByID)
		api.DELETE("/urls/:id", h.DeleteByID)
		api.POST("/urls/search", h.FindByOriginalURL)
	}

	r.GET("/:shortURL", h.Redirect)

	r.GET("/health", h.Health)
}

func (h *URLHandler) Create(c *gin.Context) {
	var req struct {
		URL string `json:"long_url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	url, err := h.service.Store(req.URL)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, url)
}

func (h *URLHandler) Redirect(c *gin.Context) {
	shortURL := c.Param("shortURL")

	url, err := h.service.FindByShortURL(shortURL)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.Redirect(http.StatusMovedPermanently, url.OriginalURL)
}

func (h *URLHandler) FindByID(c *gin.Context) {
	id := c.Param("id")

	url, err := h.service.FindByID(id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, url)
}

func (h *URLHandler) DeleteByID(c *gin.Context) {
	id := c.Param("id")

	err := h.service.DeleteByID(id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *URLHandler) FindByOriginalURL(c *gin.Context) {
	var req struct {
		URL string `json:"url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	url, err := h.service.FindByOriginalURL(req.URL)
	if err != nil {
		h.handleError(c, err)
		return
	}

	if url == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "url not found"})
		return
	}

	c.JSON(http.StatusOK, url)
}

func (h *URLHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *URLHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrURLNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrInvalidURL):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrURLNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
