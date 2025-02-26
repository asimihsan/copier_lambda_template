package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// DynamoDBRepository implements the UserRepository interface using DynamoDB
type DynamoDBRepository struct {
	client    *dynamodb.Client
	tableName string
}

var _ UserRepository = (*DynamoDBRepository)(nil)

// NewDynamoDBRepository creates a new DynamoDB repository
func NewDynamoDBRepository(endpoint, region, tableName string, isLocal bool) (*DynamoDBRepository, error) {
    // Initialize configuration options slice
    configOptions := []func(*config.LoadOptions) error{
        config.WithRegion(region),
    }
    
    // Add dummy credentials only for local development
    if isLocal {
        configOptions = append(configOptions, config.WithCredentialsProvider(
            credentials.NewStaticCredentialsProvider("dummy", "dummy", "dummy"),
        ))
    }
    
    // Load AWS configuration
    cfg, err := config.LoadDefaultConfig(context.Background(), configOptions...)
    if err != nil {
        return nil, fmt.Errorf("failed to load AWS config: %w", err)
    }
    
    // Create service options
    var options func(*dynamodb.Options)
    
    // If endpoint is specified, use BaseEndpoint (the modern approach)
    if endpoint != "" {
        options = func(o *dynamodb.Options) {
            o.BaseEndpoint = aws.String(endpoint)
        }
    }
    
    client := dynamodb.NewFromConfig(cfg, options)
    
    return &DynamoDBRepository{
        client:    client,
        tableName: tableName,
    }, nil
}

// ListUsers returns all users
func (r *DynamoDBRepository) ListUsers(ctx context.Context) ([]User, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
	}

	result, err := r.client.Scan(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to scan users: %w", err)
	}

	users := make([]User, 0, len(result.Items))
	
	for _, item := range result.Items {
		var user User
		if err := attributevalue.UnmarshalMap(item, &user); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal user")
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

// GetUser returns a user by ID
func (r *DynamoDBRepository) GetUser(ctx context.Context, id string) (*User, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	}

	result, err := r.client.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if result.Item == nil {
		return nil, nil // User not found
	}

	var user User
	if err := attributevalue.UnmarshalMap(result.Item, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &user, nil
}

// CreateUser creates a new user
func (r *DynamoDBRepository) CreateUser(ctx context.Context, userCreate UserCreate) (*User, error) {
	now := time.Now()
	user := User{
		ID:        uuid.New().String(),
		Email:     userCreate.Email,
		Name:      userCreate.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	}

	_, err = r.client.PutItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// UpdateUser updates an existing user
func (r *DynamoDBRepository) UpdateUser(ctx context.Context, id string, userUpdate UserUpdate) (*User, error) {
	// First, get the existing user
	existingUser, err := r.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	if existingUser == nil {
		return nil, nil // User not found
	}

	// Update fields if provided
	if userUpdate.Email != nil {
		existingUser.Email = *userUpdate.Email
	}
	if userUpdate.Name != nil {
		existingUser.Name = *userUpdate.Name
	}
	existingUser.UpdatedAt = time.Now()

	// Save the updated user
	item, err := attributevalue.MarshalMap(existingUser)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	}

	_, err = r.client.PutItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return existingUser, nil
}

// DeleteUser deletes a user
func (r *DynamoDBRepository) DeleteUser(ctx context.Context, id string) error {
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	}

	_, err := r.client.DeleteItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}
