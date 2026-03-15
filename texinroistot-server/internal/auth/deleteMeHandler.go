package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kokkoniemi/texinroistot/internal/db"
)

func DeleteMeHandler(c *fiber.Ctx) error {
	user, err := getUserInfo(c)
	if err != nil {
		return err
	}
	if !user.LoggedIn {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	userRepo := db.NewUserRepository()
	if err := userRepo.Remove(user.Hash); err != nil {
		return err
	}

	trashCookie(c, "a")
	trashCookie(c, "r")

	return c.JSON(fiber.Map{"deleted": true})
}
