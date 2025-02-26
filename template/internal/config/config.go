package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

// SlackConfig holds configuration for Slack integration.
type SlackConfig struct {
	SigningSecret     string
	BotToken          string
	OverrideTableName string
	RotationTableName string
}

// Config holds all configuration for the application, including Slack settings.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	LogLevel string
	Slack    SlackConfig
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
	IsLocal           bool // Added to flag local development mode
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
			Port:     getEnvAsInt("PORT", 8080), //nolint:mnd
			BasePath: getEnv("BASE_PATH", "/api/v1"),
		},
		Database: DatabaseConfig{
			DynamoDBEndpoint:  getEnv("DYNAMODB_ENDPOINT", "http://localhost:8000"),
			DynamoDBRegion:    getEnv("DYNAMODB_REGION", "us-east-1"),
			DynamoDBTableName: getEnv("DYNAMODB_TABLE_NAME", "users"),
			IsLocal:           getEnvAsBool("APP_LOCAL_MODE", false),
		},
		LogLevel: getEnv("LOG_LEVEL", "info"),
		Slack: SlackConfig{
			SigningSecret:     getEnv("SLACK_SIGNING_SECRET", ""),
			BotToken:          getEnv("SLACK_BOT_TOKEN", ""),
			OverrideTableName: getEnv("OVERRIDE_TABLE_NAME", "overrides"),
			RotationTableName: getEnv("ROTATION_TABLE_NAME", "rotations"),
		},
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

// getEnvAsBool gets an environment variable as a boolean or returns a default value
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")

	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}
