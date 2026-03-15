package auth

import (
	"strings"

	"github.com/kokkoniemi/texinroistot/internal/config"
	"github.com/kokkoniemi/texinroistot/internal/crypt"
	"github.com/kokkoniemi/texinroistot/internal/db"
)

func ensureUserProfile(email string) error {
	userRepo := db.NewUserRepository()
	_, err := userRepo.Create(db.User{
		Hash:    userHashForEmail(email),
		IsAdmin: isConfiguredAdminEmail(email),
	})
	return err
}

func isConfiguredAdminEmail(email string) bool {
	normalizedEmail := normalizeEmail(email)
	if normalizedEmail == "" {
		return false
	}

	for _, configured := range strings.Split(config.AdminEmails, ",") {
		if normalizedEmail == normalizeEmail(configured) {
			return true
		}
	}

	return false
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func userHashForEmail(email string) string {
	return crypt.Hash(normalizeEmail(email))
}
