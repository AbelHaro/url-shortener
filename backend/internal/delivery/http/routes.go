package http

import (
	"github.com/AbelHaro/url-shortener/backend/docs"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
)

func SetupRoutes(r *gin.Engine, h *URLHandler) {
	docs.SwaggerInfo.Host = "localhost:8080"

	r.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerFiles.NewHandler(),
		ginSwagger.URL("/swagger/doc.json"),
	))

	r.GET("/health", h.Health)

	api := r.Group("/api/v1")
	{
		api.POST("/shorten", h.Create)
		api.GET("/urls/:id", h.FindByID)
		api.DELETE("/urls/:id", h.DeleteByID)
		api.POST("/urls/search", h.FindByOriginalURL)
	}

	r.GET("/:shortURL", h.Redirect)
}
