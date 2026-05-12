# Example: SQS

Demonstrates using `testcontainers-floci-go` to test SQS operations against a local [Floci](https://floci.io) instance.

The test:
1. Starts a Floci container
2. Creates an SQS queue
3. Sends a message
4. Receives and verifies the message body
5. Deletes the message from the queue

## Run

```bash
go test -v -timeout 120s ./examples/sqs/...
```

## Code walkthrough

```go
// Start Floci
fc, err := floci.NewFlociContainer().Start(ctx)
defer fc.Stop(ctx)

// Wire up the AWS SDK SQS client
cfg, _ := config.LoadDefaultConfig(ctx,
    config.WithRegion(fc.GetRegion()),
    config.WithBaseEndpoint(fc.GetEndpoint()),
    config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
        fc.GetAccessKey(), fc.GetSecretKey(), "",
    )),
)
client := sqs.NewFromConfig(cfg)
```
