package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/zerolog/log"
)

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

// GetRotation retrieves a rotation by team ID and rotation label.
func (r *DynamoDBRotationRepository) GetRotation(ctx context.Context, teamID, rotationLabel string) (*Rotation, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"creator_team_id": &types.AttributeValueMemberS{Value: teamID},
			"rotation_label":  &types.AttributeValueMemberS{Value: rotationLabel},
		},
	}

	result, err := r.client.GetItem(ctx, input)

	if err != nil {
		return nil, fmt.Errorf("failed to get rotation: %w", err)
	}

	if result.Item == nil {
		return nil, nil // rotation not found
	}

	var rotation Rotation
	if err := attributevalue.UnmarshalMap(result.Item, &rotation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rotation: %w", err)
	}

	return &rotation, nil
}

// AdvanceRotation moves to the next person in the rotation.
func (r *DynamoDBRotationRepository) AdvanceRotation(ctx context.Context, teamID, rotationLabel string) error {
	rotation, err := r.GetRotation(ctx, teamID, rotationLabel)

	if err != nil {
		return err
	}

	if rotation == nil {
		return fmt.Errorf("rotation not found for team %s and label %s", teamID, rotationLabel)
	}

	// Find the current owner in the rotation order.
	currentIndex := -1

	for i, owner := range rotation.RotationOrder {
		if owner == rotation.CurrentOwner {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		log.Warn().Str("currentOwner", rotation.CurrentOwner).
			Msg("Current owner not found in rotation order, resetting to first person")

		currentIndex = 0
	}

	// Advance to the next person.
	nextIndex := (currentIndex + 1) % len(rotation.RotationOrder)

	rotation.CurrentOwner = rotation.RotationOrder[nextIndex]
	rotation.LastRotationDate = time.Now()
	rotation.NextRotationDate = calculateNextRotationDate(rotation.LastRotationDate, rotation.Frequency)

	// Save the updated rotation.
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

// CreateRotation creates a new rotation record.
func (r *DynamoDBRotationRepository) UpsertRotation(ctx context.Context, rotation Rotation) error {
	// Determine if this is a new rotation or an update.
	var input dynamodb.PutItemInput

	input.TableName = aws.String(r.tableName)

	if rotation.Version == 0 {
		// New rotation: initialize version to 1 and condition that item does not exist.
		rotation.Version = 1
		input.ConditionExpression = aws.String("attribute_not_exists(creator_team_id) AND attribute_not_exists(rotation_label)")
	} else {
		// Update: enforce version check (optimistic concurrency).
		currentVersion := rotation.Version
		// Increment version for the new update.
		rotation.Version = currentVersion + 1
		input.ConditionExpression = aws.String("version = :v")
		input.ExpressionAttributeValues = map[string]types.AttributeValue{
			":v": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", currentVersion)},
		}
	}

	item, err := attributevalue.MarshalMap(rotation)
	if err != nil {
		return fmt.Errorf("failed to marshal rotation: %w", err)
	}

	input.Item = item

	_, err = r.client.PutItem(ctx, &input)
	if err != nil {
		return fmt.Errorf("failed to update/create rotation: %w", err)
	}

	return nil
}

// ListRotations lists all rotations for the provided team.
func (r *DynamoDBRotationRepository) ListRotations(ctx context.Context, teamID string, limit int32, lastEvaluatedKey map[string]types.AttributeValue) ([]Rotation, map[string]types.AttributeValue, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("creator_team_id = :teamID"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":teamID": &types.AttributeValueMemberS{Value: teamID},
		},
		Limit:             aws.Int32(limit),
		ExclusiveStartKey: lastEvaluatedKey,
	}

	output, err := r.client.Query(ctx, input)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query rotations: %w", err)
	}

	var rotations []Rotation
	if err := attributevalue.UnmarshalListOfMaps(output.Items, &rotations); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal rotations: %w", err)
	}

	return rotations, output.LastEvaluatedKey, nil
}

// Helper to calculate the next rotation date.
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
		return lastDate.AddDate(0, 0, 7) //nolint:mnd
	}
}
