package admin

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/kokkoniemi/texinroistot/internal/config"
	"github.com/kokkoniemi/texinroistot/internal/db"
)

func parseVersionID(raw string) (int, error) {
	versionID, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || versionID <= 0 {
		return 0, errors.New("versionID must be a positive integer")
	}
	return versionID, nil
}

func ListVersionsHandler(c *fiber.Ctx) error {
	versionRepo := db.NewVersionRepository()
	versions, err := versionRepo.List()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to list versions"})
	}

	return c.JSON(fiber.Map{
		"versions":  versions,
		"importUrl": config.ImportExcelURL,
	})
}

func ActivateVersionHandler(c *fiber.Ctx) error {
	versionID, err := parseVersionID(c.Params("versionID"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	versionRepo := db.NewVersionRepository()
	if err := versionRepo.SetActive(versionID); err != nil {
		if errors.Is(err, db.ErrVersionNotFound) {
			return c.Status(404).JSON(fiber.Map{"error": "version not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "failed to set active version"})
	}

	version, err := versionRepo.Read(versionID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to load active version"})
	}

	return c.JSON(fiber.Map{"version": version})
}

func DeleteVersionHandler(c *fiber.Ctx) error {
	versionID, err := parseVersionID(c.Params("versionID"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	versionRepo := db.NewVersionRepository()
	if err := versionRepo.Remove(versionID); err != nil {
		if errors.Is(err, db.ErrCannotDeleteActiveVersion) {
			return c.Status(409).JSON(fiber.Map{"error": "active version cannot be deleted"})
		}
		if errors.Is(err, db.ErrVersionNotFound) {
			return c.Status(404).JSON(fiber.Map{"error": "version not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete version"})
	}

	return c.JSON(fiber.Map{"deleted": true, "versionID": versionID})
}
