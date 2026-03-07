package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kokkoniemi/texinroistot/internal/admin"
	"github.com/kokkoniemi/texinroistot/internal/auth"
	"github.com/kokkoniemi/texinroistot/internal/stories"
	"github.com/kokkoniemi/texinroistot/internal/villains"
)

func main() {
	app := fiber.New()

	api := app.Group("/api")
	api.Post("/login", auth.LoginHandler)
	api.Post("/logout", auth.LogoutHandler)
	api.Get("/me", auth.UserInfoHandler)
	api.Get("/stories", stories.ListStoriesHandler)
	api.Get("/villains", villains.ListVillainsHandler)

	adminapi := api.Group("/admin", auth.ProtectedRoute)
	adminapi.Get("/users", admin.ListUsersHandler)

	app.Listen(":6969") // TODO: add to .env file
}
