package flociaz

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
		WithAciConfig(AciConfig{}),
		WithAcrConfig(AcrConfig{}),
		WithAksConfig(AksConfig{}),
		WithContainerAppsConfig(ContainerAppsConfig{}),
		WithCosmosConfig(CosmosConfig{}),
		WithEventHubConfig(EventHubConfig{}),
		WithFunctionsConfig(FunctionsConfig{}),
		WithMariaDbConfig(MariaDbConfig{}),
		WithMySqlConfig(MySqlConfig{}),
		WithPostgresConfig(PostgresConfig{}),
		WithRedisConfig(RedisConfig{}),
		WithServiceBusConfig(ServiceBusConfig{}),
		WithSqlConfig(SqlConfig{}),
		WithVmConfig(VmConfig{}),
	}
}

func needsSocket(opts ...Option) bool {
	c := build(append(disableSocketServices(), opts...)...)
	return c.needsDockerSocket(c.envVars)
}

func TestDescriptor(t *testing.T) {
	d := azDescriptor
	for name, got := range map[string][2]string{
		"image":         {d.DefaultImage, "floci/floci-az:latest"},
		"port":          {d.PortSpec(), "4577/tcp"},
		"env prefix":    {d.EnvPrefix, "FLOCI_AZ_"},
		"health":        {d.HealthPath, "/_floci/health"},
		"reset":         {d.ResetPath, "/_admin/reset"},
		"log level":     {d.LogLevelEnv, "QUARKUS_LOG_CATEGORY__IO_FLOCI_AZ__LEVEL"},
		"network":       {d.NetworkEnv(), "FLOCI_AZ_SERVICES_DOCKER_NETWORK"},
		"namespace":     {d.ResourceNamespaceEnv(), "FLOCI_AZ_DOCKER_RESOURCE_NAMESPACE"},
		"service env":   {d.ServiceEnv("BLOB", "ENABLED"), "FLOCI_AZ_SERVICES_BLOB_ENABLED"},
		"mock setting":  {d.ServiceEnv("ACI", d.MockSetting), "FLOCI_AZ_SERVICES_ACI_MOCKED"},
		"subscription":  {DefaultSubscriptionID, "00000000-0000-0000-0000-000000000001"},
		"tenant":        {DefaultTenantID, "00000000-0000-0000-0000-000000000002"},
		"account name":  {DefaultAccountName, "devstoreaccount1"},
		"account key":   {DefaultAccountKey, "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw=="},
		"default image": {newBuilder().image, "floci/floci-az:latest"},
	} {
		if got[0] != got[1] {
			t.Errorf("%s = %q, want %q", name, got[0], got[1])
		}
	}
}

// The default environment: the two fixed settings, a generated namespace, and for every service
// its enabled switch plus, where it has one, its mock switch at the emulator's default.
func TestDefaultEnv(t *testing.T) {
	env := newBuilder().envVars
	want := map[string]string{
		"FLOCI_AZ_DOCKER_ENDPOINT_MODE": "published",
		"FLOCI_AZ_TLS_HTTPS_PORT":       "0",
		// The settings socket detection reads, at the emulator's defaults.
		"FLOCI_AZ_SERVICES_COSMOS_ENGINES_STARTUP":            "on-demand",
		"FLOCI_AZ_SERVICES_COSMOS_ENGINES_NOSQL_ENABLED":      "false",
		"FLOCI_AZ_SERVICES_COSMOS_ENGINES_MONGODB_ENABLED":    "false",
		"FLOCI_AZ_SERVICES_COSMOS_ENGINES_POSTGRESQL_ENABLED": "false",
		"FLOCI_AZ_SERVICES_COSMOS_ENGINES_CASSANDRA_ENABLED":  "false",
		"FLOCI_AZ_SERVICES_COSMOS_ENGINES_GREMLIN_ENABLED":    "false",
		"FLOCI_AZ_SERVICES_COSMOS_ENGINES_TABLE_ENABLED":      "false",
		"FLOCI_AZ_SERVICES_SQL_ACCEPT_EULA":                   "N",
	}
	for _, svc := range services {
		want["FLOCI_AZ_SERVICES_"+svc.token+"_ENABLED"] = "true"
		if svc.mockable {
			want["FLOCI_AZ_SERVICES_"+svc.token+"_MOCKED"] = boolStr(svc.mockedByDefault)
		}
	}
	got := maps.Clone(env)
	ns := got["FLOCI_AZ_DOCKER_RESOURCE_NAMESPACE"]
	if !strings.HasPrefix(ns, "tc-") || len(ns) != len("tc-")+8 {
		t.Errorf("namespace = %q, want tc- plus 8 hex characters", ns)
	}
	delete(got, "FLOCI_AZ_DOCKER_RESOURCE_NAMESPACE")
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

// The service list is the one the emulator's ServicesConfig declares, with its mock defaults.
func TestServices_MatchTheEmulator(t *testing.T) {
	want := map[string]string{ // token -> "" (no mock switch) or the mock default
		"ACI": "true", "ACR": "false", "AKS": "false", "APIM": "", "APP_CONFIG": "", "ARM": "", "BLOB": "",
		"CONTAINER_APPS": "true", "COSMOS": "false", "EMAIL": "", "ENTRA": "", "EVENT_GRID": "",
		"EVENT_HUB": "false", "FUNCTIONS": "false", "GRAPH": "", "KEY_VAULT": "", "MANAGED_IDENTITY": "",
		"MARIA_DB": "false", "MONITOR": "", "MYSQL": "false", "NETWORK": "", "POLICY": "", "POSTGRES": "false",
		"QUEUE": "", "REDIS": "false", "SERVICE_BUS": "true", "SIGNALR": "", "SQL": "", "TABLE": "", "VM": "true",
	}
	got := map[string]string{}
	for _, svc := range services {
		if svc.mockable {
			got[svc.token] = boolStr(svc.mockedByDefault)
		} else {
			got[svc.token] = ""
		}
	}
	if !maps.Equal(got, want) {
		t.Fatalf("services = %v, want %v", got, want)
	}
}

// Every socket service is a known service, and its mock switch agrees with the service list.
func TestSocketTable_MatchesServices(t *testing.T) {
	byToken := map[string]service{}
	for _, svc := range services {
		byToken[svc.token] = svc
	}
	var tokens []string
	for _, ss := range azDescriptor.SocketServices {
		tokens = append(tokens, ss.Token)
		svc, ok := byToken[ss.Token]
		if !ok {
			t.Errorf("socket service %s is not in the service list", ss.Token)
			continue
		}
		if ss.Mockable != svc.mockable || ss.MockedByDefault != svc.mockedByDefault {
			t.Errorf("%s: socket table says mockable=%v mockedByDefault=%v, service list says %v %v",
				ss.Token, ss.Mockable, ss.MockedByDefault, svc.mockable, svc.mockedByDefault)
		}
	}
	want := []string{"ACI", "ACR", "AKS", "CONTAINER_APPS", "COSMOS", "EVENT_HUB", "FUNCTIONS", "MARIA_DB",
		"MYSQL", "POSTGRES", "REDIS", "SERVICE_BUS", "SQL", "VM"}
	if !slices.Equal(tokens, want) {
		t.Errorf("socket services = %v, want %v", tokens, want)
	}
	// Every service with a mock switch spawns containers when it is not mocked.
	for _, svc := range services {
		if svc.mockable && !slices.Contains(tokens, svc.token) {
			t.Errorf("%s has a mock switch but is not a socket service", svc.token)
		}
	}
}

// With the emulator's defaults, these services alone make the container need the socket.
func TestDockerSocket_ServicesNeedingItByDefault(t *testing.T) {
	var got []string
	for _, ss := range azDescriptor.SocketServices {
		env := map[string]string{}
		for _, other := range azDescriptor.SocketServices {
			env[azDescriptor.ServiceEnv(other.Token, "ENABLED")] = boolStr(other.Token == ss.Token)
		}
		if azDescriptor.DockerSocketRequired(env) {
			got = append(got, ss.Token)
		}
	}
	want := []string{"ACR", "AKS", "EVENT_HUB", "FUNCTIONS", "MARIA_DB", "MYSQL", "POSTGRES", "REDIS"}
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
		{"aci default (mocked)", WithAciConfig(DefaultAciConfig()), false},
		{"aci unmocked", WithAciConfig(AciConfig{Enabled: true}), true},
		{"aci disabled", WithAciConfig(AciConfig{}), false},
		{"acr default", WithAcrConfig(DefaultAcrConfig()), true},
		{"acr mocked", WithAcrConfig(AcrConfig{Enabled: true, Mocked: true}), false},
		{"aks default", WithAksConfig(DefaultAksConfig()), true},
		{"aks mocked", WithAksConfig(AksConfig{Enabled: true, Mocked: true}), false},
		{"container apps default (mocked)", WithContainerAppsConfig(DefaultContainerAppsConfig()), false},
		{"container apps unmocked", WithContainerAppsConfig(ContainerAppsConfig{Enabled: true}), true},
		{"event hub default", WithEventHubConfig(DefaultEventHubConfig()), true},
		{"event hub mocked", WithEventHubConfig(EventHubConfig{Enabled: true, Mocked: true}), false},
		{"functions default", WithFunctionsConfig(DefaultFunctionsConfig()), true},
		{"functions mocked", WithFunctionsConfig(FunctionsConfig{Enabled: true, Mocked: true}), false},
		{"mariadb default", WithMariaDbConfig(DefaultMariaDbConfig()), true},
		{"mariadb mocked", WithMariaDbConfig(MariaDbConfig{Enabled: true, Mocked: true}), false},
		{"mysql default", WithMySqlConfig(DefaultMySqlConfig()), true},
		{"mysql mocked", WithMySqlConfig(MySqlConfig{Enabled: true, Mocked: true}), false},
		{"postgres default", WithPostgresConfig(DefaultPostgresConfig()), true},
		{"postgres mocked", WithPostgresConfig(PostgresConfig{Enabled: true, Mocked: true}), false},
		{"redis default", WithRedisConfig(DefaultRedisConfig()), true},
		{"redis mocked", WithRedisConfig(RedisConfig{Enabled: true, Mocked: true}), false},
		{"service bus default (mocked)", WithServiceBusConfig(DefaultServiceBusConfig()), false},
		{"service bus unmocked", WithServiceBusConfig(ServiceBusConfig{Enabled: true}), true},
		{"vm default (mocked)", WithVmConfig(DefaultVmConfig()), false},
		{"vm unmocked", WithVmConfig(VmConfig{Enabled: true}), true},

		// Cosmos: only the Docker-backed engines spawn a container, and none in "disabled" startup.
		{"cosmos default (no engine)", WithCosmosConfig(DefaultCosmosConfig()), false},
		{"cosmos nosql and table engines", WithCosmosConfig(CosmosConfig{Enabled: true,
			Nosql: CosmosApiConfig{Enabled: true}, Table: CosmosApiConfig{Enabled: true}}), false},
		{"cosmos mongodb engine", WithCosmosConfig(CosmosConfig{Enabled: true, Mongodb: CosmosApiConfig{Enabled: true}}), true},
		{"cosmos postgresql engine", WithCosmosConfig(CosmosConfig{Enabled: true, Postgresql: CosmosApiConfig{Enabled: true}}), true},
		{"cosmos cassandra engine", WithCosmosConfig(CosmosConfig{Enabled: true, Cassandra: CosmosApiConfig{Enabled: true}}), true},
		{"cosmos gremlin engine, eager", WithCosmosConfig(CosmosConfig{Enabled: true, EngineStartup: "eager",
			Gremlin: CosmosApiConfig{Enabled: true}}), true},
		{"cosmos engine, startup disabled", WithCosmosConfig(CosmosConfig{Enabled: true, EngineStartup: "disabled",
			Mongodb: CosmosApiConfig{Enabled: true}}), false},
		{"cosmos engine, mocked", WithCosmosConfig(CosmosConfig{Enabled: true, Mocked: true,
			Mongodb: CosmosApiConfig{Enabled: true}}), false},
		{"cosmos engine, disabled", WithCosmosConfig(CosmosConfig{Mongodb: CosmosApiConfig{Enabled: true}}), false},

		// SQL: only the "managed" data plane runs SQL Server in a sibling container.
		{"sql default (EULA not accepted)", WithSqlConfig(DefaultSqlConfig()), false},
		{"sql EULA accepted", WithSqlConfig(SqlConfig{Enabled: true, AcceptEula: "Y"}), true},
		{"sql managed", WithSqlConfig(SqlConfig{Enabled: true, DataPlaneProvider: SqlDataPlaneManaged}), true},
		{"sql none, EULA accepted", WithSqlConfig(SqlConfig{Enabled: true, AcceptEula: "Y", DataPlaneProvider: SqlDataPlaneNone}), false},
		{"sql external", WithSqlConfig(SqlConfig{Enabled: true, AcceptEula: "Y", DataPlaneProvider: SqlDataPlaneExternal}), false},
		{"sql managed, disabled", WithSqlConfig(SqlConfig{DataPlaneProvider: SqlDataPlaneManaged}), false},
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
// a missing mock switch means the emulator's default, and SQL honours its deprecated MOCKED switch.
func TestDockerSocket_FromRawEnv(t *testing.T) {
	off := map[string]string{}
	for _, ss := range azDescriptor.SocketServices {
		off[azDescriptor.ServiceEnv(ss.Token, "ENABLED")] = "false"
	}
	tests := []struct {
		name string
		env  map[string]string
		want bool
	}{
		{"nothing set: the defaults need it", nil, true},
		{"vm enabled, mock switch unset", map[string]string{"FLOCI_AZ_SERVICES_VM_ENABLED": "true"}, false},
		{"vm enabled, unmocked", map[string]string{"FLOCI_AZ_SERVICES_VM_ENABLED": "true", "FLOCI_AZ_SERVICES_VM_MOCKED": "false"}, true},
		{"redis enabled, mock switch unset", map[string]string{"FLOCI_AZ_SERVICES_REDIS_ENABLED": "true"}, true},
		{"redis enabled, mocked", map[string]string{"FLOCI_AZ_SERVICES_REDIS_ENABLED": "true", "FLOCI_AZ_SERVICES_REDIS_MOCKED": "TRUE"}, false},
		{"sql legacy unmocked", map[string]string{"FLOCI_AZ_SERVICES_SQL_ENABLED": "true", "FLOCI_AZ_SERVICES_SQL_MOCKED": "false"}, true},
		{"sql legacy mocked, EULA accepted", map[string]string{"FLOCI_AZ_SERVICES_SQL_ENABLED": "true",
			"FLOCI_AZ_SERVICES_SQL_MOCKED": "true", "FLOCI_AZ_SERVICES_SQL_ACCEPT_EULA": "Y"}, false},
		{"sql provider wins over legacy", map[string]string{"FLOCI_AZ_SERVICES_SQL_ENABLED": "true",
			"FLOCI_AZ_SERVICES_SQL_MOCKED": "true", "FLOCI_AZ_SERVICES_SQL_DATA_PLANE_PROVIDER": "MANAGED"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := map[string]string{}
			if tt.env != nil {
				env = maps.Clone(off)
				maps.Copy(env, tt.env)
			}
			if got := azDescriptor.DockerSocketRequired(env); got != tt.want {
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
			if name, ok := strings.CutPrefix(k, "FLOCI_AZ_SERVICES_"+token+"_"); ok {
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
		{"enabled only", "BLOB", []Option{WithBlobConfig(BlobConfig{Enabled: true})}, map[string]string{"ENABLED": "true"}},
		{"disabled", "QUEUE", []Option{WithQueueConfig(QueueConfig{})}, map[string]string{"ENABLED": "false"}},
		{"mocked", "REDIS", []Option{WithRedisConfig(RedisConfig{Enabled: true, Mocked: true})},
			map[string]string{"ENABLED": "true", "MOCKED": "true"}},
		{"unmocked", "SERVICE_BUS", []Option{WithServiceBusConfig(ServiceBusConfig{Enabled: true})},
			map[string]string{"ENABLED": "true", "MOCKED": "false"}},
		{"disabled drops the mock switch", "VM", []Option{WithVmConfig(VmConfig{Mocked: true})}, map[string]string{"ENABLED": "false"}},
		{"applied twice", "ACI", []Option{WithAciConfig(AciConfig{Enabled: true}), WithAciConfig(AciConfig{Enabled: true})},
			map[string]string{"ENABLED": "true", "MOCKED": "false"}},
		{"enabled then disabled", "ACR", []Option{WithAcrConfig(AcrConfig{Enabled: true, Mocked: true}), WithAcrConfig(AcrConfig{})},
			map[string]string{"ENABLED": "false"}},
		{"cosmos", "COSMOS", []Option{WithCosmosConfig(CosmosConfig{Enabled: true, EngineStartup: "eager", Mongodb: CosmosApiConfig{Enabled: true}})},
			map[string]string{"ENABLED": "true", "MOCKED": "false", "ENGINES_STARTUP": "eager",
				"ENGINES_NOSQL_ENABLED": "false", "ENGINES_MONGODB_ENABLED": "true", "ENGINES_POSTGRESQL_ENABLED": "false",
				"ENGINES_CASSANDRA_ENABLED": "false", "ENGINES_GREMLIN_ENABLED": "false", "ENGINES_TABLE_ENABLED": "false"}},
		{"cosmos without a startup mode leaves the emulator's", "COSMOS", []Option{WithCosmosConfig(CosmosConfig{Enabled: true, Mocked: true})},
			map[string]string{"ENABLED": "true", "MOCKED": "true",
				"ENGINES_NOSQL_ENABLED": "false", "ENGINES_MONGODB_ENABLED": "false", "ENGINES_POSTGRESQL_ENABLED": "false",
				"ENGINES_CASSANDRA_ENABLED": "false", "ENGINES_GREMLIN_ENABLED": "false", "ENGINES_TABLE_ENABLED": "false"}},
		{"cosmos disabled", "COSMOS", []Option{WithCosmosConfig(CosmosConfig{EngineStartup: "eager"})}, map[string]string{"ENABLED": "false"}},
		{"sql", "SQL", []Option{WithSqlConfig(SqlConfig{Enabled: true, AcceptEula: "Y", DataPlaneProvider: SqlDataPlaneManaged})},
			map[string]string{"ENABLED": "true", "ACCEPT_EULA": "Y", "DATA_PLANE_PROVIDER": "managed"}},
		{"sql provider dropped by a later config", "SQL", []Option{
			WithSqlConfig(SqlConfig{Enabled: true, DataPlaneProvider: SqlDataPlaneExternal}), WithSqlConfig(DefaultSqlConfig())},
			map[string]string{"ENABLED": "true", "ACCEPT_EULA": "N"}},
		{"sql disabled", "SQL", []Option{WithSqlConfig(SqlConfig{AcceptEula: "Y"})}, map[string]string{"ENABLED": "false"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := serviceEnv(build(tt.opts...), tt.token); !maps.Equal(got, tt.want) {
				t.Fatalf("env = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOptions_ApplyToBuilder(t *testing.T) {
	c := build(WithDedicatedNetwork(), WithDockerSocket(false), WithTableConfig(TableConfig{}))
	if got := c.envVars["FLOCI_AZ_SERVICES_TABLE_ENABLED"]; got != "false" {
		t.Errorf("FLOCI_AZ_SERVICES_TABLE_ENABLED = %q, want false", got)
	}
	if !c.dedicatedNetwork {
		t.Error("WithDedicatedNetwork option did not set the dedicated network")
	}
	if c.needsDockerSocket(c.envVars) {
		t.Error("WithDockerSocket(false) option did not override auto-detection")
	}
	if !c.request().DedicatedNetwork || !slices.Equal(c.request().Ports, []int{4577}) {
		t.Errorf("request = %+v", c.request())
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

// The fixed defaults reach the container, and a generic WithEnv overrides them.
func TestRequest_DefaultEnvIsOverridable(t *testing.T) {
	_, env := buildRequest(t, newBuilder())
	if env["FLOCI_AZ_DOCKER_ENDPOINT_MODE"] != "published" || env["FLOCI_AZ_TLS_HTTPS_PORT"] != "0" {
		t.Fatalf("default env missing: %v", env)
	}
	_, env = buildRequest(t, newBuilder(), testcontainers.WithEnv(map[string]string{
		"FLOCI_AZ_DOCKER_ENDPOINT_MODE": "auto",
		"FLOCI_AZ_TLS_HTTPS_PORT":       "8443",
	}))
	if env["FLOCI_AZ_DOCKER_ENDPOINT_MODE"] != "auto" || env["FLOCI_AZ_TLS_HTTPS_PORT"] != "8443" {
		t.Fatalf("caller's env did not win: %v", env)
	}
}

// Un-mocking a default-mocked service through a generic env override must mount the socket.
func TestRequest_GenericEnvUnmockingServiceMountsSocket(t *testing.T) {
	c := build(append(disableSocketServices(), WithVmConfig(DefaultVmConfig()))...)
	if req, _ := buildRequest(t, c); mountsDockerSocket(req) {
		t.Fatal("a mocked VM service must not mount the socket")
	}
	req, _ := buildRequest(t, c, testcontainers.WithEnv(map[string]string{"FLOCI_AZ_SERVICES_VM_MOCKED": "false"}))
	if !mountsDockerSocket(req) {
		t.Fatal("expected the socket once VM is un-mocked through testcontainers.WithEnv")
	}
}

// Disabling every container-backed service through a generic env override must drop the socket.
func TestRequest_GenericEnvDisablingServicesDropsSocket(t *testing.T) {
	env := map[string]string{}
	for _, svc := range azDescriptor.SocketServices {
		env["FLOCI_AZ_SERVICES_"+svc.Token+"_ENABLED"] = "false"
	}
	req, _ := buildRequest(t, newBuilder(), testcontainers.WithEnv(env))
	if mountsDockerSocket(req) {
		t.Fatal("expected no socket once every container-backed service is disabled through testcontainers.WithEnv")
	}
}

// Each start gets its own resource namespace; a namespace the caller chose is kept.
func TestRequest_FreshNamespacePerStart(t *testing.T) {
	key := azDescriptor.ResourceNamespaceEnv()
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

// The connection helpers follow the formats of the Java module's FlociAzContainer.
func TestContainer_ConnectionHelpers(t *testing.T) {
	c := &Container{tenantID: DefaultTenantID, accountKey: "a2V5"}
	if c.GetEndpoint() != "" || c.GetHTTPSEndpoint() != "" {
		t.Errorf("a container not started by Run has no endpoint")
	}
	if c.GetSubscriptionID() != DefaultSubscriptionID || c.GetTenantID() != DefaultTenantID ||
		c.GetAccountName() != "devstoreaccount1" || c.GetAccountKey() != "a2V5" {
		t.Errorf("identity = %q %q %q %q", c.GetSubscriptionID(), c.GetTenantID(), c.GetAccountName(), c.GetAccountKey())
	}
	want := "DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=a2V5" +
		";BlobEndpoint=/devstoreaccount1;QueueEndpoint=/devstoreaccount1-queue;TableEndpoint=/devstoreaccount1-table;"
	if got := c.GetStorageConnectionString(); got != want {
		t.Errorf("connection string = %q, want %q", got, want)
	}
}
