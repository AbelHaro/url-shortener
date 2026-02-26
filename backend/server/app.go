package server

import (
	"log"

	httpDelivery "github.com/AbelHaro/url-shortener/backend/internal/delivery/http"
	"github.com/AbelHaro/url-shortener/backend/internal/infrastructure/database"
	"github.com/AbelHaro/url-shortener/backend/internal/repository"
	"github.com/AbelHaro/url-shortener/backend/internal/service"
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

	urlRepo := repository.NewPostgresURLRepository(db)
	hashCounterRepo := repository.NewPostgresHashCounterRepository(db)

	counterSvc, err := service.NewCounterService(hashCounterRepo)
	if err != nil {
		log.Fatalf("Failed to initialize counter service: %v", err)
	}

	svc := service.NewURLService(urlRepo, counterSvc)
	err = svc.GenerateDevData()
	if err != nil {
		log.Fatalf("Failed to generate dev data: %v", err)
	}

	router := gin.Default()
	handler := httpDelivery.NewURLHandler(svc)
	httpDelivery.SetupRoutes(router, handler)

	return &App{router: router, db: db}
}

func (a *App) Run(addr string) error {
	return a.router.Run(addr)
}
