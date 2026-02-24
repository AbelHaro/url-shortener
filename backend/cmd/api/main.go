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
