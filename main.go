package main

import (
	application "newProject/app"
	middleware "newProject/middleware"
)

func main() {
	app := application.Initialize()
	app.Set("views", "views")
	app.Set("static", "static")
	app.Middleware([]string{"GET", "POST"}, "/", middleware.Auth)
	app.Middleware([]string{"GET", "POST"}, "/", middleware.Test2)
	// app.Middleware([]string{"POST"}, "/", middleware.Auth)
	app.Middleware([]string{"POST"}, "/about", middleware.Auth)
	app.Middleware([]string{"GET"}, "/about", middleware.Test2)
	app.Get("/", application.HomePage)
	app.Post("/", application.HomePost)
	app.Get("/id/{id}", application.HomePage)
	app.Get("/id/{id}/basket/{name}", application.HomePage)
	app.Get("/about", application.AboutPage)

	app.Listen(4000)
}
