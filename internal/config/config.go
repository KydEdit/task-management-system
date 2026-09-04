package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	Port        string
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	databaseURL, err := getEnv("DATABASE_URL", "", true)
	if err != nil {
		return nil, err
	}

	jwtSecret, err := getEnv("JWT_SECRET", "", true)
	if err != nil {
		return nil, err
	}

	port, _ := getEnv("PORT", "8080", false)

	return &Config{
		DatabaseURL: databaseURL,
		JWTSecret:   jwtSecret,
		Port:        port,
	}, nil
}

func getEnv(key, defaultValue string, required bool) (string, error) {
	value := os.Getenv(key)
	if value != "" {
		return value, nil
	}

	if required {
		return "", fmt.Errorf("Environment variable %s is required but not set", key)
	}

	return defaultValue, nil
}
