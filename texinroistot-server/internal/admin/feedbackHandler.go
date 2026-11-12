package admin

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/kokkoniemi/texinroistot/internal/db"
)

// maxFeedbackListed bounds the feedback shown on the admin page.
const maxFeedbackListed = 500

func ListFeedbackHandler(c *fiber.Ctx) error {
	feedback, err := db.NewFeedbackRepository().List(maxFeedbackListed)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to list feedback"})
	}
	return c.JSON(fiber.Map{"feedback": feedback})
}

func DeleteFeedbackHandler(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("feedbackID"), 10, 64)
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "invalid feedback id"})
	}

	deleted, err := db.NewFeedbackRepository().Delete(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete feedback"})
	}
	if !deleted {
		return c.Status(404).JSON(fiber.Map{"error": "feedback not found"})
	}
	return c.JSON(fiber.Map{"ok": true})
}
