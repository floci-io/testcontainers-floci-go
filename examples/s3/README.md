# Example: S3

Demonstrates using `testcontainers-floci-go` to test S3 operations against a local [Floci](https://floci.io) instance.

The test:
1. Starts a Floci container
2. Creates an S3 bucket
3. Uploads two documents (`hello.txt`, `world.txt`)
4. Lists objects in the bucket
5. Downloads and verifies the content of `hello.txt`

## Run

```bash
go test -v -timeout 300s ./examples/s3/...
```

## Code walkthrough

```go
// Start Floci
fc, err := floci.NewFlociContainer().Start(ctx)
defer fc.Stop(ctx)

// Wire up the AWS SDK S3 client
cfg, _ := config.LoadDefaultConfig(ctx,
    config.WithRegion(fc.GetRegion()),
    config.WithBaseEndpoint(fc.GetEndpoint()),
    config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
        fc.GetAccessKey(), fc.GetSecretKey(), "",
    )),
)
client := s3.NewFromConfig(cfg, func(o *s3.Options) {
    o.UsePathStyle = true // required for local endpoints
})
```

> `UsePathStyle = true` is required when pointing the AWS SDK at a local endpoint.
