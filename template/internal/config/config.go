package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	LogLevel string
}

// ServerConfig holds configuration for the server
type ServerConfig struct {
	Port     int
	BasePath string
}

// DatabaseConfig holds configuration for the database
type DatabaseConfig struct {
	DynamoDBEndpoint  string
	DynamoDBRegion    string
	DynamoDBTableName string
}

// LoadConfig loads the configuration from environment variables
// with fallback to .env file
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Info().Msg("No .env file found, using environment variables")
	}

	return &Config{
		Server: ServerConfig{
			Port:     getEnvAsInt("PORT", 8080),
			BasePath: getEnv("BASE_PATH", "/api/v1"),
		},
		Database: DatabaseConfig{
			DynamoDBEndpoint:  getEnv("DYNAMODB_ENDPOINT", "http://localhost:8000"),
			DynamoDBRegion:    getEnv("DYNAMODB_REGION", "us-east-1"),
			DynamoDBTableName: getEnv("DYNAMODB_TABLE_NAME", "users"),
		},
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt gets an environment variable as an integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
