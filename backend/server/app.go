package server

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/AbelHaro/url-shortener/backend/internal/delivery/http"
	"github.com/AbelHaro/url-shortener/backend/internal/delivery/http/auth"
	"github.com/AbelHaro/url-shortener/backend/internal/delivery/http/health"
	"github.com/AbelHaro/url-shortener/backend/internal/delivery/http/middleware"
	rangehandler "github.com/AbelHaro/url-shortener/backend/internal/delivery/http/range"
	"github.com/AbelHaro/url-shortener/backend/internal/delivery/http/url"
	"github.com/AbelHaro/url-shortener/backend/internal/infrastructure/database"
	authRepo "github.com/AbelHaro/url-shortener/backend/internal/repository/auth"
	counterRepo "github.com/AbelHaro/url-shortener/backend/internal/repository/counter"
	rangeRepo "github.com/AbelHaro/url-shortener/backend/internal/repository/range"
	urlRepo "github.com/AbelHaro/url-shortener/backend/internal/repository/url"
	authSvc "github.com/AbelHaro/url-shortener/backend/internal/service/auth"
	counterSvc "github.com/AbelHaro/url-shortener/backend/internal/service/counter"
	jwtSvc "github.com/AbelHaro/url-shortener/backend/internal/service/jwt"
	rangeSvc "github.com/AbelHaro/url-shortener/backend/internal/service/range"
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

	urlRepoInstance := urlRepo.NewPostgresRepository(db)
	counterRepoInstance := counterRepo.NewPostgresRepository(db)
	authRepoInstance := authRepo.NewPostgresRepository(db)
	rangeRepoInstance := rangeRepo.NewPostgresRepository(db)

	counter, err := counterSvc.NewService(counterRepoInstance)
	if err != nil {
		log.Fatalf("Failed to initialize counter service: %v", err)
	}

	urlService := urlSvc.NewService(urlRepoInstance, counter)
	rangeService := rangeSvc.NewService(rangeRepoInstance)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	accessTTLStr := os.Getenv("JWT_ACCESS_TOKEN_TTL")
	if accessTTLStr == "" {
		accessTTLStr = "15m"
	}
	accessTTL, err := parseTimeString(accessTTLStr)
	if err != nil {
		log.Fatalf("Invalid JWT_ACCESS_TOKEN_TTL format: %v", err)
	}

	refreshTTLStr := os.Getenv("JWT_REFRESH_TOKEN_TTL")
	if refreshTTLStr == "" {
		refreshTTLStr = "168h"
	}
	refreshTTL, err := parseTimeString(refreshTTLStr)
	if err != nil {
		log.Fatalf("Invalid JWT_REFRESH_TOKEN_TTL format: %v", err)
	}

	jwtService := jwtSvc.NewService(jwtSecret, accessTTL, refreshTTL)

	authService := authSvc.NewService(authRepoInstance, jwtService)

	router := gin.Default()

	if proxies := os.Getenv("TRUSTED_PROXIES"); proxies != "" {
		if err := router.SetTrustedProxies(strings.Split(proxies, ",")); err != nil {
			log.Fatalf("Failed to set trusted proxies: %v", err)
		}
	}

	urlHandler := url.NewHandler(urlService)
	healthHandler := health.NewHandler()
	authHandler := auth.NewHandler(authService)
	rangeHandlerInstance := rangehandler.NewHandler(rangeService)
	refererMiddleware := middleware.NewRefererMiddleware()
	jwtMiddleware := middleware.NewJWTMiddleware(authService)

	http.SetupRoutes(router, urlHandler, healthHandler, authHandler, rangeHandlerInstance, refererMiddleware, jwtMiddleware)

	return &App{router: router, db: db}
}

func (a *App) Run(addr string) error {
	return a.router.Run(addr)
}

func parseTimeString(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}
