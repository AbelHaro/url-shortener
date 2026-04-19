package http

import (
	"fmt"
	"time"

	"github.com/AbelHaro/url-shortener/backend/docs"
	"github.com/AbelHaro/url-shortener/backend/internal/config"
	"github.com/AbelHaro/url-shortener/backend/internal/delivery/http/auth"
	"github.com/AbelHaro/url-shortener/backend/internal/delivery/http/health"
	"github.com/AbelHaro/url-shortener/backend/internal/delivery/http/middleware"
	"github.com/AbelHaro/url-shortener/backend/internal/delivery/http/url"
	authRepo "github.com/AbelHaro/url-shortener/backend/internal/repository/auth"
	idrangesRepo "github.com/AbelHaro/url-shortener/backend/internal/repository/idsranges"
	urlRepo "github.com/AbelHaro/url-shortener/backend/internal/repository/url"
	authSvc "github.com/AbelHaro/url-shortener/backend/internal/service/auth"
	counterSvc "github.com/AbelHaro/url-shortener/backend/internal/service/counter"
	idrangesSvc "github.com/AbelHaro/url-shortener/backend/internal/service/idsranges"
	jwtSvc "github.com/AbelHaro/url-shortener/backend/internal/service/jwt"
	urlSvc "github.com/AbelHaro/url-shortener/backend/internal/service/url"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, urlHandler *url.Handler, healthHandler *health.Handler, authHandler *auth.Handler, refererMiddleware *middleware.RefererMiddleware, jwtMiddleware *middleware.JWTMiddleware) *gin.Engine {
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

	r.GET("/health", healthHandler.Health)

	if gin.IsDebugging() {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(
			swaggerFiles.NewHandler(),
			ginSwagger.URL("/swagger/doc.json"),
		))
	}

	api := r.Group("/api/v1")
	api.Use(refererMiddleware.Authenticate())
	{
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/anonymous", authHandler.AnonymousRegister)
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/refresh", authHandler.RefreshToken)
			authGroup.GET("/session", jwtMiddleware.Authenticate(), authHandler.Session)
		}

		authProtected := api.Group("/auth")
		authProtected.Use(jwtMiddleware.Authenticate())
		{
			authProtected.POST("/logout", authHandler.Logout)
		}

		urls := api.Group("")
		urls.GET("/urls/short/:shortCode", urlHandler.FindByShortCode)
		urls.Use(jwtMiddleware.Authenticate())
		{
			urls.POST("/shorten", urlHandler.Create)
			urls.GET("/urls/:id", urlHandler.FindByID)
			urls.DELETE("/urls/:id", urlHandler.DeleteByID)
			urls.POST("/urls/search", urlHandler.FindByOriginalURL)
		}
	}

	return r

}

// NewConfiguredRouter creates and configures a Gin router with all handlers, middleware, and services
// initialized using the provided database connection and JWT configuration
func NewConfiguredRouter(db *gorm.DB, appConfig *config.AppConfig) (*gin.Engine, error) {
	router := gin.Default()

	// Initialize repositories
	urlRepoInstance := urlRepo.NewPostgresRepository(db)
	authRepoInstance := authRepo.NewPostgresRepository(db)
	idrangesRepoInstance := idrangesRepo.NewPostgresRepository(db)

	// Initialize services
	idrangesSvcInstance := idrangesSvc.NewService(idrangesRepoInstance)

	counterService, err := counterSvc.NewService(idrangesSvcInstance)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize counter service: %w", err)
	}

	urlService := urlSvc.NewService(urlRepoInstance, counterService)
	jwtService := jwtSvc.NewService(appConfig.JWTSecret, appConfig.AccessTTL, appConfig.RefreshTTL)
	authService := authSvc.NewService(authRepoInstance, jwtService)

	// Initialize handlers and middleware
	urlHandler := url.NewHandler(urlService)
	healthHandler := health.NewHandler()
	authHandler := auth.NewHandler(authService, appConfig.Production)
	refererMiddleware := middleware.NewRefererMiddleware()
	jwtMiddleware := middleware.NewJWTMiddleware(authService, appConfig.Production)

	// Setup routes
	SetupRoutes(router, urlHandler, healthHandler, authHandler, refererMiddleware, jwtMiddleware)

	return router, nil
}
