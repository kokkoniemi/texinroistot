package config

import (
	"os"
	"strings"

	_ "github.com/joho/godotenv/autoload"
)

// Secrets & cookies
var (
	Secret              string = getEnvConfig("ROISTOT_SECRET", "")
	Salt                string = getEnvConfig("ROISTOT_SALT", "")
	CookieSecure        bool   = getEnvConfigBool("ROISTOT_COOKIE_SECURE", true)
	CookieAccessSecret  string = getEnvConfig("ROISTOT_COOKIE_ACCESS_SECRET", "")
	CookieRefreshSecret string = getEnvConfig("ROISTOT_COOKIE_REFRESH_SECRET", "")
)

var (
	GoogleOauth2ClientID string = getEnvConfig("GOOGLE_OAUTH2_CLIENT_ID", "")
	AdminEmails          string = getEnvConfig("ROISTOT_ADMIN_EMAILS", "")
	ImportExcelURL       string = getEnvConfig(
		"ROISTOT_IMPORT_EXCEL_URL",
		"https://1drv.ms/x/s!Alxd45tPW6_6iVdpB3HmJkpWXdyF?e=BNzoBz&download=1",
	)
)

var (
	DBConnectionString string = getEnvConfig("DB_CONNECTION_STRING", "")
)

func getEnvConfig(envVar string, defaultVal string) string {
	val := os.Getenv(envVar)
	if len(val) == 0 {
		val = defaultVal
	}
	return val
}

func getEnvConfigBool(envVar string, defaultVal bool) bool {
	val := os.Getenv(envVar)
	if strings.ToLower(val) == "true" {
		return true
	} else if strings.ToLower(val) == "false" {
		return false
	}
	return defaultVal
}
