package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kokkoniemi/texinroistot/internal/crypt"
	"github.com/kokkoniemi/texinroistot/internal/db"
)

type UserInfo struct {
	LoggedIn bool   `json:"loggedIn"`
	Email    string `json:"email"`
	IsAdmin  bool   `json:"isAdmin"`
	Hash     string `json:"-"`
}

func UserInfoHandler(c *fiber.Ctx) error {
	user, err := getUserInfo(c)

	if err != nil {
		return err
	}

	return c.JSON(user)
}

func getUserInfo(c *fiber.Ctx) (*UserInfo, error) {
	accessToken := authCookieValue(c, "a")
	if len(accessToken) == 0 {
		return loggedOutUserInfo(), nil
	}

	authService := NewAuthService()
	accessClaims, err := authService.VerifyAccessToken(accessToken)
	if err != nil {
		return loggedOutUserInfo(), nil
	}

	email, err := crypt.Decrypt(crypt.NewEncrypted(
		accessClaims.JWTUserClaims.EmailIv,
		accessClaims.JWTUserClaims.EmailHash,
	))
	if err != nil {
		return loggedOutUserInfo(), nil
	}
	emailHash := userHashForEmail(email)

	if emailHash != accessClaims.JWTUserClaims.UserID {
		return loggedOutUserInfo(), nil
	}

	userRepo := db.NewUserRepository()
	user, err := userRepo.ReadByHash(emailHash)
	if err != nil {
		return nil, err
	}

	return &UserInfo{
		LoggedIn: true,
		Email:    email,
		IsAdmin:  user != nil && user.IsAdmin,
		Hash:     emailHash,
	}, nil
}

func loggedOutUserInfo() *UserInfo {
	return &UserInfo{
		LoggedIn: false,
		Email:    "",
		IsAdmin:  false,
		Hash:     "",
	}
}
