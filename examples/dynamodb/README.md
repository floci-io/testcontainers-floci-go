# Example: DynamoDB

Demonstrates using `testcontainers-floci-go` to test DynamoDB operations against a local [Floci](https://floci.io) instance.

The test:
1. Starts a Floci container
2. Creates a `users` table (pay-per-request billing)
3. Inserts an item via `PutItem`
4. Retrieves it via `GetItem` and verifies the attribute value

## Run

```bash
go test -v -timeout 300s ./examples/dynamodb/...
```

## Code walkthrough

```go
// Start Floci
fc, err := floci.NewFlociContainer().Start(ctx)
defer fc.Stop(ctx)

// Wire up the AWS SDK DynamoDB client
cfg, _ := config.LoadDefaultConfig(ctx,
    config.WithRegion(fc.GetRegion()),
    config.WithBaseEndpoint(fc.GetEndpoint()),
    config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
        fc.GetAccessKey(), fc.GetSecretKey(), "",
    )),
)
client := dynamodb.NewFromConfig(cfg)
```
