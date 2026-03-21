package authors

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/kokkoniemi/texinroistot/internal/db"
)

var allowedStoryTypes = map[string]bool{
	"writer":     true,
	"drawer":     true,
	"translator": true,
}

func parseAuthorHash(raw string) (string, error) {
	authorHash := strings.TrimSpace(raw)
	if authorHash == "" {
		return "", fmt.Errorf("authorHash is required")
	}
	return authorHash, nil
}

func ListAuthorStoriesHandler(c *fiber.Ctx) error {
	authorHash, err := parseAuthorHash(c.Params("authorHash"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	storyType, err := parseAllowedValue(c.Query("type"), "", allowedStoryTypes)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "type is invalid"})
	}

	versionRepo := db.NewVersionRepository() // TODO: move active version to fiber context
	version, err := versionRepo.GetActive()
	if err != nil {
		return c.SendStatus(500)
	}

	storyRepo := db.NewStoryRepository()
	stories, authorFound, err := storyRepo.ListByAuthorHash(version, authorHash, storyType)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to list author stories"})
	}
	if !authorFound {
		return c.Status(404).JSON(fiber.Map{"error": "author not found"})
	}

	return c.JSON(fiber.Map{
		"authorHash": authorHash,
		"stories":    stories,
		"meta": fiber.Map{
			"total": len(stories),
		},
		"filters": fiber.Map{
			"type": storyType,
		},
	})
}
