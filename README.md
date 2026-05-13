# testcontainers-floci-go

A [Testcontainers](https://testcontainers.com) module for [Floci](https://floci.io) — a free, open-source local AWS emulator.

## Requirements

- Go 1.25+ (current latest; required by `testcontainers-go v0.42.0` — if you need Go 1.22/1.23/1.24 support, pin an older version of this module)
- Docker

## Installation

```sh
go get github.com/floci-io/testcontainers-floci-go
```

## Quickstart

```go
package myservice_test

import (
    "context"
    "testing"

    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/credentials"
    floci "github.com/floci-io/testcontainers-floci-go"
)

func TestMyService(t *testing.T) {
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

    // Use cfg to build any aws-sdk-go-v2 client (S3, SQS, DynamoDB, etc.)
    _ = cfg
}
```

## Configuration

All services are enabled by default. Use `With<Service>Config` to tune or disable specific services:

```go
fc, err := floci.NewFlociContainer().
    WithRegion("eu-west-1").
    WithS3Config(floci.S3Config{
        Enabled:                     true,
        DefaultPresignExpirySeconds: 7200,
    }).
    WithSqsConfig(floci.SqsConfig{
        Enabled:                  true,
        DefaultVisibilityTimeout: 60,
        MaxMessageSize:           131072,
    }).
    Start(ctx)
```

### Container-based services (Lambda, RDS, ElastiCache, etc.)

Services that spin up their own Docker containers need a shared network:

```go
fc, err := floci.NewFlociContainer().
    WithDedicatedNetwork().
    WithLambdaConfig(floci.LambdaConfig{
        Enabled:            true,
        ExposeRuntimePorts: true,
    }).
    Start(ctx)
```

## Supported services

| Service | Config type |
|---|---|
| ACM | `AcmConfig` |
| API Gateway | `ApiGatewayConfig` / `ApiGatewayV2Config` |
| AppConfig | `AppConfigConfig` / `AppConfigDataConfig` |
| Athena | `AthenaConfig` |
| Bedrock Runtime | `BedrockRuntimeConfig` |
| CloudFormation | `CloudFormationConfig` |
| CloudWatch Logs | `CloudWatchLogsConfig` |
| CloudWatch Metrics | `CloudWatchMetricsConfig` |
| CodeBuild | `CodeBuildConfig` |
| CodeDeploy | `CodeDeployConfig` |
| Cognito | `CognitoConfig` |
| DynamoDB | `DynamoDbConfig` |
| EC2 | `Ec2Config` |
| ECR | `EcrConfig` |
| ECS | `EcsConfig` |
| EKS | `EksConfig` |
| ElastiCache | `ElastiCacheConfig` |
| ELBv2 | `ElbV2Config` |
| EventBridge | `EventBridgeConfig` |
| Firehose | `FirehoseConfig` |
| Glue | `GlueConfig` |
| IAM | `IamConfig` |
| Kinesis | `KinesisConfig` |
| KMS | `KmsConfig` |
| Lambda | `LambdaConfig` |
| MSK | `MskConfig` |
| OpenSearch | `OpenSearchConfig` |
| Pipes | `PipesConfig` |
| RDS | `RdsConfig` |
| Resource Groups Tagging | `ResourceGroupsTaggingConfig` |
| S3 | `S3Config` |
| Scheduler | `SchedulerConfig` |
| Secrets Manager | `SecretsManagerConfig` |
| SES | `SesConfig` / `SesV2Config` |
| SNS | `SnsConfig` |
| SQS | `SqsConfig` |
| SSM | `SsmConfig` |
| Step Functions | `StepFunctionsConfig` |

## Examples

- [S3](examples/s3/s3_test.go)
- [DynamoDB](examples/dynamodb/dynamodb_test.go)
- [SQS](examples/sqs/sqs_test.go)
- [SNS](examples/sns/sns_test.go)
- [Lambda](examples/lambda/lambda_test.go)

## Running the tests

```sh
go test -v ./...
```

> Requires Docker running locally and the `floci/floci:latest` image available (pulled automatically on first run).

## Reference

- Java module: [testcontainers-floci](https://github.com/floci-io/testcontainers-floci)
- Floci documentation: [floci.io](https://floci.io)
