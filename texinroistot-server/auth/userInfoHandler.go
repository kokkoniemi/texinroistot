package auth

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/kokkoniemi/texinroistot/crypt"
)

type UserInfo struct {
	LoggedIn bool   `json:"loggedIn"`
	Email    string `json:"email"`
}

func UserInfoHandler(c *fiber.Ctx) error {
	user, err := getUserInfo(c)

	if err != nil {
		return err
	}

	return c.JSON(user)
}

func getUserInfo(c *fiber.Ctx) (*UserInfo, error) {
	accessToken := c.Cookies("a")
	if len(accessToken) == 0 {
		return &UserInfo{
			LoggedIn: false,
			Email:    "",
		}, nil
	}

	authService := NewAuthService()
	accessClaims, err := authService.VerifyAccessToken(accessToken)
	if err != nil {
		return nil, fmt.Errorf("access token verification failed")
	}

	email, err := crypt.Decrypt(crypt.NewEncrypted(
		accessClaims.JWTUserClaims.EmailIv,
		accessClaims.JWTUserClaims.EmailHash,
	))
	if err != nil {
		return nil, fmt.Errorf("malformed access token")
	}
	emailHash := crypt.Hash(email)

	if emailHash != accessClaims.JWTUserClaims.UserID {
		return nil, fmt.Errorf("malformed access token")
	}

	return &UserInfo{
		LoggedIn: true,
		Email:    email,
	}, nil
}
