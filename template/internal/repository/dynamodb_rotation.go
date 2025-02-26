package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/rs/zerolog/log"
)

// DynamoDBRotationRepository implements the RotationRepository interface using DynamoDB
type DynamoDBRotationRepository struct {
	client    *dynamodb.Client
	tableName string
}

var _ RotationRepository = (*DynamoDBRotationRepository)(nil)

// NewDynamoDBRotationRepository creates a new DynamoDB repository for rotations
func NewDynamoDBRotationRepository(client *dynamodb.Client, tableName string) *DynamoDBRotationRepository {
	return &DynamoDBRotationRepository{
		client:    client,
		tableName: tableName,
	}
}

// GetRotation returns the current rotation
func (r *DynamoDBRotationRepository) GetRotation(ctx context.Context) (*Rotation, error) {
	// For now, we assume there's only one rotation in the system
	// In a more complex system, you might want to pass an ID or other identifier
	
	input := &dynamodb.ScanInput{
		TableName:  aws.String(r.tableName),
		Limit:      aws.Int32(1), // Just get the first rotation
	}

	result, err := r.client.Scan(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to scan rotations: %w", err)
	}

	if len(result.Items) == 0 {
		return nil, nil // No rotation found
	}

	var rotation Rotation
	if err := attributevalue.UnmarshalMap(result.Items[0], &rotation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rotation: %w", err)
	}

	return &rotation, nil
}

// AdvanceRotation moves to the next person in the rotation
func (r *DynamoDBRotationRepository) AdvanceRotation(ctx context.Context) error {
	rotation, err := r.GetRotation(ctx)
	if err != nil {
		return err
	}
	if rotation == nil {
		return fmt.Errorf("no rotation found to advance")
	}

	// Find the current owner in the rotation order
	currentIndex := -1
	for i, owner := range rotation.RotationOrder {
		if owner == rotation.CurrentOwner {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		log.Warn().
			Str("currentOwner", rotation.CurrentOwner).
			Msg("Current owner not found in rotation order, resetting to first person")
		currentIndex = 0
	}

	// Advance to the next person in the rotation
	nextIndex := (currentIndex + 1) % len(rotation.RotationOrder)
	rotation.CurrentOwner = rotation.RotationOrder[nextIndex]
	rotation.LastRotationDate = time.Now()
	
	// Calculate next rotation date based on frequency
	// This is a simplified implementation
	rotation.NextRotationDate = calculateNextRotationDate(rotation.LastRotationDate, rotation.Frequency)

	// Save the updated rotation
	item, err := attributevalue.MarshalMap(rotation)
	if err != nil {
		return fmt.Errorf("failed to marshal rotation: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	}

	_, err = r.client.PutItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update rotation: %w", err)
	}

	return nil
}

// Helper function to calculate the next rotation date based on frequency
func calculateNextRotationDate(lastDate time.Time, frequency string) time.Time {
	switch frequency {
	case "daily":
		return lastDate.AddDate(0, 0, 1) //nolint:mnd
	case "weekly":
		return lastDate.AddDate(0, 0, 7) //nolint:mnd
	case "biweekly":
		return lastDate.AddDate(0, 0, 14) //nolint:mnd
	case "monthly":
		return lastDate.AddDate(0, 1, 0) //nolint:mnd
	default:
		// Default to weekly if frequency is unknown
		return lastDate.AddDate(0, 0, 7) //nolint:mnd
	}
}
