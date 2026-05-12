# Example: Lambda

Demonstrates using `testcontainers-floci-go` to test a Go Lambda function against a local [Floci](https://floci.io) instance.

The test:
1. Starts a Floci container with a dedicated Docker network (required for Lambda)
2. Compiles a minimal Go handler (`handler/main.go`) for Linux/amd64
3. Packages it into a zip with the `bootstrap` binary (required by `provided.al2023` runtime)
4. Creates the Lambda function via `CreateFunction`
5. Waits for the function to become Active
6. Invokes it and verifies the response payload

## Run

```bash
go test -v -timeout 120s ./examples/lambda/...
```

## Code walkthrough

```go
// Start Floci with dedicated network — required for Lambda container execution
fc, err := floci.NewFlociContainer().
    WithDedicatedNetwork().
    Start(ctx)
defer fc.Stop(ctx)

// Build the handler for the Lambda runtime
cmd := exec.CommandContext(ctx, "go", "build", "-o", binaryPath, "./handler")
cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64", "CGO_ENABLED=0")

// Wire up the AWS SDK Lambda client
cfg, _ := config.LoadDefaultConfig(ctx,
    config.WithRegion(fc.GetRegion()),
    config.WithBaseEndpoint(fc.GetEndpoint()),
    config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
        fc.GetAccessKey(), fc.GetSecretKey(), "",
    )),
)
client := lambda.NewFromConfig(cfg)
```

> **Note:** The handler is compiled for `linux/amd64`. On Apple Silicon the binary is built via Go's cross-compilation — no extra tooling needed since `CGO_ENABLED=0`.
