package server

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/AbelHaro/url-shortener/backend/internal/delivery/http"
	"github.com/AbelHaro/url-shortener/backend/internal/delivery/http/middleware"
	"github.com/AbelHaro/url-shortener/backend/internal/infrastructure/database"
	authRepo "github.com/AbelHaro/url-shortener/backend/internal/repository/auth"
	counterRepo "github.com/AbelHaro/url-shortener/backend/internal/repository/counter"
	urlRepo "github.com/AbelHaro/url-shortener/backend/internal/repository/url"
	authSvc "github.com/AbelHaro/url-shortener/backend/internal/service/auth"
	counterSvc "github.com/AbelHaro/url-shortener/backend/internal/service/counter"
	jwtSvc "github.com/AbelHaro/url-shortener/backend/internal/service/jwt"
	urlSvc "github.com/AbelHaro/url-shortener/backend/internal/service/url"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type App struct {
	router *gin.Engine
	db     *gorm.DB
}

func NewApp() *App {
	cfg := database.LoadConfig()
	db, err := database.NewDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize repositories
	urlRepoInstance := urlRepo.NewPostgresRepository(db)
	counterRepoInstance := counterRepo.NewPostgresRepository(db)
	authRepoInstance := authRepo.NewPostgresRepository(db)

	// Initialize counter service
	counter, err := counterSvc.NewService(counterRepoInstance)
	if err != nil {
		log.Fatalf("Failed to initialize counter service: %v", err)
	}

	// Initialize URL service
	urlService := urlSvc.NewService(urlRepoInstance, counter)

	// Initialize JWT service
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	accessTTLStr := os.Getenv("JWT_ACCESS_TOKEN_TTL")
	if accessTTLStr == "" {
		accessTTLStr = "15m" // Default to 15 minutes
	}
	accessTTL, err := parseTimeString(accessTTLStr)
	if err != nil {
		log.Fatalf("Invalid JWT_ACCESS_TOKEN_TTL format: %v", err)
	}

	refreshTTLStr := os.Getenv("JWT_REFRESH_TOKEN_TTL")
	if refreshTTLStr == "" {
		refreshTTLStr = "168h" // Default to 7 days
	}
	refreshTTL, err := parseTimeString(refreshTTLStr)
	if err != nil {
		log.Fatalf("Invalid JWT_REFRESH_TOKEN_TTL format: %v", err)
	}

	jwtService := jwtSvc.NewService(jwtSecret, accessTTL, refreshTTL)

	// Initialize auth service
	authService := authSvc.NewService(authRepoInstance, jwtService)

	// Initialize Gin router
	router := gin.Default()

	if proxies := os.Getenv("TRUSTED_PROXIES"); proxies != "" {
		if err := router.SetTrustedProxies(strings.Split(proxies, ",")); err != nil {
			log.Fatalf("Failed to set trusted proxies: %v", err)
		}
	}

	// Initialize handlers
	urlHandler := http.NewURLHandler(urlService)
	authHandler := http.NewAuthHandler(authService)
	refererMiddleware := middleware.NewRefererMiddleware()
	jwtMiddleware := middleware.NewJWTMiddleware(authService)

	// Setup routes
	http.SetupRoutes(router, urlHandler, authHandler, refererMiddleware, jwtMiddleware)

	return &App{router: router, db: db}
}

func (a *App) Run(addr string) error {
	return a.router.Run(addr)
}

func parseTimeString(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}
