package repository

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDBRotationScheduleRepo struct {
	client    *dynamodb.Client
	tableName string
}

var _ RotationScheduleRepository = (*DynamoDBRotationScheduleRepo)(nil)

func NewDynamoDBRotationScheduleRepo(client *dynamodb.Client, tableName string) *DynamoDBRotationScheduleRepo {
	return &DynamoDBRotationScheduleRepo{
		client:    client,
		tableName: tableName,
	}
}

func (r *DynamoDBRotationScheduleRepo) AddEvent(ctx context.Context, event RotationScheduleEvent) error {
	item, err := attributevalue.MarshalMap(event)
	if err != nil {
		return err
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	return err
}

func (r *DynamoDBRotationScheduleRepo) GetEventsByHour(ctx context.Context, hour string) ([]RotationScheduleEvent, error) {
	result, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("scheduled_hour = :hour"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":hour": &types.AttributeValueMemberS{Value: hour},
		},
	})

	if err != nil {
		return nil, err
	}

	var events []RotationScheduleEvent
	err = attributevalue.UnmarshalListOfMaps(result.Items, &events)
	return events, err
}
