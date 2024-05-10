package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	OauthGoogleClientId string
	AppPort             string
	DbApiHost           string
	DbApiPort           string
	DbApiDatabase       string
	DbApiUsername       string
	DbApiPassword       string
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
}
