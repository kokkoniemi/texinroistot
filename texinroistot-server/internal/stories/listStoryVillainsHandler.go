package stories

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/kokkoniemi/texinroistot/internal/db"
)

func parseStoryHash(raw string) (string, error) {
	storyHash := strings.TrimSpace(raw)
	if storyHash == "" {
		return "", fmt.Errorf("storyHash is required")
	}
	return storyHash, nil
}

func ListStoryVillainsHandler(c *fiber.Ctx) error {
	storyHash, err := parseStoryHash(c.Params("storyHash"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	versionRepo := db.NewVersionRepository() // TODO: move active version to fiber context
	version, err := versionRepo.GetActive()
	if err != nil {
		return c.SendStatus(500)
	}

	villainRepo := db.NewVillainRepository()
	villains, storyFound, err := villainRepo.ListByStoryHash(version, storyHash)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to list story villains"})
	}
	if !storyFound {
		return c.Status(404).JSON(fiber.Map{"error": "story not found"})
	}

	return c.JSON(fiber.Map{
		"storyHash": storyHash,
		"villains":  villains,
		"meta": fiber.Map{
			"total": len(villains),
		},
	})
}
