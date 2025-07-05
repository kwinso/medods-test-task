package config

import (
	"errors"
	"net/url"
	"os"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	Port        int
	WebhookURL  url.URL
	DatabaseURL string
}

var (
	ErrWebhookURLRequiredError       = errors.New("WEBHOOK_URL env var is required")
	ErrConnectionStringRequiredError = errors.New("DB_URL env var is required")
)

func Load() (*Config, error) {
	port := 8080
	envPort := os.Getenv("PORT")

	if envPort != "" {
		parsedPort, err := strconv.Atoi(envPort)
		if err != nil {
			return nil, err
		}
		port = parsedPort
	}

	envUrl := os.Getenv("WEBHOOK_URL")
	if envUrl == "" {
		return nil, ErrWebhookURLRequiredError
	}
	webhookURL, err := url.Parse(envUrl)
	if err != nil {
		return nil, err
	}

	dbConnStr := os.Getenv("DB_URL")
	if dbConnStr == "" {
		return nil, ErrConnectionStringRequiredError
	}

	return &Config{
		Port:        port,
		WebhookURL:  *webhookURL,
		DatabaseURL: dbConnStr,
	}, nil
}
