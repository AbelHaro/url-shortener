package http

import (
	"fmt"
	"time"

	"github.com/AbelHaro/url-shortener/backend/docs"
	"github.com/AbelHaro/url-shortener/backend/internal/config"
	"github.com/AbelHaro/url-shortener/backend/internal/delivery/http/auth"
	"github.com/AbelHaro/url-shortener/backend/internal/delivery/http/health"
	"github.com/AbelHaro/url-shortener/backend/internal/delivery/http/middleware"
	statisticHandler "github.com/AbelHaro/url-shortener/backend/internal/delivery/http/statistic"
	"github.com/AbelHaro/url-shortener/backend/internal/delivery/http/url"
	authRepo "github.com/AbelHaro/url-shortener/backend/internal/repository/auth"
	idrangesRepo "github.com/AbelHaro/url-shortener/backend/internal/repository/idsranges"
	statisticRepo "github.com/AbelHaro/url-shortener/backend/internal/repository/statistic"
	urlRepo "github.com/AbelHaro/url-shortener/backend/internal/repository/url"
	authSvc "github.com/AbelHaro/url-shortener/backend/internal/service/auth"
	counterSvc "github.com/AbelHaro/url-shortener/backend/internal/service/counter"
	idrangesSvc "github.com/AbelHaro/url-shortener/backend/internal/service/idsranges"
	jwtSvc "github.com/AbelHaro/url-shortener/backend/internal/service/jwt"
	statisticSvc "github.com/AbelHaro/url-shortener/backend/internal/service/statistic"
	urlSvc "github.com/AbelHaro/url-shortener/backend/internal/service/url"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, urlHandler *url.Handler, statisticsHandler *statisticHandler.Handler, healthHandler *health.Handler, authHandler *auth.Handler, authService *authSvc.Service, appConfig *config.AppConfig) *gin.Engine {
	docs.SwaggerInfo.Title = "URL Shortener API"
	docs.SwaggerInfo.Description = "API for shortening and managing URLs"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http"}

	// Create all middlewares
	refererMiddleware := middleware.NewRefererMiddleware()
	jwtMiddleware := middleware.NewJWTMiddleware(authService, appConfig.Production)
	// In debug mode (tests), disable rate limiting to allow bulk test requests
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(100, time.Minute, gin.IsDebugging())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "https://url-shortener.abelharo.me"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/health", rateLimitMiddleware.Limit(), healthHandler.Health)

	if gin.IsDebugging() {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(
			swaggerFiles.NewHandler(),
			ginSwagger.URL("/swagger/doc.json"),
		))
	}

	api := r.Group("/api/v1")
	api.Use(rateLimitMiddleware.Limit())
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
		urls.POST("/urls/short/:shortCode/resolve", statisticsHandler.Resolve)
		urls.Use(jwtMiddleware.Authenticate())
		{
			urls.POST("/shorten", urlHandler.Create)
			urls.GET("/urls/:id", urlHandler.FindByID)
			urls.GET("/urls/:id/statistics", statisticsHandler.GetDashboard)
			urls.PATCH("/urls/:id", urlHandler.UpdateByID)
			urls.DELETE("/urls/:id", urlHandler.DeleteByID)
			urls.POST("/urls/search", urlHandler.FindByOriginalURL)
			urls.GET("/urls", urlHandler.FindByAllByUserID)
		}
	}

	return r

}

// NewConfiguredRouter creates and configures a Gin router with all handlers, middleware, and services
// initialized using the provided database connection and JWT configuration
func NewConfiguredRouter(db *gorm.DB, urlCache urlRepo.Cache, appConfig *config.AppConfig) (*gin.Engine, error) {
	router := gin.Default()

	// Initialize repositories
	urlRepoInstance := urlRepo.NewPostgresRepository(db)
	authRepoInstance := authRepo.NewPostgresRepository(db)
	idrangesRepoInstance := idrangesRepo.NewPostgresRepository(db)
	statisticRepoInstance := statisticRepo.NewPostgresRepository(db)

	// Initialize services
	idrangesSvcInstance := idrangesSvc.NewService(idrangesRepoInstance)

	counterService, err := counterSvc.NewService(idrangesSvcInstance)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize counter service: %w", err)
	}

	urlService := urlSvc.NewService(urlRepoInstance, urlCache, counterService)
	statisticService := statisticSvc.NewService(statisticRepoInstance)
	jwtService := jwtSvc.NewService(appConfig.JWTSecret, appConfig.AccessTTL, appConfig.RefreshTTL)
	authService := authSvc.NewService(authRepoInstance, jwtService)

	// Initialize handlers
	urlHandler := url.NewHandler(urlService)
	statisticsHandler := statisticHandler.NewHandler(statisticService, urlService)
	healthHandler := health.NewHandler()
	authHandler := auth.NewHandler(authService, appConfig.Production)

	// Setup routes (all middlewares are created inside SetupRoutes)
	SetupRoutes(router, urlHandler, statisticsHandler, healthHandler, authHandler, authService, appConfig)

	return router, nil
}
