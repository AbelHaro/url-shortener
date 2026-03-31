package server

import (
	"fmt"

	"github.com/AbelHaro/url-shortener/rangealloc/config"
	"github.com/AbelHaro/url-shortener/rangealloc/internal/delivery/http"
	rangehandler "github.com/AbelHaro/url-shortener/rangealloc/internal/delivery/http/range"
	"github.com/AbelHaro/url-shortener/rangealloc/internal/infraestructure"
	rangerepository "github.com/AbelHaro/url-shortener/rangealloc/internal/repository/range"
	rangeservice "github.com/AbelHaro/url-shortener/rangealloc/internal/service/range"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type App struct {
	router *gin.Engine
	db     *gorm.DB
	cfg    *config.Config
}

func NewApp() *App {
	cfg := config.LoadConfig()

	db, err := infraestructure.NewDB(&cfg.DB)
	if err != nil {
		panic(err)
	}

	rangeRepository := rangerepository.NewPostgresRepository(db)
	rangeService := rangeservice.NewService(rangeRepository)
	rangeHandler := rangehandler.NewHandler(rangeService)

	router := gin.Default()

	http.SetupRoutes(router, rangeHandler)

	return &App{
		router: router,
		db:     db,
		cfg:    cfg,
	}
}

func (a *App) Run() {
	a.router.Run(fmt.Sprintf("%s:%s", a.cfg.Server.Address, a.cfg.Server.Port))
	fmt.Printf("Server running on %s:%s\n", a.cfg.Server.Address, a.cfg.Server.Port)
}
