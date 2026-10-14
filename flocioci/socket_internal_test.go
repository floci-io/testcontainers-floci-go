package flocioci

import (
	"crypto/x509"
	"encoding/pem"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/moby/moby/api/types/container"
	"github.com/testcontainers/testcontainers-go"

	"github.com/floci-io/testcontainers-floci-go/internal/core"
)

// Option must satisfy testcontainers.ContainerCustomizer so it can be passed to Run
// alongside generic testcontainers options.
var _ testcontainers.ContainerCustomizer = Option(nil)

func build(opts ...Option) *builder {
	c := newBuilder()
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// disableSocketServices turns off every service that spawns sibling containers, leaving
// a container that does not need the Docker socket.
func disableSocketServices() []Option {
	return []Option{
		WithFunctionsConfig(FunctionsConfig{}),
		WithOkeConfig(OkeConfig{}),
	}
}

func needsSocket(opts ...Option) bool {
	c := build(append(disableSocketServices(), opts...)...)
	return c.needsDockerSocket(c.envVars)
}

func TestDescriptor(t *testing.T) {
	d := ociDescriptor
	for name, got := range map[string][2]string{
		"image":         {d.DefaultImage, "floci/floci-oci:latest"},
		"port":          {d.PortSpec(), "4599/tcp"},
		"env prefix":    {d.EnvPrefix, "FLOCI_OCI_"},
		"health":        {d.HealthPath, "/_floci-oci/health"},
		"reset":         {d.ResetPath, "/_floci-oci/state/reset"},
		"log level":     {d.LogLevelEnv, "QUARKUS_LOG_CATEGORY__IO_FLOCI_OCI__LEVEL"},
		"network":       {d.NetworkEnv(), "FLOCI_OCI_SERVICES_DOCKER_NETWORK"},
		"namespace env": {d.ResourceNamespaceEnv(), "FLOCI_OCI_DOCKER_RESOURCE_NAMESPACE"},
		"service env":   {d.ServiceEnv("OBJECTSTORAGE", "ENABLED"), "FLOCI_OCI_SERVICES_OBJECTSTORAGE_ENABLED"},
		"mock setting":  {d.MockSetting, ""}, // the core's default, MOCK
		"region":        {DefaultRegion, "us-ashburn-1"},
		"tenancy":       {DefaultTenancyID, "ocid1.tenancy.oc1..flocilocaltenancy0000000000000000000000000000000000000000"},
		"namespace":     {DefaultNamespace, "floci-local"},
		"user":          {DefaultUserID, "ocid1.user.oc1..flocilocaluser0000000000000000000000000000000000000000000000"},
		"fingerprint":   {DefaultFingerprint, "aa:bb:cc:dd:ee:ff:00:11:22:33:44:55:66:77:88:99"},
		"default image": {newBuilder().image, "floci/floci-oci:latest"},
	} {
		if got[0] != got[1] {
			t.Errorf("%s = %q, want %q", name, got[0], got[1])
		}
	}
}

// The default environment: the fixed setting, a generated namespace, and for every service its
// enabled switch plus, where it has one, its mock switch at the emulator's default.
func TestDefaultEnv(t *testing.T) {
	env := newBuilder().envVars
	want := map[string]string{"FLOCI_OCI_TLS_HTTPS_PORT": "0"}
	for _, svc := range services {
		want["FLOCI_OCI_SERVICES_"+svc.token+"_ENABLED"] = "true"
		if svc.mockable {
			want["FLOCI_OCI_SERVICES_"+svc.token+"_MOCK"] = "false"
		}
	}
	got := maps.Clone(env)
	ns := got["FLOCI_OCI_DOCKER_RESOURCE_NAMESPACE"]
	if !strings.HasPrefix(ns, "tc-") || len(ns) != len("tc-")+8 {
		t.Errorf("namespace = %q, want tc- plus 8 hex characters", ns)
	}
	delete(got, "FLOCI_OCI_DOCKER_RESOURCE_NAMESPACE")
	if !maps.Equal(got, want) {
		for k, v := range want {
			if got[k] != v {
				t.Errorf("%s = %q, want %q", k, got[k], v)
			}
		}
		for k, v := range got {
			if _, ok := want[k]; !ok {
				t.Errorf("unexpected %s=%q", k, v)
			}
		}
	}
}

// The service list is the one the emulator's ServicesConfig declares, with its mock switches.
func TestServices_MatchTheEmulator(t *testing.T) {
	want := map[string]bool{ // token -> has a mock switch
		"IDENTITY": false, "OBJECTSTORAGE": false, "QUEUE": false, "KMS": false, "VAULT": false,
		"STREAMING": false, "FUNCTIONS": true, "OKE": true,
	}
	got := map[string]bool{}
	for _, svc := range services {
		got[svc.token] = svc.mockable
	}
	if !maps.Equal(got, want) || len(services) != len(want) {
		t.Fatalf("services = %v, want %v", got, want)
	}
}

// Every socket service is a known service with a mock switch, and the other way round.
func TestSocketTable_MatchesServices(t *testing.T) {
	byToken := map[string]service{}
	for _, svc := range services {
		byToken[svc.token] = svc
	}
	var tokens []string
	for _, ss := range ociDescriptor.SocketServices {
		tokens = append(tokens, ss.Token)
		svc, ok := byToken[ss.Token]
		if !ok {
			t.Errorf("socket service %s is not in the service list", ss.Token)
			continue
		}
		if ss.Mockable != svc.mockable || ss.MockedByDefault {
			t.Errorf("%s: socket table says mockable=%v mockedByDefault=%v, service list says mockable=%v",
				ss.Token, ss.Mockable, ss.MockedByDefault, svc.mockable)
		}
	}
	want := []string{"FUNCTIONS", "OKE"}
	if !slices.Equal(tokens, want) {
		t.Errorf("socket services = %v, want %v", tokens, want)
	}
	for _, svc := range services {
		if svc.mockable && !slices.Contains(tokens, svc.token) {
			t.Errorf("%s has a mock switch but is not a socket service", svc.token)
		}
	}
}

func TestDockerSocket_DefaultsMountIt(t *testing.T) {
	c := newBuilder()
	if !c.needsDockerSocket(c.envVars) {
		t.Fatal("expected the default configuration to need the Docker socket")
	}
}

func TestDockerSocket_NotNeededWithoutContainerServices(t *testing.T) {
	if needsSocket() {
		t.Fatal("expected no Docker socket once every container-backed service is disabled")
	}
}

func TestDockerSocket_PerService(t *testing.T) {
	tests := []struct {
		name string
		opt  Option
		want bool
	}{
		{"functions default", WithFunctionsConfig(DefaultFunctionsConfig()), true},
		{"functions mocked", WithFunctionsConfig(FunctionsConfig{Enabled: true, Mock: true}), false},
		{"functions disabled", WithFunctionsConfig(FunctionsConfig{}), false},
		{"oke default", WithOkeConfig(DefaultOkeConfig()), true},
		{"oke mocked", WithOkeConfig(OkeConfig{Enabled: true, Mock: true}), false},
		{"oke disabled", WithOkeConfig(OkeConfig{}), false},
		{"object storage", WithObjectStorageConfig(DefaultObjectStorageConfig()), false},
		{"streaming", WithStreamingConfig(DefaultStreamingConfig()), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := needsSocket(tt.opt); got != tt.want {
				t.Fatalf("needs the Docker socket = %v, want %v", got, tt.want)
			}
		})
	}
}

// Detection reads the final environment, so settings that arrive without a typed config count:
// a missing mock switch means the emulator's default, not mocked.
func TestDockerSocket_FromRawEnv(t *testing.T) {
	off := map[string]string{"FLOCI_OCI_SERVICES_FUNCTIONS_ENABLED": "false", "FLOCI_OCI_SERVICES_OKE_ENABLED": "false"}
	tests := []struct {
		name string
		env  map[string]string
		want bool
	}{
		{"nothing set: the defaults need it", nil, true},
		{"only services without containers", map[string]string{"FLOCI_OCI_SERVICES_QUEUE_ENABLED": "true"}, false},
		{"oke enabled, mock switch unset", map[string]string{"FLOCI_OCI_SERVICES_OKE_ENABLED": "true"}, true},
		{"oke enabled, mocked", map[string]string{"FLOCI_OCI_SERVICES_OKE_ENABLED": "true", "FLOCI_OCI_SERVICES_OKE_MOCK": "TRUE"}, false},
		{"functions enabled, not mocked", map[string]string{"FLOCI_OCI_SERVICES_FUNCTIONS_ENABLED": "true", "FLOCI_OCI_SERVICES_FUNCTIONS_MOCK": "false"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := map[string]string{}
			if tt.env != nil {
				env = maps.Clone(off)
				maps.Copy(env, tt.env)
			}
			if got := ociDescriptor.DockerSocketRequired(env); got != tt.want {
				t.Fatalf("needs the Docker socket = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDockerSocket_OverrideWins(t *testing.T) {
	off := build(WithDockerSocket(false))
	if off.needsDockerSocket(off.envVars) {
		t.Error("WithDockerSocket(false) should disable the mount even with container-backed services enabled")
	}
	on := build(append(disableSocketServices(), WithDockerSocket(true))...)
	if !on.needsDockerSocket(on.envVars) {
		t.Error("WithDockerSocket(true) should force the mount even with no container-backed service")
	}
}

// _ENABLED is always emitted; every other setting only while the service is enabled, and
// re-applying a config replaces what the earlier one set.
func TestServiceConfig_EnvEmission(t *testing.T) {
	serviceEnv := func(c *builder, token string) map[string]string {
		out := map[string]string{}
		for k, v := range c.envVars {
			if name, ok := strings.CutPrefix(k, "FLOCI_OCI_SERVICES_"+token+"_"); ok {
				out[name] = v
			}
		}
		return out
	}
	tests := []struct {
		name  string
		token string
		opts  []Option
		want  map[string]string
	}{
		{"enabled only", "OBJECTSTORAGE", []Option{WithObjectStorageConfig(ObjectStorageConfig{Enabled: true})}, map[string]string{"ENABLED": "true"}},
		{"disabled", "QUEUE", []Option{WithQueueConfig(QueueConfig{})}, map[string]string{"ENABLED": "false"}},
		{"mocked", "FUNCTIONS", []Option{WithFunctionsConfig(FunctionsConfig{Enabled: true, Mock: true})},
			map[string]string{"ENABLED": "true", "MOCK": "true"}},
		{"not mocked", "OKE", []Option{WithOkeConfig(OkeConfig{Enabled: true})}, map[string]string{"ENABLED": "true", "MOCK": "false"}},
		{"disabled drops the mock switch", "OKE", []Option{WithOkeConfig(OkeConfig{Mock: true})}, map[string]string{"ENABLED": "false"}},
		{"applied twice", "FUNCTIONS", []Option{WithFunctionsConfig(FunctionsConfig{Enabled: true, Mock: true}), WithFunctionsConfig(FunctionsConfig{Enabled: true, Mock: true})},
			map[string]string{"ENABLED": "true", "MOCK": "true"}},
		{"enabled then disabled", "FUNCTIONS", []Option{WithFunctionsConfig(FunctionsConfig{Enabled: true, Mock: true}), WithFunctionsConfig(FunctionsConfig{})},
			map[string]string{"ENABLED": "false"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := serviceEnv(build(tt.opts...), tt.token); !maps.Equal(got, tt.want) {
				t.Fatalf("env = %v, want %v", got, tt.want)
			}
		})
	}
}

// Every service has an option, and each option writes its own service's switch.
func TestOptions_OnePerService(t *testing.T) {
	options := map[string]Option{
		"IDENTITY": WithIdentityConfig(IdentityConfig{}), "OBJECTSTORAGE": WithObjectStorageConfig(ObjectStorageConfig{}),
		"QUEUE": WithQueueConfig(QueueConfig{}), "KMS": WithKmsConfig(KmsConfig{}), "VAULT": WithVaultConfig(VaultConfig{}),
		"STREAMING": WithStreamingConfig(StreamingConfig{}), "FUNCTIONS": WithFunctionsConfig(FunctionsConfig{}),
		"OKE": WithOkeConfig(OkeConfig{}),
	}
	if len(options) != len(services) {
		t.Fatalf("%d options for %d services", len(options), len(services))
	}
	for _, svc := range services {
		opt, ok := options[svc.token]
		if !ok {
			t.Errorf("no option for %s", svc.token)
			continue
		}
		c := build(opt)
		for _, other := range services {
			want := boolStr(other.token != svc.token)
			if got := c.envVars["FLOCI_OCI_SERVICES_"+other.token+"_ENABLED"]; got != want {
				t.Errorf("after disabling %s: %s enabled = %q, want %q", svc.token, other.token, got, want)
			}
		}
	}
}

func TestOptions_ApplyToBuilder(t *testing.T) {
	c := build(WithDedicatedNetwork(), WithDockerSocket(false), WithVaultConfig(VaultConfig{}))
	if got := c.envVars["FLOCI_OCI_SERVICES_VAULT_ENABLED"]; got != "false" {
		t.Errorf("FLOCI_OCI_SERVICES_VAULT_ENABLED = %q, want false", got)
	}
	if !c.dedicatedNetwork {
		t.Error("WithDedicatedNetwork option did not set the dedicated network")
	}
	if c.needsDockerSocket(c.envVars) {
		t.Error("WithDockerSocket(false) option did not override auto-detection")
	}
	if !c.request().DedicatedNetwork || !slices.Equal(c.request().Ports, []int{4599}) {
		t.Errorf("request = %+v", c.request())
	}
}

// The identity options set the emulator's defaults; a generic WithEnv still overrides them.
func TestOptions_Identity(t *testing.T) {
	keys := []string{"FLOCI_OCI_DEFAULT_REGION", "FLOCI_OCI_DEFAULT_TENANCY_ID", "FLOCI_OCI_DEFAULT_NAMESPACE"}
	for _, key := range keys {
		if _, set := newBuilder().envVars[key]; set {
			t.Errorf("%s should be left to the emulator by default", key)
		}
	}
	c := build(WithRegion("eu-frankfurt-1"), WithTenancyID("ocid1.tenancy.oc1..mine"), WithNamespace("my-namespace"), WithPrivateKeyPEM("pem"))
	_, env := buildRequest(t, c)
	if env[keys[0]] != "eu-frankfurt-1" || env[keys[1]] != "ocid1.tenancy.oc1..mine" || env[keys[2]] != "my-namespace" {
		t.Errorf("env = %v", env)
	}
	if c.privateKeyPEM != "pem" {
		t.Errorf("private key = %q", c.privateKeyPEM)
	}
	if _, env = buildRequest(t, c, testcontainers.WithEnv(map[string]string{keys[0]: "uk-london-1"})); env[keys[0]] != "uk-london-1" {
		t.Errorf("caller's env did not win: %v", env)
	}
}

// buildRequest applies Run's customizers to an empty request without starting Docker,
// returning the request and the final environment Run would record.
func buildRequest(t *testing.T, c *builder, customizers ...testcontainers.ContainerCustomizer) (*testcontainers.GenericContainerRequest, map[string]string) {
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
	return slices.Contains(hc.Binds, core.DockerSocket+":"+core.DockerSocket)
}

// The fixed default reaches the container, and a generic WithEnv overrides it.
func TestRequest_DefaultEnvIsOverridable(t *testing.T) {
	_, env := buildRequest(t, newBuilder())
	if env["FLOCI_OCI_TLS_HTTPS_PORT"] != "0" {
		t.Fatalf("default env missing: %v", env)
	}
	_, env = buildRequest(t, newBuilder(), testcontainers.WithEnv(map[string]string{"FLOCI_OCI_TLS_HTTPS_PORT": "8443"}))
	if env["FLOCI_OCI_TLS_HTTPS_PORT"] != "8443" {
		t.Fatalf("caller's env did not win: %v", env)
	}
}

// Mocking every container-backed service through a generic env override must drop the socket,
// and un-mocking one must bring it back.
func TestRequest_GenericEnvMockingServicesDropsSocket(t *testing.T) {
	env := map[string]string{"FLOCI_OCI_SERVICES_FUNCTIONS_MOCK": "true", "FLOCI_OCI_SERVICES_OKE_MOCK": "true"}
	if req, _ := buildRequest(t, newBuilder(), testcontainers.WithEnv(env)); mountsDockerSocket(req) {
		t.Fatal("expected no socket once every container-backed service is mocked through testcontainers.WithEnv")
	}
	env["FLOCI_OCI_SERVICES_OKE_MOCK"] = "false"
	if req, _ := buildRequest(t, newBuilder(), testcontainers.WithEnv(env)); !mountsDockerSocket(req) {
		t.Fatal("expected the socket with OKE not mocked")
	}
}

// Disabling every container-backed service through a generic env override must drop the socket.
func TestRequest_GenericEnvDisablingServicesDropsSocket(t *testing.T) {
	env := map[string]string{}
	for _, svc := range ociDescriptor.SocketServices {
		env["FLOCI_OCI_SERVICES_"+svc.Token+"_ENABLED"] = "false"
	}
	req, _ := buildRequest(t, newBuilder(), testcontainers.WithEnv(env))
	if mountsDockerSocket(req) {
		t.Fatal("expected no socket once every container-backed service is disabled through testcontainers.WithEnv")
	}
}

// Each start gets its own resource namespace; a namespace the caller chose is kept.
func TestRequest_FreshNamespacePerStart(t *testing.T) {
	key := ociDescriptor.ResourceNamespaceEnv()
	b := newBuilder()
	first, second := b.request().Env[key], b.request().Env[key]
	if first == second || first == b.generatedNamespace {
		t.Fatalf("namespaces %q and %q should be fresh and distinct", first, second)
	}
	b.withEnv(key, "ci-42")
	if got := b.request().Env[key]; got != "ci-42" {
		t.Fatalf("caller's namespace = %q, want ci-42", got)
	}
}

// Every call generates a new 2048-bit RSA key, in the PKCS #1 PEM form.
func TestPrivateKey(t *testing.T) {
	first, err := newPrivateKeyPEM()
	if err != nil {
		t.Fatal(err)
	}
	second, err := newPrivateKeyPEM()
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("two calls returned the same key")
	}
	block, rest := pem.Decode([]byte(first))
	if block == nil || block.Type != "RSA PRIVATE KEY" || len(rest) != 0 {
		t.Fatalf("not a single RSA PRIVATE KEY block: %q", first)
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if key.N.BitLen() != 2048 {
		t.Fatalf("key size = %d bits", key.N.BitLen())
	}
}

// The configuration file is one DEFAULT profile in the format of ~/.oci/config.
func TestContainer_ConnectionHelpers(t *testing.T) {
	c := &Container{region: "eu-frankfurt-1", tenancyID: "ocid1.tenancy.oc1..mine", namespace: "my-namespace", privateKeyPEM: "pem"}
	if c.GetEndpoint() != "" {
		t.Errorf("a container not started by Run has no endpoint")
	}
	if c.GetRegion() != "eu-frankfurt-1" || c.GetTenancyID() != "ocid1.tenancy.oc1..mine" || c.GetNamespace() != "my-namespace" ||
		c.GetUserID() != DefaultUserID || c.GetFingerprint() != DefaultFingerprint || c.GetPrivateKeyPEM() != "pem" {
		t.Errorf("identity = %q %q %q %q %q %q", c.GetRegion(), c.GetTenancyID(), c.GetNamespace(), c.GetUserID(), c.GetFingerprint(), c.GetPrivateKeyPEM())
	}
	want := "[DEFAULT]\n" +
		"user=ocid1.user.oc1..flocilocaluser0000000000000000000000000000000000000000000000\n" +
		"fingerprint=aa:bb:cc:dd:ee:ff:00:11:22:33:44:55:66:77:88:99\n" +
		"tenancy=ocid1.tenancy.oc1..mine\n" +
		"region=eu-frankfurt-1\n" +
		"key_file=/tmp/floci_key.pem\n"
	if got := c.GetConfigFileContents("/tmp/floci_key.pem"); got != want {
		t.Errorf("configuration file = %q, want %q", got, want)
	}
	if got := c.GetConfigFileContents("/tmp/floci_key.pem", ""); got != want {
		t.Errorf("configuration file with an empty profile name = %q, want %q", got, want)
	}
	named := strings.Replace(want, "[DEFAULT]", "[FLOCI]", 1)
	if got := c.GetConfigFileContents("/tmp/floci_key.pem", "FLOCI"); got != named {
		t.Errorf("configuration file = %q, want %q", got, named)
	}
}
