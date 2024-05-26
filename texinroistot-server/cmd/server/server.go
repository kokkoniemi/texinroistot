package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kokkoniemi/texinroistot/internal/admin"
	"github.com/kokkoniemi/texinroistot/internal/auth"
	"github.com/kokkoniemi/texinroistot/internal/stories"
)

func main() {
	app := fiber.New()

	app.Static("/", "../texinroistot-ui/dist")
	app.Static("/manage", "../texinroistot-ui/dist")

	api := app.Group("/api")
	api.Post("/login", auth.LoginHandler)
	api.Post("/logout", auth.LogoutHandler)
	api.Get("/me", auth.UserInfoHandler)
	api.Get("/stories", stories.ListStoriesHandler)

	adminapi := api.Group("/admin", auth.ProtectedRoute)
	adminapi.Get("/users", admin.ListUsersHandler)

	app.Listen(":6969")
}
