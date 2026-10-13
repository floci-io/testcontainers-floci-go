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
open-source local cloud emulators. The `flociaws` package starts a Floci (AWS) container for your integration tests and
gives you an endpoint and credentials to point the AWS SDK for Go v2 at, plus typed, per-service configuration structs
over the emulator's environment variables. The `flociaz` package does the same for Floci Azure, with connection
helpers for the Azure SDK for Go, and `flocigcp` for Floci GCP, with the emulator variables the Google Cloud client
libraries read. No cloud account, no auth token.

See the [Floci documentation](https://floci.io/floci/services/) for the full list of supported AWS services.

### The Floci emulators

testcontainers-floci-go is the Go member of the [Floci](https://github.com/floci-io) Testcontainers family. Floci is
named after [floccus](https://en.wikipedia.org/wiki/Cirrocumulus_floccus), the cloud formation that looks like popcorn.

| Emulator                                           | Cloud | Port | Supported                                                                                  |
|----------------------------------------------------|-------|:----:|:------------------------------------------------------------------------------------------:|
| [floci](https://github.com/floci-io/floci)         | AWS   | 4566 | ✅ [`floci` package](https://pkg.go.dev/github.com/floci-io/testcontainers-floci-go)        |
| [floci-az](https://github.com/floci-io/floci-az)   | Azure | 4577 | ✅ [`flociaz` package](https://pkg.go.dev/github.com/floci-io/testcontainers-floci-go/flociaz) |
| [floci-gcp](https://github.com/floci-io/floci-gcp) | GCP   | 4588 | ✅ [`flocigcp` package](https://pkg.go.dev/github.com/floci-io/testcontainers-floci-go/flocigcp) |
| [floci-oci](https://github.com/floci-io/floci-oci) | OCI   | 4599 | Planned                                                                                    |

## Installation

```bash
go get github.com/floci-io/testcontainers-floci-go
```

The module path is `github.com/floci-io/testcontainers-floci-go`. Each cloud is a package of that one module: `flociaws`
for AWS, `flociaz` for Azure, `flocigcp` for GCP:

```go
import "github.com/floci-io/testcontainers-floci-go/flociaws"
import "github.com/floci-io/testcontainers-floci-go/flociaz"
import "github.com/floci-io/testcontainers-floci-go/flocigcp"
```

The root package (`floci`, imported from the module path itself) is the original import path. It still
works, as deprecated aliases of `flociaws`, so existing code keeps compiling; move to `flociaws` at your own
pace. The remaining Floci cloud will arrive as a sibling package (`flocioci`).

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

    "github.com/floci-io/testcontainers-floci-go/flociaws"
    "github.com/testcontainers/testcontainers-go"
)

func TestS3(t *testing.T) {
    ctx := context.Background()

    fc, err := flociaws.Run(ctx, "floci/floci:latest")
    testcontainers.CleanupContainer(t, fc)
    if err != nil {
        t.Fatal(err)
    }

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

    "github.com/floci-io/testcontainers-floci-go/flociaws"
    "github.com/testcontainers/testcontainers-go"
)

var fc *flociaws.Container

func TestMain(m *testing.M) {
    ctx := context.Background()
    var err error
    fc, err = flociaws.Run(ctx, "floci/floci:latest")
    if err != nil {
        panic(err)
    }
    code := m.Run()
    _ = testcontainers.TerminateContainer(fc)
    os.Exit(code)
}
```

### Examples

- [`examples/s3`](examples/s3/) — create a bucket, upload documents, list objects
- [`examples/dynamodb`](examples/dynamodb/) — DynamoDB tables and items
- [`examples/sqs`](examples/sqs/) — queues, send and receive messages
- [`examples/sns`](examples/sns/) — topics and subscriptions
- [`examples/lambda`](examples/lambda/) — deploy and invoke a function

### Azure

The `flociaz` package starts the Floci Azure emulator (`floci/floci-az`) and follows the same conventions as
`flociaws`: `Run`, package-level options, generic testcontainers options, `Reset`, a unique resource namespace and
sibling-container cleanup on `Terminate`.

```go
package myservice_test

import (
    "context"
    "testing"

    "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"

    "github.com/floci-io/testcontainers-floci-go/flociaz"
    "github.com/testcontainers/testcontainers-go"
)

func TestBlob(t *testing.T) {
    ctx := context.Background()

    fc, err := flociaz.Run(ctx, "floci/floci-az:latest")
    testcontainers.CleanupContainer(t, fc)
    if err != nil {
        t.Fatal(err)
    }

    client, err := azblob.NewClientFromConnectionString(fc.GetStorageConnectionString(), nil)
    if err != nil {
        t.Fatal(err)
    }
    if _, err := client.CreateContainer(ctx, "documents", nil); err != nil {
        t.Fatal(err)
    }
    if _, err := client.UploadBuffer(ctx, "documents", "hello.txt", []byte("hello from floci"), nil); err != nil {
        t.Fatal(err)
    }
}
```

### GCP

The `flocigcp` package starts the Floci GCP emulator (`floci/floci-gcp`), again with the same conventions.
`GetEmulatorEnv()` returns the `*_EMULATOR_HOST` variables that point the Google Cloud client libraries at the
container; set them before creating a client. No credentials are needed.

```go
package myservice_test

import (
    "context"
    "testing"

    "cloud.google.com/go/pubsub"

    "github.com/floci-io/testcontainers-floci-go/flocigcp"
    "github.com/testcontainers/testcontainers-go"
)

func TestPubSub(t *testing.T) {
    ctx := context.Background()

    fc, err := flocigcp.Run(ctx, "floci/floci-gcp:latest")
    testcontainers.CleanupContainer(t, fc)
    if err != nil {
        t.Fatal(err)
    }
    for name, value := range fc.GetEmulatorEnv() {
        t.Setenv(name, value) // PUBSUB_EMULATOR_HOST, STORAGE_EMULATOR_HOST, GOOGLE_CLOUD_PROJECT, ...
    }

    client, err := pubsub.NewClient(ctx, fc.GetProjectID())
    if err != nil {
        t.Fatal(err)
    }
    defer client.Close()
    // ...
}
```

Services without an emulator variable (BigQuery, Cloud Tasks, Cloud KMS, ...) take the endpoint explicitly, e.g.
`option.WithEndpoint(fc.GetEndpoint())` with `option.WithoutAuthentication()`.

## Service configuration

Each AWS service emulated by Floci can be configured individually using a typed config struct. Pass the struct to the
corresponding `With*Config` method; unset fields keep their defaults. See the
[Floci documentation](https://floci.io/floci/services/) for the full list of supported services.

### Per-service examples

#### S3

```go
fc, err := flociaws.Run(ctx, "floci/floci:latest",
    flociaws.WithS3Config(flociaws.S3Config{
        Enabled:                     true,
        DefaultPresignExpirySeconds: 7200,
    }),
)
```

#### SQS

```go
fc, err := flociaws.Run(ctx, "floci/floci:latest",
    flociaws.WithSqsConfig(flociaws.SqsConfig{
        Enabled:                  true,
        DefaultVisibilityTimeout: 60,
        MaxMessageSize:           262144,
    }),
)
```

#### DynamoDB

```go
fc, err := flociaws.Run(ctx, "floci/floci:latest",
    flociaws.WithDynamoDbConfig(flociaws.DynamoDbConfig{Enabled: true}),
)
```

#### Lambda

```go
fc, err := flociaws.Run(ctx, "floci/floci:latest",
    flociaws.WithDedicatedNetwork(), // required for Lambda to reach Floci
    flociaws.WithLambdaConfig(flociaws.LambdaConfig{
        Enabled:               true,
        DefaultMemoryMb:       256,
        DefaultTimeoutSeconds: 30,
        HotReloadEnabled:      true,
        ExposeRuntimePorts:    true, // invoke Lambdas from the host
    }),
)
```

#### RDS (PostgreSQL / MySQL / MariaDB)

```go
fc, err := flociaws.Run(ctx, "floci/floci:latest",
    flociaws.WithDedicatedNetwork(),
    flociaws.WithRdsConfig(flociaws.RdsConfig{
        Enabled:              true,
        DefaultPostgresImage: "postgres:16-alpine",
        ExposeProxyPorts:     true, // reach the databases from the host
    }),
)
```

The module sets `FLOCI_SERVICES_RDS_ENDPOINT_HOST` to the Docker host, so `DescribeDBInstances` returns an
endpoint (host and published port) that a client on the host can connect to, on Linux, macOS and Windows alike.
Set `RdsConfig.EndpointHost` to advertise another hostname.

#### ElastiCache (Redis / Valkey)

```go
fc, err := flociaws.Run(ctx, "floci/floci:latest",
    flociaws.WithDedicatedNetwork(),
    flociaws.WithElastiCacheConfig(flociaws.ElastiCacheConfig{
        Enabled:      true,
        DefaultImage: "valkey/valkey:8",
    }),
)
```

#### OpenSearch

```go
fc, err := flociaws.Run(ctx, "floci/floci:latest",
    flociaws.WithDedicatedNetwork(),
    flociaws.WithOpenSearchConfig(flociaws.OpenSearchConfig{
        Enabled: true,
        Mock:    false,
    }),
)
```

#### MSK (Kafka via Redpanda)

```go
fc, err := flociaws.Run(ctx, "floci/floci:latest",
    flociaws.WithDedicatedNetwork(),
    flociaws.WithMskConfig(flociaws.MskConfig{
        Enabled:      true,
        DefaultImage: "redpandadata/redpanda:latest",
    }),
)
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

### Azure

Every Azure service has a config struct with an `Enabled` switch, and a `Mocked` switch where the emulator has one;
pass it with the matching `flociaz.With*Config` option. `flociaz.Default*Config()` returns the emulator's defaults.
More typed settings will follow service by service; until then set them with `testcontainers.WithEnv`
(`FLOCI_AZ_SERVICES_<SERVICE>_<SETTING>`). The service list is the emulator's
[`ServicesConfig`](https://github.com/floci-io/floci-az/blob/main/src/main/java/io/floci/az/config/EmulatorConfig.java),
one `<Service>Config` struct each; see the
[package reference](https://pkg.go.dev/github.com/floci-io/testcontainers-floci-go/flociaz) for all of them.

```go
fc, err := flociaz.Run(ctx, "floci/floci-az:latest",
    flociaz.WithRedisConfig(flociaz.RedisConfig{Enabled: true, Mocked: true}), // no Valkey container
    flociaz.WithServiceBusConfig(flociaz.ServiceBusConfig{Enabled: true}),      // a real broker instead of the mock
    flociaz.WithAksConfig(flociaz.AksConfig{Enabled: false}),
)
```

Services that spawn sibling containers (ACR, AKS, Event Hubs, Functions, the MariaDB, MySQL and PostgreSQL databases,
Redis) need the host Docker socket, and so do Service Bus, Container Instances, Container Apps and Virtual Machines
once their `Mocked` switch is turned off (they are mocked by default), Cosmos DB with a MongoDB, PostgreSQL, Cassandra
or Gremlin engine enabled, and SQL Database with the `managed` data plane. The module mounts the socket only while one
of them needs it; `flociaz.WithDockerSocket(bool)` overrides the detection. The module also sets
`FLOCI_AZ_DOCKER_ENDPOINT_MODE=published`, so the addresses the emulator reports for those containers are reachable
from the test process, and `FLOCI_AZ_TLS_HTTPS_PORT=0`.

### GCP

Every GCP service has a config struct with an `Enabled` switch, and a `Mock` switch where the emulator has one;
pass it with the matching `flocigcp.With*Config` option. `flocigcp.Default*Config()` returns the emulator's defaults.
More typed settings will follow service by service; until then set them with `testcontainers.WithEnv`
(`FLOCI_GCP_SERVICES_<SERVICE>_<SETTING>`). The service list is the emulator's
[`ServicesConfig`](https://github.com/floci-io/floci-gcp/blob/main/src/main/java/io/floci/gcp/config/EmulatorConfig.java),
one `<Service>Config` struct each; see the
[package reference](https://pkg.go.dev/github.com/floci-io/testcontainers-floci-go/flocigcp) for all of them.

```go
fc, err := flocigcp.Run(ctx, "floci/floci-gcp:latest",
    flocigcp.WithCloudSqlConfig(flocigcp.CloudSqlConfig{Enabled: true, Mock: true}), // no database container
    flocigcp.WithBigQueryConfig(flocigcp.BigQueryConfig{Enabled: true, Mock: true}), // built-in SQL subset, no floci-duck
    flocigcp.WithGkeConfig(flocigcp.GkeConfig{Enabled: false}),
)
```

Services that spawn sibling containers (BigQuery, Cloud Run, Cloud SQL, GKE, Managed Kafka) need the host Docker
socket unless their `Mock` switch is on; BigQuery does not need it either when `DuckURL` names a floci-duck that is
already running. The module mounts the socket only while one of them needs it; `flocigcp.WithDockerSocket(bool)`
overrides the detection. The module also sets `FLOCI_GCP_TLS_HTTPS_PORT=0`; with `FLOCI_GCP_TLS_ENABLED=true` the
emulator serves HTTPS on the same port, and `Run` waits for that listener too. The default project is `floci-local`; the
emulator accepts any project ID, and `flocigcp.WithProjectID("my-project")` changes the default one, which
`GetProjectID()` then returns.

## Container options

```go
fc, err := flociaws.Run(ctx, "floci/floci:latest", // image: pin a specific tag here
    flociaws.WithRegion("eu-west-1"),
    flociaws.WithAccountID("111122223333"),
    flociaws.WithAvailabilityZone("eu-west-1a"),
    flociaws.WithDedicatedNetwork(),            // isolated Docker network for stateful services
)
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
| `GetResourceNamespace()` | Unique prefix (`tc-…`) of the sibling containers' names, so parallel runs never collide |
| `Reset(ctx)` | Wipes all emulator state (buckets, queues, tables, …) without restarting the container |

### Dedicated network

Services that spawn real Docker containers (Lambda, RDS, ElastiCache, MSK, OpenSearch, ECR, EKS) need a Docker network to communicate with Floci. Pass `flociaws.WithDedicatedNetwork()` to have the module create and manage one automatically:

```go
fc, err := flociaws.Run(ctx, "floci/floci:latest",
    flociaws.WithDedicatedNetwork(),
    flociaws.WithLambdaConfig(flociaws.LambdaConfig{Enabled: true}),
)

// The network name is passed to Floci automatically via FLOCI_SERVICES_DOCKER_NETWORK.
// fc.GetDedicatedNetworkName() returns it if you need it elsewhere.
```

The network is removed when the container is terminated (`testcontainers.CleanupContainer` or `Terminate`).

### Sibling containers

Floci manages the containers it spawns (databases, Lambda runtimes, the ECR registry, …); the testcontainers
reaper does not track them. `Terminate` (and so `testcontainers.CleanupContainer`) removes every one of them after
stopping Floci, found by the `floci_namespace` label that holds the container's resource namespace. Containers that
share a namespace (set through `testcontainers.WithEnv` on `FLOCI_DOCKER_RESOURCE_NAMESPACE`) lose their
siblings together.

### Docker socket

Services that spawn sibling containers (Lambda, RDS, ElastiCache, ECS, EC2, EKS, ECR, MSK, OpenSearch, Athena,
CodeBuild, Neptune) need the host Docker socket. The module mounts it only while at least one of them is enabled and not in
`Mock` mode; all services are enabled by default, so a default container gets it. Override the detection with
`flociaws.WithDockerSocket(false)` on hosts where the socket cannot be mounted (rootless Podman with SELinux, some CI
sandboxes), or `flociaws.WithDockerSocket(true)` to always mount it.

### Generic testcontainers options

`Run` accepts any `testcontainers.ContainerCustomizer`, so Floci options and the standard testcontainers-go options
compose. Generic options are applied after the module defaults and can override them:

```go
fc, err := flociaws.Run(ctx, "floci/floci:latest",
    flociaws.WithRegion("eu-west-1"),
    testcontainers.WithEnv(map[string]string{"FLOCI_LOG_LEVEL": "debug"}),
)
```

`*flociaws.Container` embeds `testcontainers.Container`, so `Logs`, `Exec`, `Inspect` and the cleanup helpers work on it.

### Migrating from the builder API

`NewFlociContainer()...Start(ctx)` still works but is deprecated. Each builder method has a package-level option of
the same name:

| Before | After |
|---|---|
| `flociaws.NewFlociContainer().WithRegion("eu-west-1").Start(ctx)` | `flociaws.Run(ctx, "floci/floci:latest", flociaws.WithRegion("eu-west-1"))` |
| `flociaws.Run(ctx, func(c *flociaws.FlociContainer) { c.WithS3Config(cfg) })` | `flociaws.Run(ctx, "floci/floci:latest", flociaws.WithS3Config(cfg))` |
| `.WithImage("floci/floci:x.y.z")` | the `img` argument of `Run` |
| `*flociaws.StartedFlociContainer` | `*flociaws.Container` (the old name is a deprecated alias) |
| `fc.Stop(ctx)` | `testcontainers.CleanupContainer(t, fc)` in tests, or `fc.Terminate(ctx)` |

### Azure

`*flociaz.Container` gives the connection details of the Azure emulator:

| Method | Returns |
|---|---|
| `GetEndpoint()` | `http://host:port`, the base URL of every Azure service |
| `GetHTTPSEndpoint()` | `https://host:port`, served on the same port once TLS is enabled (`FLOCI_AZ_TLS_ENABLED=true`); `Run` then also waits for the HTTPS listener |
| `GetTLSCertificate(ctx)` | PEM certificate to trust for the HTTPS endpoint; an error while TLS is disabled |
| `GetSubscriptionID()` | Subscription ID (`00000000-0000-0000-0000-000000000001`) |
| `GetTenantID()` | Microsoft Entra tenant ID (`00000000-0000-0000-0000-000000000002`) |
| `GetAccountName()` | Storage account name (`devstoreaccount1`) |
| `GetAccountKey()` | Storage account key (the Azurite development key) |
| `GetBlobEndpoint()` | Blob Storage endpoint, `<endpoint>/devstoreaccount1` |
| `GetQueueEndpoint()` | Queue Storage endpoint, `<endpoint>/devstoreaccount1-queue` |
| `GetTableEndpoint()` | Table Storage endpoint, `<endpoint>/devstoreaccount1-table` |
| `GetStorageConnectionString()` | Connection string with the three storage endpoints, for the Azure Storage SDKs |
| `GetDedicatedNetworkName()` | Docker network name (empty if none) |
| `GetMappedPort(ctx, port)` | Host port mapped from the given container port |
| `GetResourceNamespace()` | Unique prefix (`tc-…`) of the sibling containers' names |
| `Reset(ctx)` | Wipes all emulator state without restarting the container |

### GCP

`*flocigcp.Container` gives the connection details of the GCP emulator:

| Method | Returns |
|---|---|
| `GetEndpoint()` | `http://host:port`, the base URL of every GCP service (REST and gRPC on the one port) |
| `GetProjectID()` | Default project ID (`floci-local`, or the one `WithProjectID` set) |
| `GetEmulatorEnv()` | Map of the emulator variables, below |
| `GetDedicatedNetworkName()` | Docker network name (empty if none) |
| `GetMappedPort(ctx, port)` | Host port mapped from the given container port |
| `GetResourceNamespace()` | Unique prefix (`tc-…`) of the sibling containers' names |
| `Reset(ctx)` | Wipes all emulator state without restarting the container |

| Variable in `GetEmulatorEnv()` | Value |
|---|---|
| `PUBSUB_EMULATOR_HOST`, `FIRESTORE_EMULATOR_HOST`, `DATASTORE_EMULATOR_HOST`, `SECRET_MANAGER_EMULATOR_HOST`, `FIREBASE_AUTH_EMULATOR_HOST` | `host:port` |
| `STORAGE_EMULATOR_HOST` | `http://host:port` |
| `STORAGE_EMULATOR_HOST_GRPC` | `host:port` |
| `LOGGING_EMULATOR_HOST`, `KMS_EMULATOR_HOST` | `host:port`; no Google library reads these two, they are for code that builds the client endpoint from them, as the emulator's own compatibility tests do |
| `GOOGLE_CLOUD_PROJECT` | The project ID |

## Docker image tags

By default the module runs the floating `latest` tag of the emulator image (`floci/floci:latest`), so you always test
against the current emulator. Pass a different image to `Run` to pin a release or follow `main`:

```go
flociaws.Run(ctx, "floci/floci:x.y.z")   // a specific release
flociaws.Run(ctx, "floci/floci:nightly") // built from main every night
```

The Azure and GCP modules work the same way with `floci/floci-az` and `floci/floci-gcp`
(`flociaz.Run(ctx, "floci/floci-az:x.y.z")`, `flocigcp.Run(ctx, "floci/floci-gcp:x.y.z")`).

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

Everything except the `*_internal_test.go` files starts real Floci containers, so Docker must be running; the
images are pulled automatically on first run.
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
