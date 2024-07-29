package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	OauthGoogleClientId        string
	AppPort                    string
	DbApiHost                  string
	DbApiPort                  string
	DbApiDatabase              string
	DbApiUsername              string
	DbApiPassword              string
	OauthLineRedirect          string
	CloudflareImageAuthToken   string
	ClourdlareImageAccountId   string
	ClourdlareImageAccountHash string
	ClourdlareImageDeliveryUrl string
	FacebookUrl                string
}

var APP AppConfig

func InitialiseConfig() {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	errEnv := godotenv.Load(filepath.Join(dir, ".env"))
	if errEnv != nil {
		godotenv.Load()
	}

	APP.OauthGoogleClientId = os.Getenv("OAUTH_GOOGLE_CLIENT_ID")
	APP.AppPort = os.Getenv("APP_PORT")
	APP.DbApiHost = os.Getenv("DB_API_HOST")
	APP.DbApiPort = os.Getenv("DB_API_PORT")
	APP.DbApiDatabase = os.Getenv("DB_API_DATABASE")
	APP.DbApiUsername = os.Getenv("DB_API_USERNAME")
	APP.DbApiPassword = os.Getenv("DB_API_PASSWORD")
	APP.OauthLineRedirect = os.Getenv("OAUTH_LINE_REDIRECT_URL")
	APP.CloudflareImageAuthToken = os.Getenv("CLOUDFLARE_IMAGE_AUTH_TOKEN")
	APP.ClourdlareImageAccountId = os.Getenv("CLOURDLARE_IMAGE_ACCOUNT_ID")
	APP.ClourdlareImageAccountHash = os.Getenv("CLOURDLARE_IMAGE_ACCOUNT_HASH")
	APP.ClourdlareImageDeliveryUrl = os.Getenv("CLOURDLARE_IMAGE_DELIVERY_URL")
	APP.FacebookUrl = os.Getenv("OATH_FACEBOOK_BASE_URL")
}
