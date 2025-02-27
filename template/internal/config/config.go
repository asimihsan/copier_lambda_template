package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
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
	DynamoDBEndpoint string
	DynamoDBRegion   string
	IsLocal          bool // Added to flag local development mode
}

// LoadConfig loads the configuration from environment variables
// with fallback to .env file
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Info().Msg("No .env file found, using environment variables")
	}

	cfg := &Config{
		Server: ServerConfig{
			Port:     getEnvAsInt("PORT", 8080), //nolint:mnd
			BasePath: getEnv("BASE_PATH", "/api/v1"),
		},
		Database: DatabaseConfig{
			DynamoDBEndpoint: getEnv("DYNAMODB_ENDPOINT", "http://localhost:8000"),
			DynamoDBRegion:   getEnv("DYNAMODB_REGION", "us-east-1"),
			IsLocal:          getEnvAsBool("APP_LOCAL_MODE", false),
		},
		LogLevel: getEnv("LOG_LEVEL", "debug"),
		Slack: SlackConfig{
			OverrideTableName: getEnv("OVERRIDE_TABLE_NAME", "overrides"),
			RotationTableName: getEnv("ROTATION_TABLE_NAME", "rotations"),
		},
	}

	secretsArn := getEnv("SECRETS_ARN", "")

	if secretsArn == "" {
		return nil, fmt.Errorf("SECRETS_ARN is required")
	}

	secrets, err := loadSlackSecrets(context.Background(), secretsArn)

	if err != nil {
		return nil, fmt.Errorf("failed to load Slack secrets: %w", err)
	}

	token, ok := secrets["SLACK_APP_TOKEN"]

	if !ok {
		return nil, fmt.Errorf("missing SLACK_APP_TOKEN in Secrets Manager")
	}

	cfg.Slack.BotToken = token

	secret, ok := secrets["SLACK_SIGNING_SECRET"]

	if !ok {
		return nil, fmt.Errorf("missing SLACK_SIGNING_SECRET in Secrets Manager")
	}

	cfg.Slack.SigningSecret = secret

	log.Info().Msg("Loaded Slack secrets from Secrets Manager")

	return cfg, nil
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

func loadSlackSecrets(ctx context.Context, arn string) (map[string]string, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	smClient := secretsmanager.NewFromConfig(cfg)
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(arn),
	}

	result, err := smClient.GetSecretValue(ctx, input)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch secret from Secrets Manager: %w", err)
	}

	var secrets map[string]string
	if err := json.Unmarshal([]byte(*result.SecretString), &secrets); err != nil {
		return nil, fmt.Errorf("failed to unmarshal secret string: %w", err)
	}

	return secrets, nil
}
