package main

import (
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
	server := internal.NewServer(db, *cfg)
	if err := server.Listen(); err != nil {
		log.Print(err)
	}
}
