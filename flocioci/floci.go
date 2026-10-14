// Package flocioci provides the Testcontainers module for the Floci local OCI emulator (floci/floci-oci).
//
// Example:
//
//	fc, err := flocioci.Run(ctx, "floci/floci-oci:latest")
//	testcontainers.CleanupContainer(t, fc)
//	if err != nil { ... }
//
//	provider := common.NewRawConfigurationProvider(fc.GetTenancyID(), fc.GetUserID(), fc.GetRegion(),
//	    fc.GetFingerprint(), fc.GetPrivateKeyPEM(), nil)
//	client, _ := objectstorage.NewObjectStorageClientWithConfigurationProvider(provider)
//	client.Host = fc.GetEndpoint()
package flocioci

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"maps"
	"strconv"
	"time"

	"github.com/testcontainers/testcontainers-go"

	"github.com/floci-io/testcontainers-floci-go/internal/core"
)

const (
	defaultImage = "floci/floci-oci:latest"
	flociPort    = 4599

	// DefaultRegion is the region the emulator serves.
	DefaultRegion = "us-ashburn-1"
	// DefaultTenancyID is the tenancy of requests that carry no signature, and of the credentials
	// this module hands out.
	DefaultTenancyID = "ocid1.tenancy.oc1..flocilocaltenancy0000000000000000000000000000000000000000"
	// DefaultNamespace is the Object Storage namespace of the emulated tenancy.
	DefaultNamespace = "floci-local"
	// DefaultUserID is the user OCID of the credentials this module hands out. The emulator has no
	// user list to check it against; it is the one the emulator's documentation uses.
	DefaultUserID = "ocid1.user.oc1..flocilocaluser0000000000000000000000000000000000000000000000"
	// DefaultFingerprint is the API key fingerprint of the credentials this module hands out. It is
	// a placeholder, not the fingerprint of the generated key: the emulator reads the fingerprint
	// only as part of the key ID.
	DefaultFingerprint = "aa:bb:cc:dd:ee:ff:00:11:22:33:44:55:66:77:88:99"

	regionEnv    = "FLOCI_OCI_DEFAULT_REGION"
	tenancyIDEnv = "FLOCI_OCI_DEFAULT_TENANCY_ID"
	namespaceEnv = "FLOCI_OCI_DEFAULT_NAMESPACE"

	privateKeyBits = 2048
)

// ociDescriptor is the OCI emulator (floci/floci-oci) as data, for the shared core.
var ociDescriptor = core.Descriptor{
	Name:           "oci",
	DefaultImage:   defaultImage,
	Port:           flociPort,
	EnvPrefix:      "FLOCI_OCI_",
	HealthPath:     "/_floci-oci/health",
	ResetPath:      "/_floci-oci/state/reset",
	LogLevelEnv:    "QUARKUS_LOG_CATEGORY__IO_FLOCI_OCI__LEVEL",
	StartupTimeout: 120 * time.Second,
	// Services that spawn sibling containers, from the emulator's source: the ones with a mock switch.
	SocketServices: []core.SocketService{
		{Token: "FUNCTIONS", Mockable: true},
		{Token: "OKE", Mockable: true},
	},
}

// defaultEnv is set on every container; a caller's testcontainers.WithEnv overrides it.
var defaultEnv = map[string]string{
	// The extra HTTPS port defaults to 443, which an unprivileged container cannot bind; the
	// emulator serves HTTPS on its main port anyway.
	"FLOCI_OCI_TLS_HTTPS_PORT": "0",
}

// builder collects the settings of one Floci OCI container before it starts.
type builder struct {
	image            string
	envVars          map[string]string
	dedicatedNetwork bool
	// dockerSocketOverride overrides socket auto-detection when non-nil (see WithDockerSocket).
	dockerSocketOverride *bool
	// generatedNamespace is the namespace newBuilder set; each start replaces it with a fresh
	// one unless the caller chose their own.
	generatedNamespace string
	// privateKeyPEM is the caller's signing key (see WithPrivateKeyPEM); empty means generate one.
	privateKeyPEM string
}

func newBuilder() *builder {
	c := &builder{
		image:   defaultImage,
		envVars: maps.Clone(defaultEnv),
	}
	// Sibling containers are named after the resource; a unique namespace per container
	// keeps parallel test runs from colliding and makes leftovers attributable.
	c.generatedNamespace = core.NewNamespace()
	c.withEnv(ociDescriptor.ResourceNamespaceEnv(), c.generatedNamespace)
	c.applyAllConfigs()
	return c
}

// Option configures Floci-specific settings. It implements
// testcontainers.ContainerCustomizer, so Floci options and generic testcontainers
// options can be passed to Run together:
//
//	flocioci.Run(ctx, "floci/floci-oci:latest",
//	    flocioci.WithOkeConfig(flocioci.OkeConfig{Enabled: true, Mock: true}),
//	    testcontainers.WithEnv(map[string]string{"FLOCI_OCI_DEFAULT_REGION": "eu-frankfurt-1"}),
//	)
type Option func(*builder)

// Customize is a no-op: Floci options are applied to the module's own settings
// before the container request is built.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error { return nil }

// Run creates and starts a Floci OCI container from the given image ("" for
// floci/floci-oci:latest), following the testcontainers-go module convention. Floci options
// (WithObjectStorageConfig, WithDedicatedNetwork, ...) configure the emulator; any other
// testcontainers.ContainerCustomizer is applied to the container request after the module's
// defaults, so it can override them.
//
// The returned container is non-nil whenever the container was created, even if
// Run also returns an error, so it can always be cleaned up:
//
//	fc, err := flocioci.Run(ctx, "floci/floci-oci:latest")
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
	if env[ociDescriptor.ResourceNamespaceEnv()] == c.generatedNamespace {
		env[ociDescriptor.ResourceNamespaceEnv()] = core.NewNamespace()
	}
	return core.Request{
		Descriptor:       ociDescriptor,
		Image:            c.image,
		Env:              env,
		Ports:            []int{flociPort},
		DockerSocket:     c.dockerSocketOverride,
		DedicatedNetwork: c.dedicatedNetwork,
	}
}

func (c *builder) run(ctx context.Context, customizers ...testcontainers.ContainerCustomizer) (*Container, error) {
	privateKey := c.privateKeyPEM
	if privateKey == "" {
		generated, err := newPrivateKeyPEM()
		if err != nil {
			return nil, fmt.Errorf("run floci oci: %w", err)
		}
		privateKey = generated
	}
	cc, env, err := core.Run(ctx, c.request(), customizers...)
	if cc == nil {
		return nil, err
	}
	return &Container{
		Container:     cc.Container,
		core:          cc,
		region:        envOr(env, regionEnv, DefaultRegion),
		tenancyID:     envOr(env, tenancyIDEnv, DefaultTenancyID),
		namespace:     envOr(env, namespaceEnv, DefaultNamespace),
		privateKeyPEM: privateKey,
	}, err
}

// newPrivateKeyPEM generates the throwaway RSA key of one container, in the PEM form
// "openssl genrsa -traditional" writes (PKCS #1), which every OCI SDK and the OCI CLI read.
func newPrivateKeyPEM() (string, error) {
	key, err := rsa.GenerateKey(rand.Reader, privateKeyBits)
	if err != nil {
		return "", fmt.Errorf("generating the API signing key: %w", err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})), nil
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

// Container is a running Floci OCI container. It embeds testcontainers.Container, so Logs, Exec,
// Inspect and testcontainers.CleanupContainer work on it; the shared core behind GetEndpoint,
// Reset and Terminate (which also removes the dedicated network and Floci's sibling containers)
// is kept in an unexported field.
type Container struct {
	testcontainers.Container
	core          *core.Container
	region        string
	tenancyID     string
	namespace     string
	privateKeyPEM string
}

// Container must stay a drop-in testcontainers.Container (CleanupContainer, Logs, Exec).
var _ testcontainers.Container = (*Container)(nil)

// GetEndpoint returns the base URL of every OCI service, e.g. "http://localhost:32768". Set it as
// the client's endpoint (client.Host in the Go SDK): the SDKs otherwise derive a real OCI
// address from the region.
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

// GetResourceNamespace returns the prefix of the sibling containers' names. It is unrelated to
// GetNamespace, the Object Storage namespace.
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

// Reset wipes all emulator state (buckets, queues, keys, ...) without restarting.
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

// GetRegion returns the region the emulator serves ("us-ashburn-1" unless WithRegion or
// FLOCI_OCI_DEFAULT_REGION set another).
func (s *Container) GetRegion() string { return s.region }

// GetTenancyID returns the tenancy OCID: the emulator's default tenancy, which is also the root
// compartment and the tenancy of unsigned requests. The emulator keeps each tenancy's resources
// apart, so requests signed with these credentials and unsigned ones see the same resources.
func (s *Container) GetTenancyID() string { return s.tenancyID }

// GetNamespace returns the Object Storage namespace ("floci-local" unless WithNamespace or
// FLOCI_OCI_DEFAULT_NAMESPACE set another).
func (s *Container) GetNamespace() string { return s.namespace }

// GetUserID returns the user OCID to sign requests as.
func (s *Container) GetUserID() string { return DefaultUserID }

// GetFingerprint returns the API key fingerprint to sign requests with.
func (s *Container) GetFingerprint() string { return DefaultFingerprint }

// GetPrivateKeyPEM returns the API signing key: an RSA private key in PEM form, generated for this
// container and worthless anywhere else (or the key WithPrivateKeyPEM supplied). The emulator reads
// the tenancy and user from a request's signature but never verifies it, so any key does; this one
// saves generating your own.
func (s *Container) GetPrivateKeyPEM() string { return s.privateKeyPEM }

// GetConfigFileContents returns an OCI configuration file with one profile for this container,
// the format ~/.oci/config has and the OCI SDKs and CLI read. The profile is named DEFAULT unless
// a name is given. An OCI configuration file names its key by path, so write GetPrivateKeyPEM to
// a file and pass that path as keyFile:
//
//	keyFile := filepath.Join(t.TempDir(), "floci_key.pem")
//	os.WriteFile(keyFile, []byte(fc.GetPrivateKeyPEM()), 0o600)
//	configFile := filepath.Join(t.TempDir(), "config")
//	os.WriteFile(configFile, []byte(fc.GetConfigFileContents(keyFile)), 0o600)
//
// The file carries no endpoint: clients still need GetEndpoint.
func (s *Container) GetConfigFileContents(keyFile string, profile ...string) string {
	name := "DEFAULT"
	if len(profile) > 0 && profile[0] != "" {
		name = profile[0]
	}
	return "[" + name + "]\n" +
		"user=" + s.GetUserID() + "\n" +
		"fingerprint=" + s.GetFingerprint() + "\n" +
		"tenancy=" + s.GetTenancyID() + "\n" +
		"region=" + s.GetRegion() + "\n" +
		"key_file=" + keyFile + "\n"
}
