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

	// TODO: check that user has admin role

	c.Locals("user", user)
	return c.Next()
}
