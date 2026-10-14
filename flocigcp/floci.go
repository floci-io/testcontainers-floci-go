// Package flocigcp provides the Testcontainers module for the Floci local GCP emulator (floci/floci-gcp).
//
// Example:
//
//	fc, err := flocigcp.Run(ctx, "floci/floci-gcp:latest")
//	testcontainers.CleanupContainer(t, fc)
//	if err != nil { ... }
//
//	for name, value := range fc.GetEmulatorEnv() {
//	    t.Setenv(name, value) // PUBSUB_EMULATOR_HOST, STORAGE_EMULATOR_HOST, ...
//	}
package flocigcp

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strconv"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"

	"github.com/floci-io/testcontainers-floci-go/internal/core"
)

const (
	defaultImage = "floci/floci-gcp:latest"
	flociPort    = 4588

	// DefaultProjectID is the project the emulator uses when a request names none. The emulator
	// accepts any project ID.
	DefaultProjectID = "floci-local"

	projectIDEnv = "FLOCI_GCP_DEFAULT_PROJECT_ID"
)

// gcpDescriptor is the GCP emulator (floci/floci-gcp) as data, for the shared core.
var gcpDescriptor = core.Descriptor{
	Name:           "gcp",
	DefaultImage:   defaultImage,
	Port:           flociPort,
	EnvPrefix:      "FLOCI_GCP_",
	HealthPath:     "/_floci-gcp/health",
	ResetPath:      "/_floci-gcp/state/reset",
	LogLevelEnv:    "QUARKUS_LOG_CATEGORY__IO_FLOCI_GCP__LEVEL",
	StartupTimeout: 120 * time.Second,
	// Services that spawn sibling containers, from the emulator's source: the ones with a mock
	// switch. Cloud Functions is metadata only and spawns nothing.
	SocketServices: []core.SocketService{
		{Token: "BIGQUERY", Mockable: true, Needs: bigQueryNeedsDocker},
		{Token: "CLOUDRUN", Mockable: true},
		{Token: "CLOUDSQL", Mockable: true},
		{Token: "GKE", Mockable: true},
		{Token: "KAFKA", Mockable: true},
	},
}

// defaultEnv is set on every container; a caller's testcontainers.WithEnv overrides it.
var defaultEnv = map[string]string{
	// The extra HTTPS port defaults to 443, which an unprivileged container cannot bind; the
	// emulator serves HTTPS on its main port anyway.
	"FLOCI_GCP_TLS_HTTPS_PORT": "0",
}

// builder collects the settings of one Floci GCP container before it starts.
type builder struct {
	image            string
	envVars          map[string]string
	dedicatedNetwork bool
	// dockerSocketOverride overrides socket auto-detection when non-nil (see WithDockerSocket).
	dockerSocketOverride *bool
	// generatedNamespace is the namespace newBuilder set; each start replaces it with a fresh
	// one unless the caller chose their own.
	generatedNamespace string
}

func newBuilder() *builder {
	c := &builder{
		image:   defaultImage,
		envVars: maps.Clone(defaultEnv),
	}
	// Sibling containers are named after the resource; a unique namespace per container
	// keeps parallel test runs from colliding and makes leftovers attributable.
	c.generatedNamespace = core.NewNamespace()
	c.withEnv(gcpDescriptor.ResourceNamespaceEnv(), c.generatedNamespace)
	c.applyAllConfigs()
	return c
}

// Option configures Floci-specific settings. It implements
// testcontainers.ContainerCustomizer, so Floci options and generic testcontainers
// options can be passed to Run together:
//
//	flocigcp.Run(ctx, "floci/floci-gcp:latest",
//	    flocigcp.WithCloudSqlConfig(flocigcp.CloudSqlConfig{Enabled: true, Mock: true}),
//	    testcontainers.WithEnv(map[string]string{"FLOCI_GCP_DEFAULT_PROJECT_ID": "my-project"}),
//	)
type Option func(*builder)

// Customize is a no-op: Floci options are applied to the module's own settings
// before the container request is built.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error { return nil }

// Run creates and starts a Floci GCP container from the given image ("" for
// floci/floci-gcp:latest), following the testcontainers-go module convention. Floci options
// (WithGcsConfig, WithDedicatedNetwork, ...) configure the emulator; any other
// testcontainers.ContainerCustomizer is applied to the container request after the module's
// defaults, so it can override them.
//
// The returned container is non-nil whenever the container was created, even if
// Run also returns an error, so it can always be cleaned up:
//
//	fc, err := flocigcp.Run(ctx, "floci/floci-gcp:latest")
//	testcontainers.CleanupContainer(t, fc)
//	require.NoError(t, err)
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	c := newBuilder()
	if img != "" {
		c.image = img
	}
	var customizers []testcontainers.ContainerCustomizer
	for _, opt := range opts {
		if o, ok := opt.(Option); ok {
			o(c)
			continue
		}
		customizers = append(customizers, opt)
	}
	return c.run(ctx, customizers...)
}

func (c *builder) withEnv(key, value string) *builder {
	c.envVars[key] = value
	return c
}

// needsDockerSocket reports whether the host Docker socket should be mounted for the
// given final container environment: an explicit WithDockerSocket override wins,
// otherwise any enabled service that spawns sibling containers requires it.
func (c *builder) needsDockerSocket(env map[string]string) bool {
	return c.request().NeedsDockerSocket(env)
}

// request is the builder's state as the core's start request.
func (c *builder) request() core.Request {
	env := maps.Clone(c.envVars)
	if env[gcpDescriptor.ResourceNamespaceEnv()] == c.generatedNamespace {
		env[gcpDescriptor.ResourceNamespaceEnv()] = core.NewNamespace()
	}
	return core.Request{
		Descriptor:       gcpDescriptor,
		Image:            c.image,
		Env:              env,
		Ports:            []int{flociPort},
		DockerSocket:     c.dockerSocketOverride,
		DedicatedNetwork: c.dedicatedNetwork,
	}
}

func (c *builder) run(ctx context.Context, customizers ...testcontainers.ContainerCustomizer) (*Container, error) {
	cc, env, err := core.Run(ctx, c.request(), customizers...)
	if cc == nil {
		return nil, err
	}
	return &Container{
		Container: cc.Container,
		core:      cc,
		projectID: envOr(env, projectIDEnv, DefaultProjectID),
	}, err
}

// requestOptions returns the customizers Run passes to testcontainers.Run (see core.Request.Options).
func (c *builder) requestOptions(network *testcontainers.DockerNetwork, customizers []testcontainers.ContainerCustomizer) ([]testcontainers.ContainerCustomizer, *map[string]string) {
	return c.request().Options(network, customizers)
}

func envOr(env map[string]string, key, fallback string) string {
	if v := env[key]; v != "" {
		return v
	}
	return fallback
}

// Container is a running Floci GCP container. It embeds testcontainers.Container, so Logs, Exec,
// Inspect and testcontainers.CleanupContainer work on it; the shared core behind GetEndpoint,
// Reset and Terminate (which also removes the dedicated network and Floci's sibling containers)
// is kept in an unexported field.
type Container struct {
	testcontainers.Container
	core      *core.Container
	projectID string
}

// Container must stay a drop-in testcontainers.Container (CleanupContainer, Logs, Exec).
var _ testcontainers.Container = (*Container)(nil)

// GetEndpoint returns the base URL of every GCP service, e.g. "http://localhost:32768". REST and
// gRPC are served on that one port.
func (s *Container) GetEndpoint() string {
	if s.core == nil {
		return ""
	}
	return s.core.GetEndpoint()
}

// GetDedicatedNetworkName returns the dedicated Docker network name, or "" if none.
func (s *Container) GetDedicatedNetworkName() string {
	if s.core == nil {
		return ""
	}
	return s.core.GetDedicatedNetworkName()
}

// GetResourceNamespace returns the prefix of the sibling containers' names.
func (s *Container) GetResourceNamespace() string {
	if s.core == nil {
		return ""
	}
	return s.core.GetResourceNamespace()
}

// GetMappedPort returns the host-mapped port for a given container port, asking the embedded
// Container (so a wrapper assigned to it is honoured).
func (s *Container) GetMappedPort(ctx context.Context, port int) (int, error) {
	mapped, err := s.MappedPort(ctx, fmt.Sprintf("%d/tcp", port))
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(mapped.Port())
}

// Reset wipes all emulator state (buckets, topics, documents, ...) without restarting.
func (s *Container) Reset(ctx context.Context) error {
	if s.core == nil {
		return errors.New("floci: Reset needs a container started by Run")
	}
	return s.core.Reset(ctx)
}

// Terminate stops and removes the container through the embedded Container (so a wrapper
// assigned to it runs its own cleanup), then removes Floci's sibling containers and the dedicated
// network, if any.
func (s *Container) Terminate(ctx context.Context, opts ...testcontainers.TerminateOption) error {
	err := s.Container.Terminate(ctx, opts...)
	if s.core != nil {
		err = errors.Join(err, s.core.Cleanup(ctx))
	}
	return err
}

// GetProjectID returns the emulator's default project ("floci-local" unless WithProjectID or
// FLOCI_GCP_DEFAULT_PROJECT_ID set another). The emulator accepts any other project ID as well.
func (s *Container) GetProjectID() string { return s.projectID }

// GetEmulatorEnv returns the environment variables that point Google client libraries at this
// container, to set on the test process or on the application under test:
//
//   - PUBSUB_EMULATOR_HOST, FIRESTORE_EMULATOR_HOST, DATASTORE_EMULATOR_HOST,
//     SECRET_MANAGER_EMULATOR_HOST, FIREBASE_AUTH_EMULATOR_HOST and STORAGE_EMULATOR_HOST_GRPC
//     as a bare "host:port";
//   - STORAGE_EMULATOR_HOST as a URL ("http://host:port"), the form the Cloud Storage libraries read;
//   - GOOGLE_CLOUD_PROJECT as GetProjectID.
//
// LOGGING_EMULATOR_HOST and KMS_EMULATOR_HOST (bare "host:port") are in the map too, but no Google
// library reads them: they are the names floci-gcp's own compatibility tests use, for code that
// builds the client endpoint itself. Every other service needs an explicit endpoint, GetEndpoint.
func (s *Container) GetEmulatorEnv() map[string]string {
	endpoint := s.GetEndpoint()
	hostPort := strings.TrimPrefix(endpoint, "http://")
	return map[string]string{
		"PUBSUB_EMULATOR_HOST":         hostPort,
		"FIRESTORE_EMULATOR_HOST":      hostPort,
		"DATASTORE_EMULATOR_HOST":      hostPort,
		"SECRET_MANAGER_EMULATOR_HOST": hostPort,
		"FIREBASE_AUTH_EMULATOR_HOST":  hostPort,
		"LOGGING_EMULATOR_HOST":        hostPort,
		"KMS_EMULATOR_HOST":            hostPort,
		"STORAGE_EMULATOR_HOST":        endpoint,
		"STORAGE_EMULATOR_HOST_GRPC":   hostPort,
		"GOOGLE_CLOUD_PROJECT":         s.GetProjectID(),
	}
}
