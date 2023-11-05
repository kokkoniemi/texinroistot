package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kokkoniemi/texinroistot/auth"
)

func main() {
	app := fiber.New()

	app.Static("/", "../texinroistot-ui/dist")
	app.Static("/manage", "../texinroistot-ui/dist")

	api := app.Group("/api")
	api.Post("/login", auth.LoginHandler)
	api.Post("/logout", auth.LogoutHandler)
	api.Get("/me", auth.UserInfoHandler)

	app.Listen(":6969")
}
