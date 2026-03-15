package auth

import "github.com/gofiber/fiber/v2"

func ProtectedRoute(c *fiber.Ctx) error {
	user, err := getUserInfo(c)

	if err != nil {
		return err
	}

	if !user.LoggedIn {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	if !user.IsAdmin {
		return fiber.NewError(fiber.StatusForbidden, "forbidden")
	}

	c.Locals("user", user)
	return c.Next()
}
