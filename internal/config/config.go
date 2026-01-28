// Package config provides application configuration loading for the default TaskForge implementation.
package config

import (
	"crypto/rand"
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/rs/zerolog/log"
)

// Config holds all application settings, loaded from environment variables.
type Config struct {
	// Database connection
	DBHost     string `env:"TASKFORGE_DB_HOST" envDefault:"localhost"`
	DBPort     string `env:"TASKFORGE_DB_PORT" envDefault:"5432"`
	DBUser     string `env:"TASKFORGE_DB_USER" envDefault:"postgres"`
	DBPassword string `env:"TASKFORGE_DB_PASSWORD,required"`
	DBName     string `env:"TASKFORGE_DB_NAME" envDefault:"taskforge_db"`
	DBSSLMode  string `env:"TASKFORGE_DB_SSLMODE" envDefault:"disable"`

	// HTTP server
	Port string `env:"TASKFORGE_PORT" envDefault:"8080"`

	// Logging
	LogLevel string `env:"TASKFORGE_LOG_LEVEL" envDefault:"info"`

	// Hostname (auto-detected if not set)
	HostName string `env:"TASKFORGE_HOSTNAME"`
}

// generateWorkerName is used only as a fallback if hostname detection fails.
func generateWorkerName() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("worker-%x", b)
}

// Load parses environment variables into Config.
func Load() (*Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Auto-detect hostname if not provided
	if cfg.HostName == "" {
		hostname, err := os.Hostname()
		if err != nil || hostname == "" {
			log.Warn().Err(err).Msg("Unable to detect hostname, generating random")
			cfg.HostName = generateWorkerName()
		} else {
			cfg.HostName = hostname
		}
	}

	log.Debug().
		Str("host", cfg.DBHost).
		Str("port", cfg.DBPort).
		Str("database", cfg.DBName).
		Str("server_port", cfg.Port).
		Msg("Configuration loaded")

	return &cfg, nil
}
