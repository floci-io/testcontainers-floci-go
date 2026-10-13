// Package flociaz provides the Testcontainers module for the Floci local Azure emulator (floci/floci-az).
//
// Example:
//
//	fc, err := flociaz.Run(ctx, "floci/floci-az:latest")
//	testcontainers.CleanupContainer(t, fc)
//	if err != nil { ... }
//
//	client, _ := azblob.NewClientFromConnectionString(fc.GetStorageConnectionString(), nil)
package flociaz

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"

	"github.com/floci-io/testcontainers-floci-go/internal/core"
)

const (
	defaultImage = "floci/floci-az:latest"
	flociPort    = 4577
	tlsCertPath  = "/_floci/tls-cert"

	// DefaultSubscriptionID is the subscription the emulator serves.
	DefaultSubscriptionID = "00000000-0000-0000-0000-000000000001"
	// DefaultTenantID is the Microsoft Entra tenant the emulator serves.
	DefaultTenantID = "00000000-0000-0000-0000-000000000002"
	// DefaultAccountName is the storage account, the same one Azurite uses.
	DefaultAccountName = "devstoreaccount1"
	// DefaultAccountKey is the storage account key, the same one Azurite uses.
	DefaultAccountKey = "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw=="

	tenantIDEnv   = "FLOCI_AZ_SERVICES_ENTRA_DEFAULT_TENANT_ID"
	accountKeyEnv = "FLOCI_AZ_AUTH_STORAGE_ACCOUNT_KEYS_DEVSTOREACCOUNT1"
)

// azDescriptor is the Azure emulator (floci/floci-az) as data, for the shared core.
var azDescriptor = core.Descriptor{
	Name:           "az",
	DefaultImage:   defaultImage,
	Port:           flociPort,
	EnvPrefix:      "FLOCI_AZ_",
	HealthPath:     "/_floci/health",
	ResetPath:      "/_admin/reset",
	LogLevelEnv:    "QUARKUS_LOG_CATEGORY__IO_FLOCI_AZ__LEVEL",
	MockSetting:    "MOCKED",
	StartupTimeout: 120 * time.Second,
	// Services that spawn sibling containers, mirroring requiresDockerSocket() in the Java module.
	SocketServices: []core.SocketService{
		{Token: "ACI", Mockable: true, MockedByDefault: true},
		{Token: "ACR", Mockable: true},
		{Token: "AKS", Mockable: true},
		{Token: "CONTAINER_APPS", Mockable: true, MockedByDefault: true},
		{Token: "COSMOS", Mockable: true, Needs: cosmosNeedsDocker},
		{Token: "EVENT_HUB", Mockable: true},
		{Token: "FUNCTIONS", Mockable: true},
		{Token: "MARIA_DB", Mockable: true},
		{Token: "MYSQL", Mockable: true},
		{Token: "POSTGRES", Mockable: true},
		{Token: "REDIS", Mockable: true},
		{Token: "SERVICE_BUS", Mockable: true, MockedByDefault: true},
		{Token: "SQL", Needs: sqlNeedsDocker},
		{Token: "VM", Mockable: true, MockedByDefault: true},
	},
}

// defaultEnv is set on every container; a caller's testcontainers.WithEnv overrides it.
var defaultEnv = map[string]string{
	// The host and port the emulator reports for its sibling containers must be reachable from
	// the test process, not only from inside the Docker network.
	"FLOCI_AZ_DOCKER_ENDPOINT_MODE": "published",
	// The extra HTTPS port defaults to 443, which an unprivileged container cannot bind; the
	// emulator serves HTTPS on its main port anyway.
	"FLOCI_AZ_TLS_HTTPS_PORT": "0",
}

// builder collects the settings of one Floci Azure container before it starts.
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
	c.withEnv(azDescriptor.ResourceNamespaceEnv(), c.generatedNamespace)
	c.applyAllConfigs()
	return c
}

// Option configures Floci-specific settings. It implements
// testcontainers.ContainerCustomizer, so Floci options and generic testcontainers
// options can be passed to Run together:
//
//	flociaz.Run(ctx, "floci/floci-az:latest",
//	    flociaz.WithRedisConfig(flociaz.RedisConfig{Enabled: false}),
//	    testcontainers.WithEnv(map[string]string{"FLOCI_AZ_AUTH_MODE": "dev"}),
//	)
type Option func(*builder)

// Customize is a no-op: Floci options are applied to the module's own settings
// before the container request is built.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error { return nil }

// Run creates and starts a Floci Azure container from the given image ("" for
// floci/floci-az:latest), following the testcontainers-go module convention. Floci options
// (WithBlobConfig, WithDedicatedNetwork, ...) configure the emulator; any other
// testcontainers.ContainerCustomizer is applied to the container request after the module's
// defaults, so it can override them.
//
// The returned container is non-nil whenever the container was created, even if
// Run also returns an error, so it can always be cleaned up:
//
//	fc, err := flociaz.Run(ctx, "floci/floci-az:latest")
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
	if env[azDescriptor.ResourceNamespaceEnv()] == c.generatedNamespace {
		env[azDescriptor.ResourceNamespaceEnv()] = core.NewNamespace()
	}
	return core.Request{
		Descriptor:       azDescriptor,
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
		Container:  cc.Container,
		core:       cc,
		tenantID:   envOr(env, tenantIDEnv, DefaultTenantID),
		accountKey: envOr(env, accountKeyEnv, DefaultAccountKey),
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

// Container is a running Floci Azure container. It embeds testcontainers.Container, so Logs, Exec,
// Inspect and testcontainers.CleanupContainer work on it; the shared core behind GetEndpoint,
// Reset and Terminate (which also removes the dedicated network and Floci's sibling containers)
// is kept in an unexported field.
type Container struct {
	testcontainers.Container
	core       *core.Container
	tenantID   string
	accountKey string
}

// Container must stay a drop-in testcontainers.Container (CleanupContainer, Logs, Exec).
var _ testcontainers.Container = (*Container)(nil)

// GetEndpoint returns the base URL, e.g. "http://localhost:32768".
func (s *Container) GetEndpoint() string {
	if s.core == nil {
		return ""
	}
	return s.core.GetEndpoint()
}

// GetHTTPSEndpoint returns the base URL over TLS, e.g. "https://localhost:32768". The emulator
// serves HTTP and HTTPS on the same port once TLS is enabled (FLOCI_AZ_TLS_ENABLED=true).
func (s *Container) GetHTTPSEndpoint() string {
	endpoint := s.GetEndpoint()
	if endpoint == "" {
		return ""
	}
	return "https://" + strings.TrimPrefix(endpoint, "http://")
}

// GetTLSCertificate returns the PEM certificate clients must trust to use GetHTTPSEndpoint.
// It fails unless TLS is enabled.
func (s *Container) GetTLSCertificate(ctx context.Context) (string, error) {
	if s.core == nil {
		return "", errors.New("floci: GetTLSCertificate needs a container started by Run")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.GetEndpoint()+tlsCertPath, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching the TLS certificate: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading the TLS certificate: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("TLS certificate not available (HTTP %d): %s", resp.StatusCode, body)
	}
	return string(body), nil
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

// Reset wipes all emulator state (containers, blobs, queues, tables, ...) without restarting.
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

// GetSubscriptionID returns the Azure subscription ID, which is fixed in the emulator.
func (s *Container) GetSubscriptionID() string { return DefaultSubscriptionID }

// GetTenantID returns the Microsoft Entra tenant ID.
func (s *Container) GetTenantID() string { return s.tenantID }

// GetAccountName returns the storage account name (always "devstoreaccount1").
func (s *Container) GetAccountName() string { return DefaultAccountName }

// GetAccountKey returns the storage account key.
func (s *Container) GetAccountKey() string { return s.accountKey }

// GetBlobEndpoint returns the Blob Storage endpoint, e.g. "http://localhost:32768/devstoreaccount1".
func (s *Container) GetBlobEndpoint() string {
	return s.GetEndpoint() + "/" + s.GetAccountName()
}

// GetQueueEndpoint returns the Queue Storage endpoint, e.g. "http://localhost:32768/devstoreaccount1-queue".
func (s *Container) GetQueueEndpoint() string {
	return s.GetEndpoint() + "/" + s.GetAccountName() + "-queue"
}

// GetTableEndpoint returns the Table Storage endpoint, e.g. "http://localhost:32768/devstoreaccount1-table".
func (s *Container) GetTableEndpoint() string {
	return s.GetEndpoint() + "/" + s.GetAccountName() + "-table"
}

// GetStorageConnectionString returns the connection string the Azure Storage SDKs accept, with
// the blob, queue and table endpoints of this container.
func (s *Container) GetStorageConnectionString() string {
	return "DefaultEndpointsProtocol=http" +
		";AccountName=" + s.GetAccountName() +
		";AccountKey=" + s.GetAccountKey() +
		";BlobEndpoint=" + s.GetBlobEndpoint() +
		";QueueEndpoint=" + s.GetQueueEndpoint() +
		";TableEndpoint=" + s.GetTableEndpoint() +
		";"
}
