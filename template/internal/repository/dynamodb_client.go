package repository

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// NewDynamoDBClient creates a new DynamoDB client
func NewDynamoDBClient(endpoint, region string, isLocal bool) (*dynamodb.Client, error) {
	ctx := context.Background()
	var cfg aws.Config
	var err error

	if isLocal {
		// Local DynamoDB configuration
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) { //nolint:staticcheck
			return aws.Endpoint{ //nolint:staticcheck
				PartitionID:   "aws",
				URL:           endpoint,
				SigningRegion: region,
			}, nil
		})

		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithEndpointResolverWithOptions(customResolver), //nolint:staticcheck
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("dummy", "dummy", "dummy")),
		)
	} else {
		// AWS DynamoDB configuration
		cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(region))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return dynamodb.NewFromConfig(cfg), nil
}
