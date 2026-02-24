// @title           URL Shortener API
// @version         1.0
// @description     API for shortening URLs
// @host            localhost:8080
// @BasePath        /api/v1
// @tag.name        URLs
// @tag.description Operations with URLs
// @tag.name        Health
// @tag.description Health check
package main

import (
	"fmt"

	"github.com/AbelHaro/url-shortener/backend/server"
)

func main() {
	app := server.NewApp()
	err := app.Run(":8080")
	if err != nil {
		fmt.Println(err)
		return
	}
}
