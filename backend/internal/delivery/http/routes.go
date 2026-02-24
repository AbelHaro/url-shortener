package http

import (
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
)

func SetupRoutes(r *gin.Engine, h *URLHandler) {
	api := r.Group("/api/v1")
	{
		api.POST("/shorten", h.Create)
		api.GET("/urls/:id", h.FindByID)
		api.DELETE("/urls/:id", h.DeleteByID)
		api.POST("/urls/search", h.FindByOriginalURL)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/:shortURL", h.Redirect)

	r.GET("/health", h.Health)
}
