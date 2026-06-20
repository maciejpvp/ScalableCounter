package database

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// VideoRepository provides data access for Video items in DynamoDB
type VideoRepository struct {
	db *DB
}

// NewVideoRepository creates a new VideoRepository
func NewVideoRepository(db *DB) *VideoRepository {
	return &VideoRepository{
		db: db,
	}
}

// VideoRecord represents an item in our DynamoDB table
type VideoRecord struct {
	VideoID string `dynamodbav:"PK" json:"video_id"` // Partition Key changed to PK
	Likes   int    `dynamodbav:"likes" json:"likes"`
}

// PutVideo inserts or replaces a video item
func (r *VideoRepository) PutVideo(ctx context.Context, videoID string) (*VideoRecord, error) {
	record := VideoRecord{
		VideoID: fmt.Sprintf("VIDEO#%s", videoID),
		Likes:   0,
	}

	item, err := attributevalue.MarshalMap(record)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record: %w", err)
	}

	_, err = r.db.DynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("Videos"), // Replace with your actual table name
		Item:      item,
	})
	if err != nil {
		return nil, err
	}
	
	return &record, nil
}

// GetVideo reads a video item by its key
func (r *VideoRepository) GetVideo(ctx context.Context, videoID string) (*VideoRecord, error) {
	prefixedID := fmt.Sprintf("VIDEO#%s", videoID)
	// Key query updated to use "PK"
	key, err := attributevalue.MarshalMap(map[string]string{"PK": prefixedID})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal key: %w", err)
	}

	result, err := r.db.DynamoClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String("Videos"),
		Key:       key,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	if result.Item == nil {
		return nil, nil // Item not found
	}

	var record VideoRecord
	err = attributevalue.UnmarshalMap(result.Item, &record)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal item: %w", err)
	}

	return &record, nil
}

func (r *VideoRepository) IncrementVideoLikes(ctx context.Context, videoID string, likes int) error {
	prefixedID := fmt.Sprintf("VIDEO#%s", videoID)
	key, err := attributevalue.MarshalMap(map[string]string{"PK": prefixedID})
	if err != nil {
		return fmt.Errorf("failed to marshal key: %w", err)
	}

	updateResult, err := r.db.DynamoClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String("Videos"),
		Key:       key,
		UpdateExpression: aws.String("ADD likes :inc"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":inc": &types.AttributeValueMemberN{Value: strconv.Itoa(likes)},
		},
		ReturnValues: types.ReturnValueUpdatedNew,
	})

	if err != nil {
		return fmt.Errorf("failed to increment likes: %w", err)
	}

	var updatedData struct {
		Likes int `dynamodbav:"likes"`
	}
	err = attributevalue.UnmarshalMap(updateResult.Attributes, &updatedData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal updated item: %w", err)
	}

	return nil
}