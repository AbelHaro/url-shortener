package http

import (
	"fmt"
	"net/url"
	"time"

	"github.com/AbelHaro/url-shortener/backend/docs"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var allowedOrigins = []string{
	"http://localhost:5173/",
	"https://url-shortener.abelharo.me/",
}

func SetupRoutes(r *gin.Engine, h *URLHandler) {
	docs.SwaggerInfo.Title = "URL Shortener API"
	docs.SwaggerInfo.Description = "API for shortening and managing URLs"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http"}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "https://url-shortener.abelharo.me"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/health", h.Health)

	if gin.IsDebugging() {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(
			swaggerFiles.NewHandler(),
			ginSwagger.URL("/swagger/doc.json"),
		))
	}

	api := r.Group("/api/v1")
	api.Use(refererMiddleware())
	{
		api.POST("/shorten", h.Create)
		api.GET("/urls/short/:shortCode", h.FindByShortCode)
		api.GET("/urls/:id", h.FindByID)
		api.DELETE("/urls/:id", h.DeleteByID)
		api.POST("/urls/search", h.FindByOriginalURL)
		api.GET("/:shortURL", h.Redirect)
	}

}

func refererMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if gin.IsDebugging() {
			c.Next()
			return
		}

		referer := c.GetHeader("Referer")
		if referer == "" {
			c.AbortWithStatusJSON(401, ErrorResponse{Error: "unauthorized access"})
			return
		}

		valid := false
		refererURL, err := url.Parse(referer)
		if err == nil {
			refererHost := refererURL.Host
			for _, origin := range allowedOrigins {
				originURL, err := url.Parse(origin)
				if err == nil && refererHost == originURL.Host {
					valid = true
					break
				}
			}
		}

		fmt.Printf("Referer: %s, Valid: %t\n", referer, valid)

		if !valid {
			c.AbortWithStatusJSON(401, ErrorResponse{Error: "unauthorized access"})
			return
		}

		c.Next()
	}
}
