// cmd/taskforge/main.go
package main

import (
	"fmt"
	"net/http"

	"github.com/agincgit/taskforge"
	"github.com/agincgit/taskforge/config"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg := config.GetConfig("config.json")

	// Build the Postgres DSN
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort,
	)

	// Open GORM database connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize TaskForge router
	router, err := taskforge.NewRouter(db)
	if err != nil {
		log.Fatalf("Failed to initialize TaskForge: %v", err)
	}

	// Start HTTP server
	log.Infof("TaskForge listening on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatalf("Server exited with error: %v", err)
	}
}
