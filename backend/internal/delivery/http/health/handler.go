// Package health provides HTTP handler for health check endpoint.
// @title           URL Shortener API
// @version         1.0
// @description     API for shortening URLs
// @host            localhost:8080
// @BasePath        /api/v1
package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

type Response struct {
	Status string `json:"status"`
}

// Health check
// @Summary Health check
// @Description Returns the health status of the API
// @Tags Health
// @Produce json
// @Success 200 {object} Response
// @Router /health [get]
// @ID getHealth
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, Response{Status: "ok"})
}
