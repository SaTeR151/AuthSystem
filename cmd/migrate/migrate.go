package main

import (
	"errors"

	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/joho/godotenv"
	"github.com/sater-151/AuthSystem/internal/config"
	"github.com/sater-151/AuthSystem/internal/database/postgresql"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := godotenv.Load(); err != nil {
		logrus.Error(err)
		return
	}
	psqlConfig := config.GetPostresqlConfig()
	db, close, err := postgresql.Open(psqlConfig)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer close()
	if len(os.Args) < 2 {
		logrus.Fatal("Usage: migrate <command>\nAvailable commands: up, down")
	}
	command := os.Args[1]
	switch command {
	case "up":
		if err := db.MigrationUp(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			logrus.Fatalf("Failed to apply migrations: %v", err)
		}
		logrus.Printf("Migrations applied successfully!")
	case "down":
		if err := db.MigrationDown(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			logrus.Fatalf("Failed to rollback back migrations: %v", err)
		}
		logrus.Printf("Migrations rolled back successfully!")

	default:
		logrus.Fatalf("Unknown command: %s\nAvailable commands: up, down", command)
	}
}
