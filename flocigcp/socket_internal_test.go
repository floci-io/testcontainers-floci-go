package flocigcp

import (
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
		WithBigQueryConfig(BigQueryConfig{}),
		WithCloudRunConfig(CloudRunConfig{}),
		WithCloudSqlConfig(CloudSqlConfig{}),
		WithGkeConfig(GkeConfig{}),
		WithKafkaConfig(KafkaConfig{}),
	}
}

func needsSocket(opts ...Option) bool {
	c := build(append(disableSocketServices(), opts...)...)
	return c.needsDockerSocket(c.envVars)
}

func TestDescriptor(t *testing.T) {
	d := gcpDescriptor
	for name, got := range map[string][2]string{
		"image":         {d.DefaultImage, "floci/floci-gcp:latest"},
		"port":          {d.PortSpec(), "4588/tcp"},
		"env prefix":    {d.EnvPrefix, "FLOCI_GCP_"},
		"health":        {d.HealthPath, "/_floci-gcp/health"},
		"reset":         {d.ResetPath, "/_floci-gcp/state/reset"},
		"log level":     {d.LogLevelEnv, "QUARKUS_LOG_CATEGORY__IO_FLOCI_GCP__LEVEL"},
		"network":       {d.NetworkEnv(), "FLOCI_GCP_SERVICES_DOCKER_NETWORK"},
		"namespace":     {d.ResourceNamespaceEnv(), "FLOCI_GCP_DOCKER_RESOURCE_NAMESPACE"},
		"service env":   {d.ServiceEnv("GCS", "ENABLED"), "FLOCI_GCP_SERVICES_GCS_ENABLED"},
		"mock setting":  {d.MockSetting, ""}, // the core's default, MOCK
		"project":       {DefaultProjectID, "floci-local"},
		"default image": {newBuilder().image, "floci/floci-gcp:latest"},
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
	want := map[string]string{"FLOCI_GCP_TLS_HTTPS_PORT": "0"}
	for _, svc := range services {
		want["FLOCI_GCP_SERVICES_"+svc.token+"_ENABLED"] = "true"
		if svc.mockable {
			want["FLOCI_GCP_SERVICES_"+svc.token+"_MOCK"] = "false"
		}
	}
	got := maps.Clone(env)
	ns := got["FLOCI_GCP_DOCKER_RESOURCE_NAMESPACE"]
	if !strings.HasPrefix(ns, "tc-") || len(ns) != len("tc-")+8 {
		t.Errorf("namespace = %q, want tc- plus 8 hex characters", ns)
	}
	delete(got, "FLOCI_GCP_DOCKER_RESOURCE_NAMESPACE")
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
		"COMPUTE": false, "GCS": false, "PUBSUB": false, "FIRESTORE": false, "DATASTORE": false, "IAM": false,
		"IAMCREDENTIALS": false, "STS": false, "SECRETMANAGER": false, "LOGGING": false, "KMS": false,
		"KAFKA": true, "CLOUDSQL": true, "CLOUDTASKS": false, "CLOUDRUN": true, "CLOUDFUNCTIONS": false,
		"MONITORING": false, "SCHEDULER": false, "EVENTARC": false, "GKE": true, "SERVICEUSAGE": false,
		"RESOURCEMANAGER": false, "FIREBASEAUTH": false, "BIGQUERY": true,
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
	for _, ss := range gcpDescriptor.SocketServices {
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
	want := []string{"BIGQUERY", "CLOUDRUN", "CLOUDSQL", "GKE", "KAFKA"}
	if !slices.Equal(tokens, want) {
		t.Errorf("socket services = %v, want %v", tokens, want)
	}
	for _, svc := range services {
		if svc.mockable && !slices.Contains(tokens, svc.token) {
			t.Errorf("%s has a mock switch but is not a socket service", svc.token)
		}
	}
}

// With the emulator's defaults, each socket service alone makes the container need the socket.
func TestDockerSocket_ServicesNeedingItByDefault(t *testing.T) {
	var got []string
	for _, ss := range gcpDescriptor.SocketServices {
		env := map[string]string{}
		for _, other := range gcpDescriptor.SocketServices {
			env[gcpDescriptor.ServiceEnv(other.Token, "ENABLED")] = boolStr(other.Token == ss.Token)
		}
		if gcpDescriptor.DockerSocketRequired(env) {
			got = append(got, ss.Token)
		}
	}
	want := []string{"BIGQUERY", "CLOUDRUN", "CLOUDSQL", "GKE", "KAFKA"}
	if !slices.Equal(got, want) {
		t.Errorf("services needing the socket by default = %v, want %v", got, want)
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
		{"bigquery default", WithBigQueryConfig(DefaultBigQueryConfig()), true},
		{"bigquery mocked", WithBigQueryConfig(BigQueryConfig{Enabled: true, Mock: true}), false},
		{"bigquery with its own floci-duck", WithBigQueryConfig(BigQueryConfig{Enabled: true, DuckURL: "http://duck:3000"}), false},
		{"bigquery disabled", WithBigQueryConfig(BigQueryConfig{}), false},
		{"cloud run default", WithCloudRunConfig(DefaultCloudRunConfig()), true},
		{"cloud run mocked", WithCloudRunConfig(CloudRunConfig{Enabled: true, Mock: true}), false},
		{"cloud sql default", WithCloudSqlConfig(DefaultCloudSqlConfig()), true},
		{"cloud sql mocked", WithCloudSqlConfig(CloudSqlConfig{Enabled: true, Mock: true}), false},
		{"gke default", WithGkeConfig(DefaultGkeConfig()), true},
		{"gke mocked", WithGkeConfig(GkeConfig{Enabled: true, Mock: true}), false},
		{"kafka default", WithKafkaConfig(DefaultKafkaConfig()), true},
		{"kafka mocked", WithKafkaConfig(KafkaConfig{Enabled: true, Mock: true}), false},
		{"kafka disabled", WithKafkaConfig(KafkaConfig{Mock: false}), false},
		// Cloud Functions is metadata only in the emulator: no build, no runtime container.
		{"cloud functions", WithCloudFunctionsConfig(DefaultCloudFunctionsConfig()), false},
		{"gcs", WithGcsConfig(DefaultGcsConfig()), false},
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
	off := map[string]string{}
	for _, ss := range gcpDescriptor.SocketServices {
		off[gcpDescriptor.ServiceEnv(ss.Token, "ENABLED")] = "false"
	}
	tests := []struct {
		name string
		env  map[string]string
		want bool
	}{
		{"nothing set: the defaults need it", nil, true},
		{"only services without containers", map[string]string{"FLOCI_GCP_SERVICES_CLOUDFUNCTIONS_ENABLED": "true"}, false},
		{"gke enabled, mock switch unset", map[string]string{"FLOCI_GCP_SERVICES_GKE_ENABLED": "true"}, true},
		{"gke enabled, mocked", map[string]string{"FLOCI_GCP_SERVICES_GKE_ENABLED": "true", "FLOCI_GCP_SERVICES_GKE_MOCK": "TRUE"}, false},
		{"bigquery enabled", map[string]string{"FLOCI_GCP_SERVICES_BIGQUERY_ENABLED": "true"}, true},
		{"bigquery with its own floci-duck", map[string]string{"FLOCI_GCP_SERVICES_BIGQUERY_ENABLED": "true",
			"FLOCI_GCP_SERVICES_BIGQUERY_DUCK_URL": "http://duck:3000"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := map[string]string{}
			if tt.env != nil {
				env = maps.Clone(off)
				maps.Copy(env, tt.env)
			}
			if got := gcpDescriptor.DockerSocketRequired(env); got != tt.want {
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
			if name, ok := strings.CutPrefix(k, "FLOCI_GCP_SERVICES_"+token+"_"); ok {
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
		{"enabled only", "GCS", []Option{WithGcsConfig(GcsConfig{Enabled: true})}, map[string]string{"ENABLED": "true"}},
		{"disabled", "PUBSUB", []Option{WithPubSubConfig(PubSubConfig{})}, map[string]string{"ENABLED": "false"}},
		{"mocked", "CLOUDSQL", []Option{WithCloudSqlConfig(CloudSqlConfig{Enabled: true, Mock: true})},
			map[string]string{"ENABLED": "true", "MOCK": "true"}},
		{"not mocked", "KAFKA", []Option{WithKafkaConfig(KafkaConfig{Enabled: true})},
			map[string]string{"ENABLED": "true", "MOCK": "false"}},
		{"disabled drops the mock switch", "GKE", []Option{WithGkeConfig(GkeConfig{Mock: true})}, map[string]string{"ENABLED": "false"}},
		{"applied twice", "CLOUDRUN", []Option{WithCloudRunConfig(CloudRunConfig{Enabled: true, Mock: true}), WithCloudRunConfig(CloudRunConfig{Enabled: true, Mock: true})},
			map[string]string{"ENABLED": "true", "MOCK": "true"}},
		{"enabled then disabled", "KAFKA", []Option{WithKafkaConfig(KafkaConfig{Enabled: true, Mock: true}), WithKafkaConfig(KafkaConfig{})},
			map[string]string{"ENABLED": "false"}},
		{"bigquery", "BIGQUERY", []Option{WithBigQueryConfig(BigQueryConfig{Enabled: true, DuckURL: "http://duck:3000"})},
			map[string]string{"ENABLED": "true", "MOCK": "false", "DUCK_URL": "http://duck:3000"}},
		{"bigquery url dropped by a later config", "BIGQUERY", []Option{
			WithBigQueryConfig(BigQueryConfig{Enabled: true, DuckURL: "http://duck:3000"}), WithBigQueryConfig(DefaultBigQueryConfig())},
			map[string]string{"ENABLED": "true", "MOCK": "false"}},
		{"bigquery disabled", "BIGQUERY", []Option{WithBigQueryConfig(BigQueryConfig{DuckURL: "http://duck:3000"})}, map[string]string{"ENABLED": "false"}},
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
		"COMPUTE": WithComputeConfig(ComputeConfig{}), "GCS": WithGcsConfig(GcsConfig{}), "PUBSUB": WithPubSubConfig(PubSubConfig{}),
		"FIRESTORE": WithFirestoreConfig(FirestoreConfig{}), "DATASTORE": WithDatastoreConfig(DatastoreConfig{}),
		"IAM": WithIamConfig(IamConfig{}), "IAMCREDENTIALS": WithIamCredentialsConfig(IamCredentialsConfig{}),
		"STS": WithStsConfig(StsConfig{}), "SECRETMANAGER": WithSecretManagerConfig(SecretManagerConfig{}),
		"LOGGING": WithLoggingConfig(LoggingConfig{}), "KMS": WithKmsConfig(KmsConfig{}), "KAFKA": WithKafkaConfig(KafkaConfig{}),
		"CLOUDSQL": WithCloudSqlConfig(CloudSqlConfig{}), "CLOUDTASKS": WithCloudTasksConfig(CloudTasksConfig{}),
		"CLOUDRUN": WithCloudRunConfig(CloudRunConfig{}), "CLOUDFUNCTIONS": WithCloudFunctionsConfig(CloudFunctionsConfig{}),
		"MONITORING": WithMonitoringConfig(MonitoringConfig{}), "SCHEDULER": WithSchedulerConfig(SchedulerConfig{}),
		"EVENTARC": WithEventarcConfig(EventarcConfig{}), "GKE": WithGkeConfig(GkeConfig{}),
		"SERVICEUSAGE": WithServiceUsageConfig(ServiceUsageConfig{}), "RESOURCEMANAGER": WithResourceManagerConfig(ResourceManagerConfig{}),
		"FIREBASEAUTH": WithFirebaseAuthConfig(FirebaseAuthConfig{}), "BIGQUERY": WithBigQueryConfig(BigQueryConfig{}),
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
			if got := c.envVars["FLOCI_GCP_SERVICES_"+other.token+"_ENABLED"]; got != want {
				t.Errorf("after disabling %s: %s enabled = %q, want %q", svc.token, other.token, got, want)
			}
		}
	}
}

func TestOptions_ApplyToBuilder(t *testing.T) {
	c := build(WithDedicatedNetwork(), WithDockerSocket(false), WithFirestoreConfig(FirestoreConfig{}))
	if got := c.envVars["FLOCI_GCP_SERVICES_FIRESTORE_ENABLED"]; got != "false" {
		t.Errorf("FLOCI_GCP_SERVICES_FIRESTORE_ENABLED = %q, want false", got)
	}
	if !c.dedicatedNetwork {
		t.Error("WithDedicatedNetwork option did not set the dedicated network")
	}
	if c.needsDockerSocket(c.envVars) {
		t.Error("WithDockerSocket(false) option did not override auto-detection")
	}
	if !c.request().DedicatedNetwork || !slices.Equal(c.request().Ports, []int{4588}) {
		t.Errorf("request = %+v", c.request())
	}
}

// WithProjectID sets the emulator's default project; a generic WithEnv still overrides it.
func TestOptions_ProjectID(t *testing.T) {
	if _, set := newBuilder().envVars["FLOCI_GCP_DEFAULT_PROJECT_ID"]; set {
		t.Error("the default project should be left to the emulator")
	}
	c := build(WithProjectID("my-project"))
	if _, env := buildRequest(t, c); env["FLOCI_GCP_DEFAULT_PROJECT_ID"] != "my-project" {
		t.Errorf("env = %v", env)
	}
	_, env := buildRequest(t, c, testcontainers.WithEnv(map[string]string{"FLOCI_GCP_DEFAULT_PROJECT_ID": "other"}))
	if env["FLOCI_GCP_DEFAULT_PROJECT_ID"] != "other" {
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

// The fixed default reaches the container, and a generic WithEnv overrides it. The GKE endpoint
// mode is left to the emulator: its default, "host", already advertises an address the test
// process can reach.
func TestRequest_DefaultEnvIsOverridable(t *testing.T) {
	_, env := buildRequest(t, newBuilder())
	if env["FLOCI_GCP_TLS_HTTPS_PORT"] != "0" {
		t.Fatalf("default env missing: %v", env)
	}
	if mode, set := env["FLOCI_GCP_SERVICES_GKE_ENDPOINT_MODE"]; set {
		t.Fatalf("the GKE endpoint mode should be left to the emulator, got %q", mode)
	}
	_, env = buildRequest(t, newBuilder(), testcontainers.WithEnv(map[string]string{"FLOCI_GCP_TLS_HTTPS_PORT": "8443"}))
	if env["FLOCI_GCP_TLS_HTTPS_PORT"] != "8443" {
		t.Fatalf("caller's env did not win: %v", env)
	}
}

// Mocking every container-backed service through a generic env override must drop the socket,
// and un-mocking one must bring it back.
func TestRequest_GenericEnvMockingServicesDropsSocket(t *testing.T) {
	env := map[string]string{}
	for _, svc := range gcpDescriptor.SocketServices {
		env["FLOCI_GCP_SERVICES_"+svc.Token+"_MOCK"] = "true"
	}
	if req, _ := buildRequest(t, newBuilder(), testcontainers.WithEnv(env)); mountsDockerSocket(req) {
		t.Fatal("expected no socket once every container-backed service is mocked through testcontainers.WithEnv")
	}
	env["FLOCI_GCP_SERVICES_CLOUDSQL_MOCK"] = "false"
	if req, _ := buildRequest(t, newBuilder(), testcontainers.WithEnv(env)); !mountsDockerSocket(req) {
		t.Fatal("expected the socket with Cloud SQL not mocked")
	}
}

// Disabling every container-backed service through a generic env override must drop the socket.
func TestRequest_GenericEnvDisablingServicesDropsSocket(t *testing.T) {
	env := map[string]string{}
	for _, svc := range gcpDescriptor.SocketServices {
		env["FLOCI_GCP_SERVICES_"+svc.Token+"_ENABLED"] = "false"
	}
	req, _ := buildRequest(t, newBuilder(), testcontainers.WithEnv(env))
	if mountsDockerSocket(req) {
		t.Fatal("expected no socket once every container-backed service is disabled through testcontainers.WithEnv")
	}
}

// Each start gets its own resource namespace; a namespace the caller chose is kept.
func TestRequest_FreshNamespacePerStart(t *testing.T) {
	key := gcpDescriptor.ResourceNamespaceEnv()
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

// The emulator variables follow the formats in the floci-gcp documentation: a bare host:port,
// except STORAGE_EMULATOR_HOST, which is a URL.
func TestContainer_EmulatorEnv(t *testing.T) {
	c := &Container{projectID: "my-project"}
	if c.GetEndpoint() != "" {
		t.Errorf("a container not started by Run has no endpoint")
	}
	if c.GetProjectID() != "my-project" {
		t.Errorf("project = %q", c.GetProjectID())
	}
	want := []string{"DATASTORE_EMULATOR_HOST", "FIREBASE_AUTH_EMULATOR_HOST", "FIRESTORE_EMULATOR_HOST", "GOOGLE_CLOUD_PROJECT",
		"KMS_EMULATOR_HOST", "LOGGING_EMULATOR_HOST", "PUBSUB_EMULATOR_HOST", "SECRET_MANAGER_EMULATOR_HOST",
		"STORAGE_EMULATOR_HOST", "STORAGE_EMULATOR_HOST_GRPC"}
	env := c.GetEmulatorEnv()
	if got := slices.Sorted(maps.Keys(env)); !slices.Equal(got, want) {
		t.Errorf("variables = %v, want %v", got, want)
	}
	if env["GOOGLE_CLOUD_PROJECT"] != "my-project" {
		t.Errorf("GOOGLE_CLOUD_PROJECT = %q", env["GOOGLE_CLOUD_PROJECT"])
	}
}
