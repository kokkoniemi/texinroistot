package auth

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kokkoniemi/texinroistot/config"
	"github.com/kokkoniemi/texinroistot/crypt"
	"google.golang.org/api/idtoken"
)

type LoginPayload struct {
	Credential string `form:"credential"`
	CSRFToken  string `form:"g_csrf_token"`
}

func LoginHandler(c *fiber.Ctx) error {
	c.Accepts("application/x-www-form-urlencoded")

	payload := new(LoginPayload)

	if err := c.BodyParser(payload); err != nil {
		return err
	}

	csrf_cookie := c.Cookies("g_csrf_token")
	token, err := validateLogin(payload, csrf_cookie)

	if err != nil {
		return err
	}
	fmt.Print(token.Claims)

	err = setAuthenticationCookies(token, c)
	if err != nil {
		return err
	}

	return c.Redirect("/manage")
}

func LogoutHandler(c *fiber.Ctx) error {
	trashCookie(c, "a")
	trashCookie(c, "r")
	return c.JSON(fiber.Map{"loggedOut": true})
}

func trashCookie(c *fiber.Ctx, cookieName string) {
	if config.CookieSecure {
		cookieName = "__Host-" + cookieName
	}

	c.Cookie(&fiber.Cookie{
		Name:     cookieName,
		Value:    "",
		HTTPOnly: true,
		SameSite: "lax",
		Secure:   config.CookieSecure,
		Expires:  time.Now(),
		MaxAge:   0,
	})
}

func validateLogin(payload *LoginPayload, csrfCookie string) (*idtoken.Payload, error) {
	if len(payload.CSRFToken) <= 0 {
		return nil, fmt.Errorf("no CSRF token in request body")
	}

	if len(csrfCookie) <= 0 {
		return nil, fmt.Errorf("no CSRF token in Cookie")
	}

	if payload.CSRFToken != csrfCookie {
		return nil, fmt.Errorf("failed to verify double submit cookie")
	}

	return idtoken.Validate(
		context.Background(),
		payload.Credential,
		config.GoogleOauth2ClientID,
	)
}

func setAuthenticationCookies(token *idtoken.Payload, c *fiber.Ctx) error {
	keyBytes, err := crypt.RandomBytes(8)
	if err != nil {
		return err
	}

	sharedKey := hex.EncodeToString(keyBytes)

	// Authentication cookie
	cookieName := "a"
	if config.CookieSecure {
		cookieName = "__Host-" + cookieName
	}

	email, ok := token.Claims["email"].(string)
	if !ok {
		return fmt.Errorf("email not found")
	}

	authService := NewAuthService()

	accessToken, err := authService.CreateAccessToken(sharedKey, email)
	if err != nil {
		return err
	}

	// Both access and refresh cookie need the same max age, although token
	// max age differs
	maxAge := int(time.Hour * 24 * 7)

	c.Cookie(&fiber.Cookie{
		Name:     cookieName,
		Value:    accessToken,
		HTTPOnly: true,
		SameSite: "lax",
		Secure:   config.CookieSecure,
		MaxAge:   maxAge,
	})

	// Refresh cookie
	cookieName = "r"
	if config.CookieSecure {
		cookieName = "__Host-" + cookieName
	}

	refreshToken, err := authService.CreateRefreshToken(sharedKey)
	if err != nil {
		return err
	}

	c.Cookie(&fiber.Cookie{
		Name:     cookieName,
		Value:    refreshToken,
		HTTPOnly: true,
		SameSite: "lax",
		Secure:   config.CookieSecure,
		MaxAge:   maxAge,
	})

	return nil
}
