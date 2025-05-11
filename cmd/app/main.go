package main

import (
	"log"

	"github.com/CAATHARSIS/task-tracking/internal/config"
	"github.com/CAATHARSIS/task-tracking/pkg/database"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("PostgreSQL connection error: %v", err)
	}
	defer db.Close()

	var version string
	if err := db.QueryRow("SELECT version();").Scan(&version); err != nil {
		log.Fatalf("PosgrteSQL version check failed: %v", version)
	}
	log.Printf("PostgreSQL connected! Version: %s", version)
}
