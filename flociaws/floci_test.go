package flociaws_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	floci "github.com/floci-io/testcontainers-floci-go/flociaws"
	"github.com/testcontainers/testcontainers-go"
)

const testImage = "floci/floci:latest"

func TestRun_DefaultConfig(t *testing.T) {
	ctx := context.Background()

	container, err := floci.Run(ctx, testImage)
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("starting container: %v", err)
	}

	if container.GetEndpoint() == "" {
		t.Error("expected non-empty endpoint")
	}
	if container.GetRegion() != floci.DefaultRegion {
		t.Errorf("expected region %q, got %q", floci.DefaultRegion, container.GetRegion())
	}
	if container.GetAccessKey() != floci.DefaultAccessKey {
		t.Errorf("expected access key %q, got %q", floci.DefaultAccessKey, container.GetAccessKey())
	}
	if container.GetSecretKey() != floci.DefaultSecretKey {
		t.Errorf("expected secret key %q, got %q", floci.DefaultSecretKey, container.GetSecretKey())
	}
	if container.GetAccountID() != floci.DefaultAccountID {
		t.Errorf("expected account ID %q, got %q", floci.DefaultAccountID, container.GetAccountID())
	}
	t.Logf("endpoint: %s", container.GetEndpoint())
}

func TestRun_CustomRegion(t *testing.T) {
	ctx := context.Background()

	container, err := floci.Run(ctx, testImage, floci.WithRegion("eu-west-1"))
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("starting container: %v", err)
	}

	if container.GetRegion() != "eu-west-1" {
		t.Errorf("expected region %q, got %q", "eu-west-1", container.GetRegion())
	}
}

func TestRun_DedicatedNetwork(t *testing.T) {
	ctx := context.Background()

	container, err := floci.Run(ctx, testImage, floci.WithDedicatedNetwork())
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("starting container: %v", err)
	}

	if container.GetDedicatedNetworkName() == "" {
		t.Error("expected non-empty dedicated network name")
	}
	t.Logf("network: %s", container.GetDedicatedNetworkName())
}

func TestRun_ServiceConfigs(t *testing.T) {
	ctx := context.Background()

	container, err := floci.Run(ctx, testImage,
		floci.WithS3Config(floci.S3Config{
			Enabled:                     true,
			DefaultPresignExpirySeconds: 7200,
		}),
		floci.WithSqsConfig(floci.SqsConfig{
			Enabled:                  true,
			DefaultVisibilityTimeout: 60,
			MaxMessageSize:           131072,
		}),
		floci.WithDynamoDbConfig(floci.DynamoDbConfig{Enabled: true}),
	)
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("starting container: %v", err)
	}

	t.Logf("endpoint: %s", container.GetEndpoint())
}

// Generic testcontainers options compose with Floci options, and the returned
// Container exposes the embedded testcontainers.Container API.
func TestRun_GenericCustomizer(t *testing.T) {
	ctx := context.Background()

	container, err := floci.Run(ctx, testImage,
		floci.WithRegion("eu-west-1"),
		testcontainers.WithEnv(map[string]string{
			"FLOCI_TEST_MARKER":        "from-generic-option",
			"FLOCI_DEFAULT_ACCOUNT_ID": "111122223333", // overrides the module default
		}),
	)
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("starting container: %v", err)
	}

	inspect, err := container.Inspect(ctx)
	if err != nil {
		t.Fatalf("inspecting container: %v", err)
	}
	want := map[string]bool{
		"FLOCI_TEST_MARKER=from-generic-option": false,
		"FLOCI_DEFAULT_REGION=eu-west-1":        false,
	}
	for _, env := range inspect.Config.Env {
		if _, ok := want[env]; ok {
			want[env] = true
		}
	}
	for env, found := range want {
		if !found {
			t.Errorf("expected container env %q", env)
		}
	}
	// The getters report what the emulator actually runs with, including generic overrides.
	if got := container.GetAccountID(); got != "111122223333" {
		t.Errorf("GetAccountID() = %q, want the WithEnv override %q", got, "111122223333")
	}
	if got := container.GetRegion(); got != "eu-west-1" {
		t.Errorf("GetRegion() = %q, want %q", got, "eu-west-1")
	}
}

// The deprecated builder API keeps working on top of Run.
func TestDeprecatedBuilder_StillStarts(t *testing.T) {
	ctx := context.Background()

	//nolint:staticcheck // exercising the deprecated API on purpose
	container, err := floci.NewFlociContainer().WithRegion("eu-west-2").Start(ctx)
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("starting container: %v", err)
	}
	if container.GetRegion() != "eu-west-2" {
		t.Errorf("expected region %q, got %q", "eu-west-2", container.GetRegion())
	}
}

// Reset wipes emulator state through the core's reset endpoint; the container keeps running
// and serves new requests.
func TestRun_Reset(t *testing.T) {
	ctx := context.Background()

	container, err := floci.Run(ctx, testImage)
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("starting container: %v", err)
	}
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(container.GetRegion()),
		config.WithBaseEndpoint(container.GetEndpoint()),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			container.GetAccessKey(), container.GetSecretKey(), "")))
	if err != nil {
		t.Fatalf("aws config: %v", err)
	}
	client := sqs.NewFromConfig(cfg)
	if _, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{QueueName: aws.String("before-reset")}); err != nil {
		t.Fatalf("create queue: %v", err)
	}

	if err := container.Reset(ctx); err != nil {
		t.Fatalf("reset: %v", err)
	}

	queues, err := client.ListQueues(ctx, &sqs.ListQueuesInput{})
	if err != nil {
		t.Fatalf("list queues after reset: %v", err)
	}
	if len(queues.QueueUrls) != 0 {
		t.Fatalf("queues survived the reset: %v", queues.QueueUrls)
	}
	if _, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{QueueName: aws.String("after-reset")}); err != nil {
		t.Fatalf("create queue after reset: %v", err)
	}
}

// terminateRecorder wraps a container and records that its own Terminate ran.
type terminateRecorder struct {
	testcontainers.Container
	called bool
}

func (w *terminateRecorder) Terminate(ctx context.Context, opts ...testcontainers.TerminateOption) error {
	w.called = true
	return w.Container.Terminate(ctx, opts...)
}

// A wrapper assigned to the embedded Container after Run is the one Terminate goes through, so
// its own cleanup runs.
func TestRun_TerminateGoesThroughAssignedWrapper(t *testing.T) {
	ctx := context.Background()

	container, err := floci.Run(ctx, testImage)
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("starting container: %v", err)
	}
	wrapper := &terminateRecorder{Container: container.Container}
	container.Container = wrapper

	if _, err := container.GetMappedPort(ctx, 4566); err != nil {
		t.Fatalf("mapped port through the wrapper: %v", err)
	}
	if err := container.Terminate(ctx); err != nil {
		t.Fatalf("terminate: %v", err)
	}
	if !wrapper.called {
		t.Fatal("Terminate skipped the assigned wrapper")
	}
}
