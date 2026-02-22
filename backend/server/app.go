package server

import (
	httpDelivery "github.com/AbelHaro/url-shortener/backend/internal/delivery/http"
	"github.com/AbelHaro/url-shortener/backend/internal/repository"
	"github.com/AbelHaro/url-shortener/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type App struct {
	router *gin.Engine
}

func NewApp() *App {
	repo := repository.NewInMemoryURLRepository()
	svc := service.NewURLService(repo)

	router := gin.Default()
	handler := httpDelivery.NewURLHandler(svc)
	handler.RegisterRoutes(router)

	return &App{router: router}
}

func (a *App) Run(addr string) error {
	return a.router.Run(addr)
}
