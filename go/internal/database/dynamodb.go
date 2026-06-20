package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DB struct {
	DynamoClient *dynamodb.Client
}

func NewDB(ctx context.Context) (*DB, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}
	client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		endpoint := os.Getenv("DYNAMODB_ENDPOINT")
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
			o.Credentials = aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
				return aws.Credentials{
					AccessKeyID:     "dummy",
					SecretAccessKey: "dummy",
				}, nil
			})
			o.Region = "us-east-1"
		}
	})

	db := &DB{DynamoClient: client}

	if err := db.ensureTables(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure tables exist: %w", err)
	}

	return db, nil
}

// ensureTables creates required DynamoDB tables if they don't already exist.
func (db *DB) ensureTables(ctx context.Context) error {
	_, err := db.DynamoClient.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String("Videos"),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("PK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("PK"),
				KeyType:       types.KeyTypeHash,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})

	if err != nil {
		// Ignore "table already exists" errors — that's fine.
		var resourceInUse *types.ResourceInUseException
		if errors.As(err, &resourceInUse) {
			log.Println("Table 'Videos' already exists, skipping creation.")
			return nil
		}
		return fmt.Errorf("failed to create Videos table: %w", err)
	}

	log.Println("Table 'Videos' created successfully.")
	return nil
}

// Health checks if the connection to DynamoDB is active
func (db *DB) Health(ctx context.Context) map[string]string {
	// We can try a simple operation like listing tables
	_, err := db.DynamoClient.ListTables(ctx, &dynamodb.ListTablesInput{
		Limit: aws.Int32(1),
	})

	status := "up"
	if err != nil {
		status = "down"
	}

	return map[string]string{
		"status": status,
		"type":   "dynamodb",
	}
}


