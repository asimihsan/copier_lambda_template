package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// DynamoDBRepository implements the UserRepository interface using DynamoDB
type DynamoDBRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewDynamoDBRepository creates a new DynamoDB repository
func NewDynamoDBRepository(endpoint, region, tableName string) (*DynamoDBRepository, error) {
	// Load the configuration
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if endpoint != "" {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           endpoint,
				SigningRegion: region,
			}, nil
		}
		// Return EndpointNotFoundError to allow the service to fallback to its default resolution
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create DynamoDB client
	client := dynamodb.NewFromConfig(cfg)

	// Check if table exists, create if it doesn't
	_, err = client.DescribeTable(context.Background(), &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		log.Info().Msgf("Table %s does not exist, creating...", tableName)
		_, err = client.CreateTable(context.Background(), &dynamodb.CreateTableInput{
			AttributeDefinitions: []types.AttributeDefinition{
				{
					AttributeName: aws.String("id"),
					AttributeType: types.ScalarAttributeTypeS,
				},
			},
			KeySchema: []types.KeySchemaElement{
				{
					AttributeName: aws.String("id"),
					KeyType:       types.KeyTypeHash,
				},
			},
			ProvisionedThroughput: &types.ProvisionedThroughput{
				ReadCapacityUnits:  aws.Int64(5),
				WriteCapacityUnits: aws.Int64(5),
			},
			TableName: aws.String(tableName),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create table: %w", err)
		}
	}

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

	var users []User
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
