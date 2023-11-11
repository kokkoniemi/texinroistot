package admin

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kokkoniemi/texinroistot/internal/db"
)

type UserInfo struct {
	LoggedIn bool   `json:"loggedIn"`
	Email    string `json:"email"`
}

func ListUsersHandler(c *fiber.Ctx) error {
	userRepo := db.NewUserRepository()
	users, _, err := userRepo.List(0)

	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"users": users})
}
