package flociaws

import (
	"testing"

	"github.com/moby/moby/api/types/container"
	"github.com/testcontainers/testcontainers-go"

	"github.com/floci-io/testcontainers-floci-go/internal/core"
)

// Option must satisfy testcontainers.ContainerCustomizer so it can be passed to Run
// alongside generic testcontainers options.
var _ testcontainers.ContainerCustomizer = Option(nil)

// disableSocketServices turns off every service that spawns sibling containers, leaving
// a container that should not need the Docker socket.
func disableSocketServices(c *FlociContainer) {
	c.WithAthenaConfig(AthenaConfig{})
	c.WithCodeBuildConfig(CodeBuildConfig{})
	c.WithEc2Config(Ec2Config{})
	c.WithEcrConfig(EcrConfig{})
	c.WithEcsConfig(EcsConfig{})
	c.WithEksConfig(EksConfig{})
	c.WithElastiCacheConfig(ElastiCacheConfig{})
	c.WithLambdaConfig(LambdaConfig{})
	c.WithMskConfig(MskConfig{})
	c.WithOpenSearchConfig(OpenSearchConfig{})
	c.WithRdsConfig(RdsConfig{})
}

func TestDockerSocket_DefaultsMountIt(t *testing.T) {
	// Every service is enabled by default, including the container-backed ones.
	if !newBuilder().needsDockerSocket(newBuilder().envVars) {
		t.Fatal("expected the default configuration to need the Docker socket")
	}
}

func TestDockerSocket_NotNeededWithoutContainerServices(t *testing.T) {
	c := newBuilder()
	disableSocketServices(c)
	if c.needsDockerSocket(c.envVars) {
		t.Fatal("expected no Docker socket once every container-backed service is disabled")
	}
}

func TestDockerSocket_EachServiceRequiresIt(t *testing.T) {
	tests := []struct {
		name   string
		enable func(c *FlociContainer)
	}{
		{"athena", func(c *FlociContainer) { c.WithAthenaConfig(AthenaConfig{Enabled: true}) }},
		{"codebuild", func(c *FlociContainer) { c.WithCodeBuildConfig(CodeBuildConfig{Enabled: true}) }},
		{"ec2", func(c *FlociContainer) { c.WithEc2Config(Ec2Config{Enabled: true}) }},
		{"ecr", func(c *FlociContainer) { c.WithEcrConfig(EcrConfig{Enabled: true}) }},
		{"ecs", func(c *FlociContainer) { c.WithEcsConfig(EcsConfig{Enabled: true}) }},
		{"eks", func(c *FlociContainer) { c.WithEksConfig(EksConfig{Enabled: true}) }},
		{"elasticache", func(c *FlociContainer) { c.WithElastiCacheConfig(ElastiCacheConfig{Enabled: true}) }},
		{"lambda", func(c *FlociContainer) { c.WithLambdaConfig(LambdaConfig{Enabled: true}) }},
		{"msk", func(c *FlociContainer) { c.WithMskConfig(MskConfig{Enabled: true}) }},
		{"opensearch", func(c *FlociContainer) { c.WithOpenSearchConfig(OpenSearchConfig{Enabled: true}) }},
		{"rds", func(c *FlociContainer) { c.WithRdsConfig(RdsConfig{Enabled: true}) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newBuilder()
			disableSocketServices(c)
			tt.enable(c)
			if !c.needsDockerSocket(c.envVars) {
				t.Fatalf("expected %s to require the Docker socket", tt.name)
			}
		})
	}
}

func TestDockerSocket_MockModeDoesNotRequireIt(t *testing.T) {
	tests := []struct {
		name   string
		enable func(c *FlociContainer)
	}{
		{"athena", func(c *FlociContainer) { c.WithAthenaConfig(AthenaConfig{Enabled: true, Mock: true}) }},
		{"ec2", func(c *FlociContainer) { c.WithEc2Config(Ec2Config{Enabled: true, Mock: true}) }},
		{"ecs", func(c *FlociContainer) { c.WithEcsConfig(EcsConfig{Enabled: true, Mock: true}) }},
		{"eks", func(c *FlociContainer) { c.WithEksConfig(EksConfig{Enabled: true, Mock: true}) }},
		{"msk", func(c *FlociContainer) { c.WithMskConfig(MskConfig{Enabled: true, Mock: true}) }},
		{"opensearch", func(c *FlociContainer) { c.WithOpenSearchConfig(OpenSearchConfig{Enabled: true, Mock: true}) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newBuilder()
			disableSocketServices(c)
			tt.enable(c)
			if c.needsDockerSocket(c.envVars) {
				t.Fatalf("expected mocked %s not to require the Docker socket", tt.name)
			}
		})
	}
}

func TestDockerSocket_OverrideWins(t *testing.T) {
	off := newBuilder().WithDockerSocket(false)
	if off.needsDockerSocket(off.envVars) {
		t.Error("WithDockerSocket(false) should disable the mount even with container-backed services enabled")
	}

	on := newBuilder()
	disableSocketServices(on)
	on.WithDockerSocket(true)
	if !on.needsDockerSocket(on.envVars) {
		t.Error("WithDockerSocket(true) should force the mount even with no container-backed service")
	}
}

// Package-level options apply the same settings as the builder methods they mirror.
func TestOptions_ApplyToBuilder(t *testing.T) {
	c := newBuilder()
	for _, opt := range []Option{
		WithRegion("eu-west-1"),
		WithAccountID("123456789012"),
		WithAvailabilityZone("eu-west-1b"),
		WithDedicatedNetwork(),
		WithDockerSocket(false),
		WithSqsConfig(SqsConfig{Enabled: false}),
	} {
		opt(c)
	}
	checks := map[string]string{
		"FLOCI_DEFAULT_REGION":            "eu-west-1",
		"FLOCI_DEFAULT_ACCOUNT_ID":        "123456789012",
		"FLOCI_DEFAULT_AVAILABILITY_ZONE": "eu-west-1b",
		"FLOCI_SERVICES_SQS_ENABLED":      "false",
	}
	for k, want := range checks {
		if got := c.envVars[k]; got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}
	if !c.dedicatedNetwork {
		t.Error("WithDedicatedNetwork option did not set the dedicated network")
	}
	if c.needsDockerSocket(c.envVars) {
		t.Error("WithDockerSocket(false) option did not override auto-detection")
	}
}

// buildRequest applies Run's customizers to an empty request without starting Docker,
// returning the request and the final environment Run would record.
func buildRequest(t *testing.T, c *FlociContainer, customizers ...testcontainers.ContainerCustomizer) (*testcontainers.GenericContainerRequest, map[string]string) {
	t.Helper()
	opts, finalEnv := c.requestOptions(nil, customizers)
	req := &testcontainers.GenericContainerRequest{}
	for _, opt := range opts {
		if err := opt.Customize(req); err != nil {
			t.Fatalf("customize: %v", err)
		}
	}
	return req, *finalEnv
}

func mountsDockerSocket(req *testcontainers.GenericContainerRequest) bool {
	if req.HostConfigModifier == nil {
		return false
	}
	hc := &container.HostConfig{}
	req.HostConfigModifier(hc)
	for _, b := range hc.Binds {
		if b == core.DockerSocket+":"+core.DockerSocket {
			return true
		}
	}
	return false
}

// A generic WithEnv that overrides an identity key must be what the getters report.
func TestRequest_GenericEnvOverridesIdentity(t *testing.T) {
	c := newBuilder()
	WithRegion("eu-west-1")(c)
	_, env := buildRequest(t, c, testcontainers.WithEnv(map[string]string{
		"FLOCI_DEFAULT_ACCOUNT_ID":        "111122223333",
		"FLOCI_DEFAULT_REGION":            "ap-southeast-2",
		"FLOCI_DEFAULT_AVAILABILITY_ZONE": "ap-southeast-2b",
	}))
	for key, want := range map[string]string{
		"FLOCI_DEFAULT_ACCOUNT_ID":        "111122223333",
		"FLOCI_DEFAULT_REGION":            "ap-southeast-2",
		"FLOCI_DEFAULT_AVAILABILITY_ZONE": "ap-southeast-2b",
	} {
		if got := envOr(env, key, ""); got != want {
			t.Errorf("final %s = %q, want %q", key, got, want)
		}
	}
}

// Enabling a container-backed service through a generic env override must mount the socket.
func TestRequest_GenericEnvEnablingServiceMountsSocket(t *testing.T) {
	c := newBuilder()
	disableSocketServices(c)
	req, _ := buildRequest(t, c, testcontainers.WithEnv(map[string]string{"FLOCI_SERVICES_LAMBDA_ENABLED": "true"}))
	if !mountsDockerSocket(req) {
		t.Fatal("expected the socket once Lambda is enabled through testcontainers.WithEnv")
	}
}

// Disabling every container-backed service through a generic env override must drop the socket.
func TestRequest_GenericEnvDisablingServicesDropsSocket(t *testing.T) {
	env := map[string]string{}
	for _, svc := range awsDescriptor.SocketServices {
		env["FLOCI_SERVICES_"+svc.Token+"_ENABLED"] = "false"
	}
	req, _ := buildRequest(t, newBuilder(), testcontainers.WithEnv(env))
	if mountsDockerSocket(req) {
		t.Fatal("expected no socket once every container-backed service is disabled through testcontainers.WithEnv")
	}
}

// The explicit override still wins over the final environment.
func TestRequest_OverrideWinsOverGenericEnv(t *testing.T) {
	c := newBuilder()
	WithDockerSocket(false)(c)
	req, _ := buildRequest(t, c, testcontainers.WithEnv(map[string]string{"FLOCI_SERVICES_LAMBDA_ENABLED": "true"}))
	if mountsDockerSocket(req) {
		t.Fatal("WithDockerSocket(false) must win over an env override")
	}
}

// A caller's WithHostConfigModifier is chained, not replaced, when the socket is mounted.
func TestRequest_CallerHostConfigModifierIsKept(t *testing.T) {
	req, _ := buildRequest(t, newBuilder(), testcontainers.WithHostConfigModifier(func(hc *container.HostConfig) {
		hc.Binds = append(hc.Binds, "/tmp/data:/data")
	}))
	hc := &container.HostConfig{}
	req.HostConfigModifier(hc)
	var callerBind, socketBind bool
	for _, b := range hc.Binds {
		callerBind = callerBind || b == "/tmp/data:/data"
		socketBind = socketBind || b == core.DockerSocket+":"+core.DockerSocket
	}
	if !callerBind || !socketBind {
		t.Fatalf("expected both the caller's bind and the socket bind, got %v", hc.Binds)
	}
}
