package stories

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/kokkoniemi/texinroistot/internal/db"
)

func ListStoriesHandler(c *fiber.Ctx) error {
	versionRepo := db.NewVersionRepository() // TODO: move active version to fiber context
	version, err := versionRepo.GetActive()
	if err != nil {
		c.SendStatus(500)
	}
	storyRepo := db.NewStoryRepository()
	limit, err := strconv.ParseInt(c.Params("limit", "25"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err}) // TODO: do not let errors through
	}
	stories, err := storyRepo.List(version, int(limit), 0)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err}) // TODO: do not let errors through
	}

	return c.JSON(fiber.Map{"stories": stories})
}
