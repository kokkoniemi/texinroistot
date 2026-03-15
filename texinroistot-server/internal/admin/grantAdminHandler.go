package admin

import (
	"database/sql"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/kokkoniemi/texinroistot/internal/crypt"
	"github.com/kokkoniemi/texinroistot/internal/db"
)

type GrantAdminPayload struct {
	Email string `json:"email"`
}

func GrantAdminHandler(c *fiber.Ctx) error {
	payload := new(GrantAdminPayload)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	email := strings.ToLower(strings.TrimSpace(payload.Email))
	if email == "" {
		return c.Status(400).JSON(fiber.Map{"error": "email is required"})
	}

	userHash := crypt.Hash(email)
	userRepo := db.NewUserRepository()
	user, err := userRepo.SetAdmin(userHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "user not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "failed to grant admin rights"})
	}

	return c.JSON(fiber.Map{"user": user})
}
