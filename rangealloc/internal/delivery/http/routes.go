package http

import (
	rangehandler "github.com/AbelHaro/url-shortener/rangealloc/internal/delivery/http/range"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag/example/basic/docs"
)

func SetupRoutes(r *gin.Engine, rangeHandler *rangehandler.Handler) {
	docs.SwaggerInfo.Title = "Range Allocation API"
	docs.SwaggerInfo.Description = "API for allocating ranges of IDs for URL shortening"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8081"
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http"}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	if gin.IsDebugging() {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(
			swaggerFiles.NewHandler(),
			ginSwagger.URL("/swagger/doc.json"),
		))
	}

	api := r.Group("/api/v1")
	{
		rangeGroup := api.Group("/range")
		{
			rangeGroup.POST("/allocate", rangeHandler.Allocate)
			rangeGroup.PUT("/:id/offset", rangeHandler.UpdateOffset)
		}
	}
}
