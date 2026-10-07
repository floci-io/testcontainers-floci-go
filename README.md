<p align="center">
  <img src="https://raw.githubusercontent.com/floci-io/.github/main/floci.svg#gh-light-mode-only" alt="Floci" width="500" />
  <img src="https://github.com/user-attachments/assets/edfff8b3-926c-471e-9549-77fb90a21b49#gh-dark-mode-only" alt="Floci" width="500" />
</p>

<p align="center">
  <strong>Any Cloud. Locally.</strong><br />
  Light, fluffy, and always free: Testcontainers for Go<br />
  No account. No auth token. No feature gates.
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/floci-io/testcontainers-floci-go"><img src="https://pkg.go.dev/badge/github.com/floci-io/testcontainers-floci-go.svg" alt="Go Reference"></a>
  <a href="https://github.com/floci-io/testcontainers-floci-go/actions/workflows/ci.yml"><img src="https://github.com/floci-io/testcontainers-floci-go/actions/workflows/ci.yml/badge.svg?branch=main" alt="CI"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/license-MIT-green" alt="License: MIT"></a>
  <a href="https://github.com/floci-io/testcontainers-floci-go/stargazers"><img src="https://img.shields.io/github/stars/floci-io/testcontainers-floci-go?style=flat" alt="GitHub Stars"></a>
</p>

<p align="center">
  <a href="#quick-start">Quick Start</a> ·
  <a href="#service-configuration">Configuration</a> ·
  <a href="#the-floci-emulators">Emulators</a> ·
  <a href="https://floci.io/floci/testcontainers/go/">Docs</a>
</p>

---

## What is this?

A [Testcontainers for Go](https://golang.testcontainers.org) module for [Floci](https://github.com/floci-io), the free,
open-source local cloud emulators. The `floci` package starts a Floci (AWS) container for your integration tests and
gives you an endpoint and credentials to point the AWS SDK for Go v2 at, plus typed, per-service configuration structs
over the emulator's environment variables. No cloud account, no auth token.

See the [Floci documentation](https://floci.io/floci/services/) for the full list of supported AWS services.

### The Floci emulators

testcontainers-floci-go is the Go member of the [Floci](https://github.com/floci-io) Testcontainers family. Floci is
named after [floccus](https://en.wikipedia.org/wiki/Cirrocumulus_floccus), the cloud formation that looks like popcorn.

| Emulator                                           | Cloud | Port | Supported                                                                                  |
|----------------------------------------------------|-------|:----:|:------------------------------------------------------------------------------------------:|
| [floci](https://github.com/floci-io/floci)         | AWS   | 4566 | ✅ [`floci` package](https://pkg.go.dev/github.com/floci-io/testcontainers-floci-go)        |
| [floci-az](https://github.com/floci-io/floci-az)   | Azure | 4577 | Planned                                                                                    |
| [floci-gcp](https://github.com/floci-io/floci-gcp) | GCP   | 4588 | Planned                                                                                    |
| [floci-oci](https://github.com/floci-io/floci-oci) | OCI   | 4599 | Planned                                                                                    |

## Installation

```bash
go get github.com/floci-io/testcontainers-floci-go
```

The module path is `github.com/floci-io/testcontainers-floci-go`; the package name is `floci`.

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

### Sharing a container across tests

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

### Examples

- [`examples/s3`](examples/s3/) — create a bucket, upload documents, list objects
- [`examples/dynamodb`](examples/dynamodb/) — DynamoDB tables and items
- [`examples/sqs`](examples/sqs/) — queues, send and receive messages
- [`examples/sns`](examples/sns/) — topics and subscriptions
- [`examples/lambda`](examples/lambda/) — deploy and invoke a function

## Service configuration

Each AWS service emulated by Floci can be configured individually using a typed config struct. Pass the struct to the
corresponding `With*Config` method; unset fields keep their defaults. See the
[Floci documentation](https://floci.io/floci/services/) for the full list of supported services.

### Per-service examples

#### S3

```go
fc, _ := floci.NewFlociContainer().
    WithS3Config(floci.S3Config{
        Enabled:                     true,
        DefaultPresignExpirySeconds: 7200,
    }).
    Start(ctx)
```

#### SQS

```go
fc, _ := floci.NewFlociContainer().
    WithSqsConfig(floci.SqsConfig{
        Enabled:                  true,
        DefaultVisibilityTimeout: 60,
        MaxMessageSize:           262144,
    }).
    Start(ctx)
```

#### DynamoDB

```go
fc, _ := floci.NewFlociContainer().
    WithDynamoDbConfig(floci.DynamoDbConfig{Enabled: true}).
    Start(ctx)
```

#### Lambda

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

#### RDS (PostgreSQL / MySQL / MariaDB)

```go
fc, _ := floci.NewFlociContainer().
    WithDedicatedNetwork().
    WithRdsConfig(floci.RdsConfig{
        Enabled:              true,
        DefaultPostgresImage: "postgres:16-alpine",
    }).
    Start(ctx)
```

#### ElastiCache (Redis / Valkey)

```go
fc, _ := floci.NewFlociContainer().
    WithDedicatedNetwork().
    WithElastiCacheConfig(floci.ElastiCacheConfig{
        Enabled:      true,
        DefaultImage: "valkey/valkey:8",
    }).
    Start(ctx)
```

#### OpenSearch

```go
fc, _ := floci.NewFlociContainer().
    WithDedicatedNetwork().
    WithOpenSearchConfig(floci.OpenSearchConfig{
        Enabled: true,
        Mock:    false,
    }).
    Start(ctx)
```

#### MSK (Kafka via Redpanda)

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

### Dedicated network

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

## Docker image tags

By default the module runs the floating `latest` tag of the emulator image (`floci/floci:latest`), so you always test
against the current emulator. Use `WithImage` to pin a release or follow `main`:

```go
floci.NewFlociContainer().WithImage("floci/floci:x.y.z")   // a specific release
floci.NewFlociContainer().WithImage("floci/floci:nightly") // built from main every night
```

Every emulator publishes `latest`, `x.y.z` and `nightly` tags. The AWS emulator also publishes a compat variant:

| Tag | Description |
|---|---|
| `floci/floci:latest` | Native image (default, recommended) |
| `floci/floci:x.y.z` | Pinned release |
| `floci/floci:nightly` | Latest nightly build from `main` |
| `floci/floci:latest-compat` | Includes Python 3, AWS CLI, and boto3 |

## Requirements

- Go 1.25+
- Docker (running locally or in CI)
- [testcontainers-go](https://github.com/testcontainers/testcontainers-go) (version pinned in `go.mod`)

## Building and testing

```bash
go build ./... && go vet ./...            # compile and vet all packages
go test ./...                             # all tests
go test -v -run TestRun_DefaultConfig .   # a single test
```

Everything except `ports_internal_test.go` starts real Floci containers, so Docker must be running; the
`floci/floci:latest` image is pulled automatically on first run.
See [CONTRIBUTING.md](CONTRIBUTING.md) for the branching model and how to add a service.

## Other languages

| Language | Repository |
|---|---|
| Java | [testcontainers-floci](https://github.com/floci-io/testcontainers-floci) |
| Node.js / TypeScript | [testcontainers-floci-node](https://github.com/floci-io/testcontainers-floci-node) |
| Python | [testcontainers-floci-python](https://github.com/floci-io/testcontainers-floci-python) |
| Go | **testcontainers-floci-go** (this repo) |
| .NET | [testcontainers-floci-dotnet](https://github.com/floci-io/testcontainers-floci-dotnet) |

## Community

- 💬 [Slack](https://join.slack.com/t/floci/shared_invite/zt-3tjn02s3q-A00kEjJ1cZxsg_imTfy6Cw): quick questions and community chat
- 🗣️ [GitHub Discussions](https://github.com/orgs/floci-io/discussions): ideas, design tradeoffs, and proposals
- [CONTRIBUTING.md](CONTRIBUTING.md) · [SECURITY.md](SECURITY.md) · [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) · [MAINTAINERS.md](MAINTAINERS.md)

## License

MIT. See [LICENSE](LICENSE).

---

<div align="center">

Floci™ is a trademark of Hector Ventura. Code is MIT-licensed; see
[TRADEMARK.md](https://github.com/floci-io/.github/blob/main/TRADEMARK.md) for name and logo use.

</div>
