package server

import (
	"log"
	"os"
	"strings"

	"github.com/AbelHaro/url-shortener/backend/internal/config"
	"github.com/AbelHaro/url-shortener/backend/internal/delivery/http"
	"github.com/AbelHaro/url-shortener/backend/internal/infrastructure/database"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type App struct {
	router *gin.Engine
	db     *gorm.DB
}

func NewApp() *App {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	db, err := database.NewDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	router, err := http.NewConfiguredRouter(db, cfg)
	if err != nil {
		log.Fatalf("Failed to configure router: %v", err)
	}

	if proxies := os.Getenv("TRUSTED_PROXIES"); proxies != "" {
		if err := router.SetTrustedProxies(strings.Split(proxies, ",")); err != nil {
			log.Fatalf("Failed to set trusted proxies: %v", err)
		}
	}

	return &App{router: router, db: db}
}

func (a *App) Run(addr string) error {
	return a.router.Run(addr)
}
