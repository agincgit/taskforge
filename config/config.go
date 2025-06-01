package config

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config holds application settings
type Config struct {
	DBDSN string // Database DSN
	Port  string // Server port
}

// GetConfig loads config from file or environment variables
func GetConfig(name string) *Config {
	viper.SetConfigName(name)
	viper.SetConfigType("json")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Warn("No config file found, falling back to env variables")
	}

	dsn := viper.GetString("DB_DSN")
	if dsn == "" {
		dsn = os.Getenv("DB_DSN")
	}
	port := viper.GetString("PORT")
	if port == "" {
		port = os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
	}

	return &Config{
		DBDSN: dsn,
		Port:  port,
	}
}
