package floci_test

import (
	"context"
	"testing"

	floci "github.com/floci-io/testcontainers-floci-go"
)

func TestFlociContainer_DefaultConfig(t *testing.T) {
	ctx := context.Background()

	container, err := floci.NewFlociContainer().Start(ctx)
	if err != nil {
		t.Fatalf("starting container: %v", err)
	}
	t.Cleanup(func() {
		if err := container.Stop(ctx); err != nil {
			t.Errorf("stopping container: %v", err)
		}
	})

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

func TestFlociContainer_CustomRegion(t *testing.T) {
	ctx := context.Background()

	container, err := floci.NewFlociContainer().
		WithRegion("eu-west-1").
		Start(ctx)
	if err != nil {
		t.Fatalf("starting container: %v", err)
	}
	t.Cleanup(func() { _ = container.Stop(ctx) })

	if container.GetRegion() != "eu-west-1" {
		t.Errorf("expected region %q, got %q", "eu-west-1", container.GetRegion())
	}
}

func TestFlociContainer_DedicatedNetwork(t *testing.T) {
	ctx := context.Background()

	container, err := floci.NewFlociContainer().
		WithDedicatedNetwork().
		Start(ctx)
	if err != nil {
		t.Fatalf("starting container: %v", err)
	}
	t.Cleanup(func() { _ = container.Stop(ctx) })

	if container.GetDedicatedNetworkName() == "" {
		t.Error("expected non-empty dedicated network name")
	}
	t.Logf("network: %s", container.GetDedicatedNetworkName())
}

func TestFlociContainer_ServiceConfigs(t *testing.T) {
	ctx := context.Background()

	container, err := floci.NewFlociContainer().
		WithS3Config(floci.S3Config{
			Enabled:                     true,
			DefaultPresignExpirySeconds: 7200,
		}).
		WithSqsConfig(floci.SqsConfig{
			Enabled:                  true,
			DefaultVisibilityTimeout: 60,
			MaxMessageSize:           131072,
		}).
		WithDynamoDbConfig(floci.DynamoDbConfig{Enabled: true}).
		Start(ctx)
	if err != nil {
		t.Fatalf("starting container: %v", err)
	}
	t.Cleanup(func() { _ = container.Stop(ctx) })

	t.Logf("endpoint: %s", container.GetEndpoint())
}
