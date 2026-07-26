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
// @Param request body dtos.CreateShortenRequest true "Request body"
// @Success 201 {object} dtos.URLResponse
// @Failure 400 {object} dtos.ErrorResponse
// @Router /shorten [post]
// @ID postShortenURL
func (h *Handler) Create(c *gin.Context) {

	ownerIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{Error: "user not authenticated"})
		return
	}

	ownerID := uuid.MustParse(fmt.Sprintf("%v", ownerIDRaw))

	if ownerID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{Error: "invalid user ID"})
		return
	}

	var req dtos.CreateShortenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: "invalid request body"})
		return
	}

	urlCreated, err := h.Service.Store(req.OriginalUrl, ownerID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	urlCreatedResponse := dtos.URLResponse{
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
// @ID getRedirect
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
// @Failure 404 {object} dtos.ErrorResponse
// @Router /urls/{id} [get]
// @ID getURLByID
func (h *Handler) FindByID(c *gin.Context) {
	id := c.Param("id")

	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	urlFound, err := h.Service.FindByIDForUser(c.Request.Context(), id, userID)
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
// @Failure 404 {object} dtos.ErrorResponse
// @Router /urls/short/{shortCode} [get]
// @ID getURLByShortCode
func (h *Handler) FindByShortCode(c *gin.Context) {
	shortCode := c.Param("shortCode")

	urlFound, err := h.Service.FindByShortCode(shortCode)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, urlFound)
}

// UpdateByID changes the destination of a URL owned by the current user.
// @Summary Update URL destination
// @Description Change the original URL while preserving the short code
// @Tags URLs
// @Accept json
// @Produce json
// @Param id path string true "URL ID"
// @Param request body dtos.UpdateURLRequest true "New destination"
// @Success 200 {object} dtos.URLResponse
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Failure 404 {object} dtos.ErrorResponse
// @Router /urls/{id} [patch]
// @ID updateURLByID
func (h *Handler) UpdateByID(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	var request dtos.UpdateURLRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: "invalid request body"})
		return
	}

	updatedURL, err := h.Service.UpdateOriginalURLForUser(
		c.Request.Context(),
		c.Param("id"),
		userID,
		request.OriginalURL,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dtos.URLResponse{
		ID:          updatedURL.ID,
		OriginalURL: updatedURL.OriginalURL,
		ShortCode:   updatedURL.ShortCode,
		UserID:      updatedURL.UserID,
		CreatedAt:   updatedURL.CreatedAt,
		UpdatedAt:   updatedURL.UpdatedAt,
	})
}

// DeleteByID Delete URL by ID
// @Summary Delete URL
// @Description Delete a URL by its ID
// @Tags URLs
// @Param id path string true "URL ID"
// @Success 204
// @Failure 404 {object} dtos.ErrorResponse
// @Router /urls/{id} [delete]
// @ID deleteURLByID
func (h *Handler) DeleteByID(c *gin.Context) {
	id := c.Param("id")

	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	err := h.Service.DeleteByIDForUser(c.Request.Context(), id, userID)
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
// @Param request body dtos.SearchByOriginalURLRequest true "Request body"
// @Success 200 {object} domain.URL
// @Failure 404 {object} dtos.ErrorResponse
// @Router /urls/search [post]
// @ID postURLsSearch
func (h *Handler) FindByOriginalURL(c *gin.Context) {
	var req dtos.SearchByOriginalURLRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: "invalid request body"})
		return
	}

	urlFound, err := h.Service.FindByOriginalURL(req.OriginalURL)
	if err != nil {
		h.handleError(c, err)
		return
	}

	if urlFound == nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{Error: "url not found"})
		return
	}

	c.JSON(http.StatusOK, urlFound)
}

// FindByAllByUserID Find all URLs by user ID
// @Summary Get all URLs by user ID
// @Description Retrieve all shortened URLs created by a specific user
// @Tags URLs
// @Produce json
// @Success 200 {array} domain.URL
// @Failure 401 {object} dtos.ErrorResponse
// @Router /urls [get]
// @ID getURLsByUserID
func (h *Handler) FindByAllByUserID(c *gin.Context) {
	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{Error: "user not authenticated"})
		return
	}

	userID := uuid.MustParse(fmt.Sprintf("%v", userIDRaw))

	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{Error: "invalid user ID"})
		return
	}

	urlsFound, err := h.Service.FindAllByUserID(userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, urlsFound)
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrURLNotFound):
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrInvalidURL):
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrInvalidID):
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: "internal server error"})
	}
}

func authenticatedUserID(c *gin.Context) (uuid.UUID, bool) {
	rawUserID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{Error: "user not authenticated"})
		return uuid.Nil, false
	}

	userID, err := uuid.Parse(fmt.Sprintf("%v", rawUserID))
	if err != nil || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{Error: "invalid user ID"})
		return uuid.Nil, false
	}
	return userID, true
}
