package server

import (
	"log"
	"os"
	"strings"

	"github.com/AbelHaro/url-shortener/backend/internal/delivery/http"
	"github.com/AbelHaro/url-shortener/backend/internal/infrastructure/database"
	counterRepo "github.com/AbelHaro/url-shortener/backend/internal/repository/counter"
	urlRepo "github.com/AbelHaro/url-shortener/backend/internal/repository/url"
	counterSvc "github.com/AbelHaro/url-shortener/backend/internal/service/counter"
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

	counter, err := counterSvc.NewService(counterRepoInstance)
	if err != nil {
		log.Fatalf("Failed to initialize counter service: %v", err)
	}

	svc := urlSvc.NewService(urlRepoInstance, counter)
	err = svc.GenerateDevData()
	if err != nil {
		log.Fatalf("Failed to generate dev data: %v", err)
	}

	router := gin.Default()

	if proxies := os.Getenv("TRUSTED_PROXIES"); proxies != "" {
		if err := router.SetTrustedProxies(strings.Split(proxies, ",")); err != nil {
			log.Fatalf("Failed to set trusted proxies: %v", err)
		}
	}

	handler := http.NewURLHandler(svc)
	http.SetupRoutes(router, handler)

	return &App{router: router, db: db}
}

func (a *App) Run(addr string) error {
	return a.router.Run(addr)
}
