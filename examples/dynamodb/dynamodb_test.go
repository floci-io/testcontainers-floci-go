package dynamodb_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	floci "github.com/floci-io/testcontainers-floci-go"
)

func TestDynamoDBExample(t *testing.T) {
	ctx := context.Background()

	fc, err := floci.NewFlociContainer().Start(ctx)
	if err != nil {
		t.Fatalf("starting floci: %v", err)
	}
	t.Cleanup(func() { _ = fc.Stop(ctx) })

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(fc.GetRegion()),
		config.WithBaseEndpoint(fc.GetEndpoint()),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			fc.GetAccessKey(), fc.GetSecretKey(), "",
		)),
	)
	if err != nil {
		t.Fatalf("loading AWS config: %v", err)
	}

	client := dynamodb.NewFromConfig(cfg)

	table := "users"

	_, err = client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(table),
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("id"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("id"), KeyType: types.KeyTypeHash},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		t.Fatalf("creating table: %v", err)
	}
	t.Logf("created table: %s", table)

	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(table),
		Item: map[string]types.AttributeValue{
			"id":    &types.AttributeValueMemberS{Value: "user-1"},
			"name":  &types.AttributeValueMemberS{Value: "Alice"},
			"email": &types.AttributeValueMemberS{Value: "alice@example.com"},
		},
	})
	if err != nil {
		t.Fatalf("putting item: %v", err)
	}
	t.Log("put item user-1")

	result, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: "user-1"},
		},
	})
	if err != nil {
		t.Fatalf("getting item: %v", err)
	}

	name, ok := result.Item["name"].(*types.AttributeValueMemberS)
	if !ok {
		t.Fatal("expected name attribute to be a string")
	}
	if name.Value != "Alice" {
		t.Errorf("expected name %q, got %q", "Alice", name.Value)
	}
	t.Logf("retrieved item: id=user-1 name=%s", name.Value)
}
