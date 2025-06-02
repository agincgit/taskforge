package config

import (
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config holds all application settings.
type Config struct {
	// Logging and paths
	LogDepth        string
	LogLocation     string
	AutoAppPath     string
	AutoAppConfPath string
	AutoAppLogPath  string

	// TaskForge API
	TaskForgeAPIURL string

	// Database connection fields
	DBHost     string
	DBUser     string
	DBPassword string
	DBName     string
	DBPort     string

	// HTTP port to listen on
	Port string

	// Hostname
	HostName string
}

// generateWorkerName is used only as a fallback if hostname detection fails.
func generateWorkerName() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("worker-%x", b)
}

// GetConfig loads app settings from a JSON file under ./appconfig/<name>.json
func GetConfig(name string) *Config {
	var (
		err                               error
		curDir, confDir, logDir, hostName string
	)

	// 1) Determine current working directory
	if curDir, err = os.Getwd(); err != nil {
		log.Fatal(err)
	}
	confDir = fmt.Sprintf("%s/appconfig", curDir)
	logDir = fmt.Sprintf("%s/applogs", curDir)

	// 2) Ensure config file exists
	if _, err := os.Stat(confDir + "/" + name); os.IsNotExist(err) {
		log.Panicln("Config file not found:", name)
	}

	log.Debug("Loading config from:", confDir+"/"+name)
	viper.SetConfigName(name)
	viper.SetConfigType("json")
	viper.AddConfigPath(confDir)
	viper.AddConfigPath(".")

	// 3) Read or initialize defaults
	if err = viper.ReadInConfig(); err != nil {
		log.Debug("Config file missing or invalid, initializing defaults")

		// Set defaults here:
		viper.Set("LogLevel", "Warning")
		viper.Set("LogLocation", logDir)
		viper.Set("TaskForgeAPIURL", "https://api.taskforge.local")

		// DB defaults (override in your JSON as needed):
		viper.Set("DBHost", "localhost")
		viper.Set("DBPort", "5432")
		viper.Set("DBUser", "postgres")
		viper.Set("DBPassword", "password")
		viper.Set("DBName", "taskforge_db")

		// HTTP listening port default:
		viper.Set("Port", "8080")

		_ = viper.WriteConfigAs(confDir + "/config.json")
	} else {
		log.Debug("Config Loaded from file")
	}

	// 4) Determine HostName (primary: os.Hostname(); fallback: `hostname` exec; fallback: random)
	hostName, err = os.Hostname()
	if err != nil || hostName == "" {
		out, cmdErr := exec.Command("hostname").Output()
		if cmdErr == nil {
			hostName = strings.TrimSpace(string(out))
		} else {
			log.Warnf("Unable to detect hostname, generating random: %v", err)
			hostName = generateWorkerName()
		}
	}

	// 5) Populate & return Config
	return &Config{
		LogDepth:        viper.GetString("LogLevel"),
		LogLocation:     viper.GetString("LogLocation"),
		AutoAppPath:     curDir,
		AutoAppConfPath: confDir,
		AutoAppLogPath:  logDir,
		TaskForgeAPIURL: viper.GetString("TaskForgeAPIURL"),

		DBHost:     viper.GetString("DBHost"),
		DBPort:     viper.GetString("DBPort"),
		DBUser:     viper.GetString("DBUser"),
		DBPassword: viper.GetString("DBPassword"),
		DBName:     viper.GetString("DBName"),

		Port:     viper.GetString("Port"),
		HostName: hostName,
	}
}
