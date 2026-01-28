// cmd/taskforge/main.go
package main

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/agincgit/taskforge/internal/config"
	"github.com/agincgit/taskforge/internal/server"
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
		log.Fatal().Err(err).Msg("Failed to connect database")
	}

	// 4) Create TaskForge router (with migrations & handlers)
	router, err := server.NewRouter(db)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize TaskForge")
	}

	// 5) Start HTTP server
	log.Info().Str("port", cfg.Port).Msg("TaskForge listening")
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatal().Err(err).Msg("Server exited with error")
	}
}
