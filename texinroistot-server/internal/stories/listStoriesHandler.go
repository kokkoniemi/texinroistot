package stories

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/kokkoniemi/texinroistot/internal/db"
)

func ListStoriesHandler(c *fiber.Ctx) error {
	storyRepo := db.NewStoryRepository()
	limit, err := strconv.ParseInt(c.Params("limit", "25"), 10, 64)
	if err != nil {
		return c.SendStatus(400)
	}
	stories, err := storyRepo.List(limit, 0)

	return c.JSON(fiber.Map{"stories": stories})
}
