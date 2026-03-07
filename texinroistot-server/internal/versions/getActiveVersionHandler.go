package versions

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kokkoniemi/texinroistot/internal/db"
)

func GetActiveVersionHandler(c *fiber.Ctx) error {
	versionRepo := db.NewVersionRepository()
	version, err := versionRepo.GetActive()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to load active version"})
	}
	stats, err := versionRepo.GetStats(version.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to load active version stats"})
	}

	return c.JSON(fiber.Map{
		"version": version,
		"stats":   stats,
	})
}
