# Example: SNS

Demonstrates using `testcontainers-floci-go` to test SNS→SQS fan-out against a local [Floci](https://floci.io) instance.

The test:
1. Starts a Floci container
2. Creates an SNS topic
3. Creates an SQS queue and subscribes it to the topic
4. Publishes a message to the topic
5. Receives the message from SQS and verifies the SNS notification payload

## Run

```bash
go test -v -timeout 300s ./examples/sns/...
```

## Code walkthrough

```go
// Start Floci
fc, err := floci.NewFlociContainer().Start(ctx)
defer fc.Stop(ctx)

// Wire up SNS and SQS clients from the same config
cfg, _ := config.LoadDefaultConfig(ctx,
    config.WithRegion(fc.GetRegion()),
    config.WithBaseEndpoint(fc.GetEndpoint()),
    config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
        fc.GetAccessKey(), fc.GetSecretKey(), "",
    )),
)
snsClient := sns.NewFromConfig(cfg)
sqsClient := sqs.NewFromConfig(cfg)
```
