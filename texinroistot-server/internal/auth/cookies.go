package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kokkoniemi/texinroistot/internal/config"
)

func authCookieName(baseName string) string {
	if config.CookieSecure {
		return "__Host-" + baseName
	}
	return baseName
}

func authCookieValue(c *fiber.Ctx, baseName string) string {
	primaryName := authCookieName(baseName)
	primaryValue := c.Cookies(primaryName)
	if len(primaryValue) > 0 {
		return primaryValue
	}

	// Support both secure/non-secure naming in case deployment mode changes.
	if config.CookieSecure {
		return c.Cookies(baseName)
	}

	return c.Cookies("__Host-" + baseName)
}
