package platform

import (
	"fmt"
	"os"
)

// Config holds application configuration
type Config struct {
	Port       string
	DataPath   string
	DataS3URL  string
	LogLevel   string
	DBPath     string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dataPath := os.Getenv("DATA_PATH")
	if dataPath == "" {
		dataPath = "./data/sample_100_pubmed.jsonl"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = ":memory:" // Use in-memory DB by default, can be changed to file
	}

	// Validate log level
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[logLevel] {
		return nil, fmt.Errorf("invalid log level: %s", logLevel)
	}

	return &Config{
		Port:      port,
		DataPath:  dataPath,
		DataS3URL: os.Getenv("DATA_S3_URL"),
		LogLevel:  logLevel,
		DBPath:    dbPath,
	}, nil
}

// GetLogLevel returns the slog level from string
func GetLogLevel(level string) int {
	switch level {
	case "debug":
		return -4 // slog.LevelDebug
	case "info":
		return 0 // slog.LevelInfo
	case "warn":
		return 4 // slog.LevelWarn
	case "error":
		return 8 // slog.LevelError
	default:
		return 0
	}
}

