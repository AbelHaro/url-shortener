package main

import (
	"fmt"

	"github.com/AbelHaro/url-shortener/rangealloc/server"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	app := server.NewApp()
	app.Run()
}
