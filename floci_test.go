package floci_test

import (
	"context"
	"testing"

	floci "github.com/floci-io/testcontainers-floci-go"
)

func TestRun_DefaultConfig(t *testing.T) {
	ctx := context.Background()

	container, err := floci.Run(ctx)
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

func TestRun_CustomRegion(t *testing.T) {
	ctx := context.Background()

	container, err := floci.Run(ctx, func(c *floci.FlociContainer) {
		c.WithRegion("eu-west-1")
	})
	if err != nil {
		t.Fatalf("starting container: %v", err)
	}
	t.Cleanup(func() { _ = container.Stop(ctx) })

	if container.GetRegion() != "eu-west-1" {
		t.Errorf("expected region %q, got %q", "eu-west-1", container.GetRegion())
	}
}

func TestRun_DedicatedNetwork(t *testing.T) {
	ctx := context.Background()

	container, err := floci.Run(ctx, func(c *floci.FlociContainer) {
		c.WithDedicatedNetwork()
	})
	if err != nil {
		t.Fatalf("starting container: %v", err)
	}
	t.Cleanup(func() { _ = container.Stop(ctx) })

	if container.GetDedicatedNetworkName() == "" {
		t.Error("expected non-empty dedicated network name")
	}
	t.Logf("network: %s", container.GetDedicatedNetworkName())
}

func TestRun_ServiceConfigs(t *testing.T) {
	ctx := context.Background()

	container, err := floci.Run(ctx, func(c *floci.FlociContainer) {
		c.WithS3Config(floci.S3Config{
			Enabled:                     true,
			DefaultPresignExpirySeconds: 7200,
		})
		c.WithSqsConfig(floci.SqsConfig{
			Enabled:                  true,
			DefaultVisibilityTimeout: 60,
			MaxMessageSize:           131072,
		})
		c.WithDynamoDbConfig(floci.DynamoDbConfig{Enabled: true})
	})
	if err != nil {
		t.Fatalf("starting container: %v", err)
	}
	t.Cleanup(func() { _ = container.Stop(ctx) })

	t.Logf("endpoint: %s", container.GetEndpoint())
}
