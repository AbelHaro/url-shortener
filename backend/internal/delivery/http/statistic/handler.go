// Package statistic provides HTTP handlers for URL analytics.
package statistic

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/AbelHaro/url-shortener/backend/internal/domain"
	"github.com/AbelHaro/url-shortener/backend/internal/dtos"
	statisticSvc "github.com/AbelHaro/url-shortener/backend/internal/service/statistic"
	urlSvc "github.com/AbelHaro/url-shortener/backend/internal/service/url"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler resolves short URLs and serves owner-only analytics.
type Handler struct {
	statistics *statisticSvc.Service
	urls       *urlSvc.Service
}

// NewHandler creates a URL statistics handler.
func NewHandler(statistics *statisticSvc.Service, urls *urlSvc.Service) *Handler {
	return &Handler{statistics: statistics, urls: urls}
}

// Resolve records a click and returns the URL destination.
// @Summary Resolve a short URL
// @Description Resolve a short code and record one click
// @Tags URLs
// @Accept json
// @Produce json
// @Param shortCode path string true "Short Code"
// @Param request body dtos.ResolveShortURLRequest false "Resolution context"
// @Success 200 {object} dtos.ResolveShortURLResponse
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 404 {object} dtos.ErrorResponse
// @Router /urls/short/{shortCode}/resolve [post]
// @ID resolveShortURL
func (handler *Handler) Resolve(c *gin.Context) {
	var request dtos.ResolveShortURLRequest
	if err := c.ShouldBindJSON(&request); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: "invalid request body"})
		return
	}

	storedURL, err := handler.urls.FindByShortCode(c.Param("shortCode"))
	if err != nil {
		handleError(c, err)
		return
	}

	if err := handler.statistics.RecordClick(c.Request.Context(), storedURL.ID.String(), request.Referrer); err != nil {
		log.Printf("record URL click for %s: %v", storedURL.ID, err)
	}

	c.JSON(http.StatusOK, dtos.ResolveShortURLResponse{OriginalURL: storedURL.OriginalURL})
}

// GetDashboard returns aggregated analytics for a URL owned by the current user.
// @Summary Get URL statistics
// @Description Retrieve owner-only click analytics for a shortened URL
// @Tags URLs
// @Produce json
// @Param id path string true "URL ID"
// @Success 200 {object} dtos.URLStatisticsResponse
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Failure 404 {object} dtos.ErrorResponse
// @Router /urls/{id}/statistics [get]
// @ID getURLStatistics
func (handler *Handler) GetDashboard(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	storedURL, err := handler.urls.FindByIDForUser(c.Request.Context(), c.Param("id"), userID)
	if err != nil {
		handleError(c, err)
		return
	}

	dashboard, err := handler.statistics.GetDashboard(c.Request.Context(), storedURL.ID.String())
	if err != nil {
		handleError(c, err)
		return
	}

	clicksByDay := make([]dtos.DailyClickResponse, 0, len(dashboard.ClicksByDay))
	for _, item := range dashboard.ClicksByDay {
		clicksByDay = append(clicksByDay, dtos.DailyClickResponse{
			Date:   item.Date.UTC().Format(time.DateOnly),
			Clicks: item.Clicks,
		})
	}
	topReferrers := make([]dtos.ReferrerCountResponse, 0, len(dashboard.TopReferrers))
	for _, item := range dashboard.TopReferrers {
		topReferrers = append(topReferrers, dtos.ReferrerCountResponse{
			Referrer: item.Referrer,
			Clicks:   item.Clicks,
		})
	}
	recentClicks := make([]dtos.RecentClickResponse, 0, len(dashboard.RecentClicks))
	for _, item := range dashboard.RecentClicks {
		recentClicks = append(recentClicks, dtos.RecentClickResponse{
			ClickedAt: item.ClickedAt,
			Referrer:  item.Referer,
		})
	}

	c.JSON(http.StatusOK, dtos.URLStatisticsResponse{
		URL: dtos.URLResponse{
			ID:          storedURL.ID,
			OriginalURL: storedURL.OriginalURL,
			ShortCode:   storedURL.ShortCode,
			UserID:      storedURL.UserID,
			CreatedAt:   storedURL.CreatedAt,
			UpdatedAt:   storedURL.UpdatedAt,
		},
		TotalClicks:   dashboard.TotalClicks,
		LastClickedAt: dashboard.LastClickedAt,
		ClicksByDay:   clicksByDay,
		TopReferrers:  topReferrers,
		RecentClicks:  recentClicks,
	})
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

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrURLNotFound):
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{Error: domain.ErrURLNotFound.Error()})
	case errors.Is(err, domain.ErrInvalidID), errors.Is(err, domain.ErrInvalidURL):
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{Error: "internal server error"})
	}
}
