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
	Port        int
	WebhookURL  url.URL
	DatabaseURL string
	JwtKey      string
	TokenTTL    time.Duration
	AuthTTL     time.Duration
}

var (
	ErrWebhookURLRequiredError       = errors.New("WEBHOOK_URL env var is required")
	ErrConnectionStringRequiredError = errors.New("DB_URL env var is required")
	ErrJWTKeyRequiredError           = errors.New("JWT_KEY env var is required")
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

	key := os.Getenv("JWT_KEY")
	if key == "" {
		return nil, ErrJWTKeyRequiredError
	}

	tokenTTL := os.Getenv("TOKEN_TTL")
	if tokenTTL == "" {
		tokenTTL = "1m"
	}

	tokenTTLDuration, err := time.ParseDuration(tokenTTL)
	if err != nil {
		return nil, err
	}

	authTTL := os.Getenv("AUTH_TTL")
	if authTTL == "" {
		authTTL = "1h"
	}

	authTTLDuration, err := time.ParseDuration(authTTL)
	if err != nil {
		return nil, err
	}

	return &Config{
		Port:        port,
		WebhookURL:  *webhookURL,
		DatabaseURL: dbConnStr,
		JwtKey:      key,
		TokenTTL:    tokenTTLDuration,
		AuthTTL:     authTTLDuration,
	}, nil
}
