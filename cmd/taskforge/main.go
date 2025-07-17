// cmd/taskforge/main.go
package main

import (
	"fmt"
	"net/http"

	"github.com/agincgit/taskforge/config"
	"github.com/agincgit/taskforge/server"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 1) Load configuration (including DB fields you just added)
	cfg := config.GetConfig("config.json")

	// 2) Build the Postgres DSN from cfg
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort,
	)

	// 3) Open GORM database connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// 4) Create TaskForge router (with migrations & handlers)
	router, err := server.NewRouter(db)
	if err != nil {
		log.Fatalf("Failed to initialize TaskForge: %v", err)
	}

	// 5) Start HTTP server
	log.Infof("TaskForge listening on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatalf("Server exited with error: %v", err)
	}
}
