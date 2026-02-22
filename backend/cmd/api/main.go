package main

import "github.com/AbelHaro/url-shortener/backend/server"

func main() {
	app := server.NewApp()
	app.Run(":8080")
}
