#!/bin/bash

# Define endpoint and credentials for local DynamoDB
export AWS_ACCESS_KEY_ID=dummy
export AWS_SECRET_ACCESS_KEY=dummy
export AWS_REGION=us-east-1
export AWS_PAGER=""
ENDPOINT_URL="http://localhost:8000"

echo "Checking if 'Videos' table exists..."

# Check if the table already exists
if aws dynamodb describe-table --table-name Videos --endpoint-url $ENDPOINT_URL --output json > /dev/null 2>&1; then
    echo "Table 'Videos' already exists."
else
    echo "Creating 'Videos' table..."
    aws dynamodb create-table \
        --endpoint-url $ENDPOINT_URL \
        --region $AWS_REGION \
        --table-name Videos \
        --attribute-definitions AttributeName=PK,AttributeType=S \
        --key-schema AttributeName=PK,KeyType=HASH \
        --billing-mode PAY_PER_REQUEST \
        --output json
    
    echo "Table 'Videos' created successfully!"
fi
