package http

import (
	"time"

	"github.com/AbelHaro/url-shortener/backend/docs"
	"github.com/AbelHaro/url-shortener/backend/internal/delivery/http/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(r *gin.Engine, urlHandler *URLHandler, authHandler *AuthHandler, refererMiddleware *middleware.RefererMiddleware, jwtMiddleware *middleware.JWTMiddleware) {
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

	r.GET("/health", urlHandler.Health)

	if gin.IsDebugging() {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(
			swaggerFiles.NewHandler(),
			ginSwagger.URL("/swagger/doc.json"),
		))
	}

	api := r.Group("/api/v1")
	api.Use(refererMiddleware.Authenticate())
	{
		// Public auth endpoints
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Protected auth endpoints
		authProtected := api.Group("/auth")
		authProtected.Use(jwtMiddleware.Authenticate())
		{
			authProtected.POST("/logout", authHandler.Logout)
		}

		// URL endpoints (protected with JWT)
		urls := api.Group("")
		urls.Use(jwtMiddleware.Authenticate())
		{
			urls.POST("/shorten", urlHandler.Create)
			urls.GET("/urls/short/:shortCode", urlHandler.FindByShortCode)
			urls.GET("/urls/:id", urlHandler.FindByID)
			urls.DELETE("/urls/:id", urlHandler.DeleteByID)
			urls.POST("/urls/search", urlHandler.FindByOriginalURL)
		}

		// Public redirect endpoint (no auth required)
		api.GET("/:shortURL", urlHandler.Redirect)
	}

}
