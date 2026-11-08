package main

import (
	"context"
	"log"

	"github.com/elson08/Group07_CSCE3444-FA26/backend/config"
	"github.com/elson08/Group07_CSCE3444-FA26/backend/internal"
	"github.com/elson08/Group07_CSCE3444-FA26/backend/internal/database"
)

// main loads the config, opens SQLite, and starts the API.
func main() {
	cfg, err := config.New("config.json")
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.New(cfg.DbPath)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	if count, err := db.SeedCatalog(context.Background()); err != nil {
		log.Printf("IMDb import failed, it will retry on next start: %v", err)
	} else if count > 0 {
		log.Printf("imported %d titles from IMDb", count)
	}

	server := internal.NewServer(db, *cfg)
	if err := server.Listen(); err != nil {
		log.Print(err)
	}
}
