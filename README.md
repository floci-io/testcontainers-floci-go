# testcontainers-floci-go

[![Go Reference](https://pkg.go.dev/badge/github.com/floci-io/testcontainers-floci-go.svg)](https://pkg.go.dev/github.com/floci-io/testcontainers-floci-go)
[![CI](https://github.com/floci-io/testcontainers-floci-go/actions/workflows/ci.yml/badge.svg)](https://github.com/floci-io/testcontainers-floci-go/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Go [Testcontainers](https://testcontainers.com) module for [Floci](https://github.com/floci-io/floci) — the open-source, drop-in replacement for LocalStack Community Edition.

Floci emulates **42 AWS services** in a single container with:
- **~24 ms** startup time (native image)
- **~13 MiB** idle memory
- **~90 MB** Docker image
- No auth tokens, no feature gates, MIT license

## Installation

```bash
go get github.com/floci-io/testcontainers-floci-go
```

Requires Go 1.25+ and a running Docker daemon.

## Quick start

```go
package myservice_test

import (
    "context"
    "strings"
    "testing"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/credentials"
    "github.com/aws/aws-sdk-go-v2/service/s3"

    floci "github.com/floci-io/testcontainers-floci-go"
)

func TestS3(t *testing.T) {
    ctx := context.Background()

    fc, err := floci.NewFlociContainer().Start(ctx)
    if err != nil {
        t.Fatal(err)
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
        t.Fatal(err)
    }

    client := s3.NewFromConfig(cfg, func(o *s3.Options) {
        o.UsePathStyle = true // required for local endpoints
    })

    _, err = client.CreateBucket(ctx, &s3.CreateBucketInput{
        Bucket: aws.String("my-bucket"),
    })
    if err != nil {
        t.Fatal(err)
    }

    _, err = client.PutObject(ctx, &s3.PutObjectInput{
        Bucket: aws.String("my-bucket"),
        Key:    aws.String("hello.txt"),
        Body:   strings.NewReader("hello from floci"),
    })
    if err != nil {
        t.Fatal(err)
    }

    out, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
        Bucket: aws.String("my-bucket"),
    })
    if err != nil {
        t.Fatal(err)
    }

    t.Logf("objects: %d", len(out.Contents))
}
```

> **S3 note:** always use `strings.NewReader` or `bytes.NewReader` (seekable) when uploading objects.
> `bytes.NewBufferString` is not seekable and causes the AWS SDK to attempt trailing checksums,
> which require TLS and fail against a plain HTTP local endpoint.

## Sharing a container across tests

Use `TestMain` to start the container once for the whole package:

```go
package myservice_test

import (
    "context"
    "os"
    "testing"

    floci "github.com/floci-io/testcontainers-floci-go"
)

var fc *floci.StartedFlociContainer

func TestMain(m *testing.M) {
    ctx := context.Background()
    var err error
    fc, err = floci.NewFlociContainer().Start(ctx)
    if err != nil {
        panic(err)
    }
    code := m.Run()
    _ = fc.Stop(ctx)
    os.Exit(code)
}
```

## Service configuration

Each of Floci's 42 services can be configured individually using typed config structs. Pass any struct to the corresponding `With*Config` method — unset fields keep their defaults.

### S3

```go
fc, _ := floci.NewFlociContainer().
    WithS3Config(floci.S3Config{
        Enabled:                     true,
        DefaultPresignExpirySeconds: 7200,
    }).
    Start(ctx)
```

### SQS

```go
fc, _ := floci.NewFlociContainer().
    WithSqsConfig(floci.SqsConfig{
        Enabled:                  true,
        DefaultVisibilityTimeout: 60,
        MaxMessageSize:           262144,
    }).
    Start(ctx)
```

### DynamoDB

```go
fc, _ := floci.NewFlociContainer().
    WithDynamoDbConfig(floci.DynamoDbConfig{Enabled: true}).
    Start(ctx)
```

### Lambda

```go
fc, _ := floci.NewFlociContainer().
    WithDedicatedNetwork(). // required for Lambda to reach Floci
    WithLambdaConfig(floci.LambdaConfig{
        Enabled:               true,
        DefaultMemoryMb:       256,
        DefaultTimeoutSeconds: 30,
        HotReloadEnabled:      true,
        ExposeRuntimePorts:    true, // invoke Lambdas from the host
    }).
    Start(ctx)
```

### RDS (PostgreSQL / MySQL / MariaDB)

```go
fc, _ := floci.NewFlociContainer().
    WithDedicatedNetwork().
    WithRdsConfig(floci.RdsConfig{
        Enabled:              true,
        DefaultPostgresImage: "postgres:16-alpine",
    }).
    Start(ctx)
```

### ElastiCache (Redis / Valkey)

```go
fc, _ := floci.NewFlociContainer().
    WithDedicatedNetwork().
    WithElastiCacheConfig(floci.ElastiCacheConfig{
        Enabled:      true,
        DefaultImage: "valkey/valkey:8",
    }).
    Start(ctx)
```

### OpenSearch

```go
fc, _ := floci.NewFlociContainer().
    WithDedicatedNetwork().
    WithOpenSearchConfig(floci.OpenSearchConfig{
        Enabled: true,
        Mock:    false,
    }).
    Start(ctx)
```

### MSK (Kafka via Redpanda)

```go
fc, _ := floci.NewFlociContainer().
    WithDedicatedNetwork().
    WithMskConfig(floci.MskConfig{
        Enabled:      true,
        DefaultImage: "redpandadata/redpanda:latest",
    }).
    Start(ctx)
```

### All available config structs

| Struct | AWS service |
|---|---|
| `AcmConfig` | AWS Certificate Manager |
| `ApiGatewayConfig` | API Gateway (v1) |
| `ApiGatewayV2Config` | API Gateway (v2) |
| `AppConfigConfig` | AppConfig |
| `AppConfigDataConfig` | AppConfig Data |
| `AthenaConfig` | Athena |
| `BedrockRuntimeConfig` | Bedrock Runtime |
| `CloudFormationConfig` | CloudFormation |
| `CloudWatchLogsConfig` | CloudWatch Logs |
| `CloudWatchMetricsConfig` | CloudWatch Metrics |
| `CodeBuildConfig` | CodeBuild |
| `CodeDeployConfig` | CodeDeploy |
| `CognitoConfig` | Cognito |
| `DynamoDbConfig` | DynamoDB |
| `Ec2Config` | EC2 |
| `EcrConfig` | ECR |
| `EcsConfig` | ECS |
| `EksConfig` | EKS |
| `ElastiCacheConfig` | ElastiCache |
| `ElbV2Config` | ELB v2 |
| `EventBridgeConfig` | EventBridge |
| `FirehoseConfig` | Kinesis Firehose |
| `GlueConfig` | Glue |
| `IamConfig` | IAM |
| `KinesisConfig` | Kinesis |
| `KmsConfig` | KMS |
| `LambdaConfig` | Lambda |
| `MskConfig` | MSK (Kafka) |
| `OpenSearchConfig` | OpenSearch |
| `PipesConfig` | EventBridge Pipes |
| `RdsConfig` | RDS |
| `ResourceGroupsTaggingConfig` | Resource Groups Tagging |
| `S3Config` | S3 |
| `SchedulerConfig` | EventBridge Scheduler |
| `SecretsManagerConfig` | Secrets Manager |
| `SesConfig` | SES |
| `SesV2Config` | SES v2 |
| `SnsConfig` | SNS |
| `SqsConfig` | SQS |
| `SsmConfig` | SSM Parameter Store |
| `StepFunctionsConfig` | Step Functions |

## Container options

```go
fc, _ := floci.NewFlociContainer().
    WithImage("floci/floci:latest").   // pin a specific tag
    WithRegion("eu-west-1").
    WithAccountID("111122223333").
    WithAvailabilityZone("eu-west-1a").
    WithDedicatedNetwork().            // isolated Docker network for stateful services
    Start(ctx)
```

### Connection details

| Method | Returns |
|---|---|
| `GetEndpoint()` | `http://host:port` — pass as base endpoint to AWS SDK clients |
| `GetRegion()` | AWS region string |
| `GetAccessKey()` | Access key (`"test"`) |
| `GetSecretKey()` | Secret key (`"test"`) |
| `GetAccountID()` | AWS account ID |
| `GetAvailabilityZone()` | Availability zone |
| `GetDedicatedNetworkName()` | Docker network name (empty if none) |
| `GetMappedPort(ctx, port)` | Host port mapped from the given container port |

## Dedicated network

Services that spawn real Docker containers (Lambda, RDS, ElastiCache, MSK, OpenSearch, ECR, EKS) need a Docker network to communicate with Floci. Call `WithDedicatedNetwork()` to have the module create and manage one automatically:

```go
fc, _ := floci.NewFlociContainer().
    WithDedicatedNetwork().
    WithLambdaConfig(floci.LambdaConfig{Enabled: true}).
    Start(ctx)

// The network name is passed to Floci automatically via FLOCI_SERVICES_DOCKER_NETWORK.
// fc.GetDedicatedNetworkName() returns it if you need it elsewhere.
```

The network is removed when `Stop` is called.

## Docker image variants

| Tag | Description |
|---|---|
| `floci/floci:latest` | Native image — sub-second startup (recommended) |
| `floci/floci:x.y.z` | Pinned release |
| `floci/floci:latest-compat` | Includes Python 3, AWS CLI, and boto3 |
| `floci/floci:nightly` | Latest nightly build from `main` |

## Requirements

- Go 1.25+
- Docker (running locally or in CI)
- `github.com/testcontainers/testcontainers-go v0.42.0`

## Examples

- [`examples/s3`](examples/s3/) — create a bucket, upload documents, list objects
- [`examples/dynamodb`](examples/dynamodb/) — DynamoDB tables and items
- [`examples/sqs`](examples/sqs/) — queues, send and receive messages
- [`examples/sns`](examples/sns/) — topics and subscriptions
- [`examples/lambda`](examples/lambda/) — deploy and invoke a function

## Running the tests

```bash
go test -v ./...
```

Requires Docker running locally; the `floci/floci:latest` image is pulled automatically on first run.

## Related projects

- [Floci](https://github.com/floci-io/floci) — the emulator itself
- [testcontainers-floci](https://github.com/floci-io/testcontainers-floci) — Java / Spring Boot module
- [testcontainers-floci-node](https://github.com/floci-io/testcontainers-floci-node) — Node.js module
- [testcontainers-floci-python](https://github.com/floci-io/testcontainers-floci-python) — Python module
- [Testcontainers for Go](https://golang.testcontainers.org)

## License

MIT
