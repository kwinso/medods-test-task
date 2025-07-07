package config

import (
	"errors"
	"net/url"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	Port             int
	WebhookURL       url.URL
	DatabaseURL      string
	JwtKey           string
	TokenTTL         time.Duration
	AuthTTL          time.Duration
	MigrationsSource string
}

var (
	ErrWebhookURLRequiredError       = errors.New("AUTH_WEBHOOK_URL env var is required")
	ErrConnectionStringRequiredError = errors.New("AUTH_DB_URL env var is required")
	ErrJWTKeyRequiredError           = errors.New("AUTH_JWT_KEY env var is required")
)

func Load() (*Config, error) {
	port := 8080
	envPort := os.Getenv("AUTH_PORT")

	if envPort != "" {
		parsedPort, err := strconv.Atoi(envPort)
		if err != nil {
			return nil, err
		}
		port = parsedPort
	}

	envUrl := os.Getenv("AUTH_WEBHOOK_URL")
	if envUrl == "" {
		return nil, ErrWebhookURLRequiredError
	}
	webhookURL, err := url.Parse(envUrl)
	if err != nil {
		return nil, err
	}

	dbConnStr := os.Getenv("AUTH_DB_URL")
	if dbConnStr == "" {
		return nil, ErrConnectionStringRequiredError
	}

	key := os.Getenv("AUTH_JWT_KEY")
	if key == "" {
		return nil, ErrJWTKeyRequiredError
	}

	tokenTTL := os.Getenv("AUTH_TOKEN_TTL")
	if tokenTTL == "" {
		tokenTTL = "5m"
	}

	tokenTTLDuration, err := time.ParseDuration(tokenTTL)
	if err != nil {
		return nil, err
	}

	authTTL := os.Getenv("AUTH_AUTH_TTL")
	if authTTL == "" {
		authTTL = "1h"
	}

	authTTLDuration, err := time.ParseDuration(authTTL)
	if err != nil {
		return nil, err
	}

	migrationsSource := os.Getenv("AUTH_MIGRATIONS_SOURCE")

	return &Config{
		Port:             port,
		WebhookURL:       *webhookURL,
		DatabaseURL:      dbConnStr,
		JwtKey:           key,
		TokenTTL:         tokenTTLDuration,
		AuthTTL:          authTTLDuration,
		MigrationsSource: migrationsSource,
	}, nil
}
