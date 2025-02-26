package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// DynamoDBOverrideRepository implements the OverrideRepository interface using DynamoDB
type DynamoDBOverrideRepository struct {
	client    *dynamodb.Client
	tableName string
}

var _ OverrideRepository = (*DynamoDBOverrideRepository)(nil)

// NewDynamoDBOverrideRepository creates a new DynamoDB repository for overrides
func NewDynamoDBOverrideRepository(client *dynamodb.Client, tableName string) *DynamoDBOverrideRepository {
	return &DynamoDBOverrideRepository{
		client:    client,
		tableName: tableName,
	}
}

// ListOverrides returns all overrides
func (r *DynamoDBOverrideRepository) ListOverrides(ctx context.Context) ([]Override, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
	}

	result, err := r.client.Scan(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to scan overrides: %w", err)
	}

	overrides := make([]Override, 0, len(result.Items))
	
	for _, item := range result.Items {
		var override Override
		if err := attributevalue.UnmarshalMap(item, &override); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal override")
			continue
		}
		overrides = append(overrides, override)
	}

	return overrides, nil
}

// GetOverride returns an override by ID
func (r *DynamoDBOverrideRepository) GetOverride(ctx context.Context, id string) (*Override, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"OverrideID": &types.AttributeValueMemberS{Value: id},
		},
	}

	result, err := r.client.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get override: %w", err)
	}

	if result.Item == nil {
		return nil, nil // Override not found
	}

	var override Override
	if err := attributevalue.UnmarshalMap(result.Item, &override); err != nil {
		return nil, fmt.Errorf("failed to unmarshal override: %w", err)
	}

	return &override, nil
}

// CreateOverride creates a new override
func (r *DynamoDBOverrideRepository) CreateOverride(ctx context.Context, requestedBy string, startDate, endDate time.Time) (*Override, error) {
	now := time.Now()
	override := Override{
		OverrideID:  uuid.New().String(),
		RequestedBy: requestedBy,
		StartDate:   startDate,
		EndDate:     endDate,
		Status:      OverrideStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	item, err := attributevalue.MarshalMap(override)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal override: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	}

	_, err = r.client.PutItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create override: %w", err)
	}

	return &override, nil
}

// ApproveOverride approves an override
func (r *DynamoDBOverrideRepository) ApproveOverride(ctx context.Context, id string, approvedBy string) error {
	// First, get the existing override
	existingOverride, err := r.GetOverride(ctx, id)
	if err != nil {
		return err
	}
	if existingOverride == nil {
		return fmt.Errorf("override not found: %s", id)
	}

	// Update the override
	existingOverride.Status = OverrideStatusApproved
	existingOverride.ApprovedBy = approvedBy
	existingOverride.UpdatedAt = time.Now()

	// Save the updated override
	item, err := attributevalue.MarshalMap(existingOverride)
	if err != nil {
		return fmt.Errorf("failed to marshal override: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	}

	_, err = r.client.PutItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update override: %w", err)
	}

	return nil
}

// RejectOverride rejects an override
func (r *DynamoDBOverrideRepository) RejectOverride(ctx context.Context, id string, rejectedBy string) error {
	// First, get the existing override
	existingOverride, err := r.GetOverride(ctx, id)
	if err != nil {
		return err
	}
	if existingOverride == nil {
		return fmt.Errorf("override not found: %s", id)
	}

	// Update the override
	existingOverride.Status = OverrideStatusRejected
	existingOverride.ApprovedBy = rejectedBy // Reusing ApprovedBy field to store who rejected it
	existingOverride.UpdatedAt = time.Now()

	// Save the updated override
	item, err := attributevalue.MarshalMap(existingOverride)
	if err != nil {
		return fmt.Errorf("failed to marshal override: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	}

	_, err = r.client.PutItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update override: %w", err)
	}

	return nil
}
