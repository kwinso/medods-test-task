package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/kwinso/medods-test-task/internal"
	"github.com/kwinso/medods-test-task/internal/config"
	"github.com/kwinso/medods-test-task/internal/db"
)

func main() {
	logger := log.New(os.Stdout, "[medods-auth] ", log.LstdFlags)

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal(err)
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Fatal(err)
	}
	defer conn.Close(ctx)

	if cfg.MigrationsSource != "" {
		run, err := db.ApplyMigrations(cfg.DatabaseURL, cfg.MigrationsSource)
		if err != nil {
			logger.Fatal("Failed to apply migrations: ", err)
		}
		if run {
			logger.Println("Applied migrations")
		}
	}

	logger.Printf("Starting server on port %d\n", cfg.Port)
	if err := internal.ServeWithConfig(*cfg, conn, logger); err != nil {
		log.Fatal(err)
	}
}
