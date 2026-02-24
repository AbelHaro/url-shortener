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

	repo := repository.NewPostgresURLRepository(db)
	svc := service.NewURLService(repo)

	router := gin.Default()
	handler := httpDelivery.NewURLHandler(svc)
	handler.RegisterRoutes(router)

	return &App{router: router, db: db}
}

func (a *App) Run(addr string) error {
	return a.router.Run(addr)
}
