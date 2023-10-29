package config

import (
	"os"

	"github.com/joho/godotenv"
)

type config struct {
	PORT             string
	CORS             string
	SQL_DATABASE_URL string
}

var Cfg *config = &config{}

func LoadEnvs() error {
	if err := godotenv.Load(); err != nil {
		return err
	}

	Cfg.PORT = getEnvWithDefault("PORT", "4700")
	Cfg.CORS = getEnvWithDefault("CORS", "*, http://127.0.0.1:3000")
	Cfg.SQL_DATABASE_URL = getEnvWithDefault("SQL_DATABASE_URL", "./tmp/main.db")

	return nil
}

func getEnvWithDefault(args ...string) string {
	envValue := os.Getenv(args[0])

	if envValue == "" && len(args[1]) > 0 {
		envValue = args[1]
	}

	return envValue
}
