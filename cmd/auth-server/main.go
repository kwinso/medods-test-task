package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/kwinso/medods-test-task/internal"
	"github.com/kwinso/medods-test-task/internal/config"
	"github.com/kwinso/medods-test-task/internal/db"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	run, err := db.ApplyMigrations(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to apply migrations: ", err)
	}
	if run {
		log.Println("Applied migrations")
	}

	log.Printf("Starting server on port %d\n", cfg.Port)
	if err := internal.ServeWithConfig(*cfg, conn); err != nil {
		log.Fatal(err)
	}
}
