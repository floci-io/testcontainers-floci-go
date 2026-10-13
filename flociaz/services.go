package flociaz

import "strings"

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// setting is one service setting other than ENABLED; an empty value leaves the emulator's default.
type setting struct{ name, value string }

// applyService emits <prefix>SERVICES_<token>_ENABLED always and the other settings only while
// the service is enabled, dropping what an earlier config of the same service set.
func (c *builder) applyService(token string, enabled bool, settings ...setting) {
	c.withEnv(azDescriptor.ServiceEnv(token, "ENABLED"), boolStr(enabled))
	for _, s := range settings {
		key := azDescriptor.ServiceEnv(token, s.name)
		if enabled && s.value != "" {
			c.withEnv(key, s.value)
		} else {
			delete(c.envVars, key)
		}
	}
}

// service is one Floci Azure service: its token in FLOCI_AZ_SERVICES_<token>_ENABLED, whether it
// has a mock switch (FLOCI_AZ_SERVICES_<token>_MOCKED) and that switch's default. The list mirrors
// ServicesConfig in the emulator's EmulatorConfig.
type service struct {
	token           string
	mockable        bool
	mockedByDefault bool
}

var services = []service{
	{token: "ACI", mockable: true, mockedByDefault: true},
	{token: "ACR", mockable: true},
	{token: "AKS", mockable: true},
	{token: "APIM"},
	{token: "APP_CONFIG"},
	{token: "ARM"},
	{token: "BLOB"},
	{token: "CONTAINER_APPS", mockable: true, mockedByDefault: true},
	{token: "COSMOS", mockable: true},
	{token: "EMAIL"},
	{token: "ENTRA"},
	{token: "EVENT_GRID"},
	{token: "EVENT_HUB", mockable: true},
	{token: "FUNCTIONS", mockable: true},
	{token: "GRAPH"},
	{token: "KEY_VAULT"},
	{token: "MANAGED_IDENTITY"},
	{token: "MARIA_DB", mockable: true},
	{token: "MONITOR"},
	{token: "MYSQL", mockable: true},
	{token: "NETWORK"},
	{token: "POLICY"},
	{token: "POSTGRES", mockable: true},
	{token: "QUEUE"},
	{token: "REDIS", mockable: true},
	{token: "SERVICE_BUS", mockable: true, mockedByDefault: true},
	{token: "SIGNALR"},
	{token: "SQL"},
	{token: "TABLE"},
	{token: "VM", mockable: true, mockedByDefault: true},
}

// AciConfig configures the Container Instances service. It is mocked by default; with Mocked false it runs in sibling containers.
type AciConfig struct {
	Enabled bool
	Mocked  bool
}

func DefaultAciConfig() AciConfig { return AciConfig{Enabled: true, Mocked: true} }

func (c AciConfig) applyEnvVars(t *builder) {
	t.applyService("ACI", c.Enabled, setting{"MOCKED", boolStr(c.Mocked)})
}

// AcrConfig configures the Container Registry service. With Mocked true it spawns no sibling containers.
type AcrConfig struct {
	Enabled bool
	Mocked  bool
}

func DefaultAcrConfig() AcrConfig { return AcrConfig{Enabled: true, Mocked: false} }

func (c AcrConfig) applyEnvVars(t *builder) {
	t.applyService("ACR", c.Enabled, setting{"MOCKED", boolStr(c.Mocked)})
}

// AksConfig configures the Kubernetes Service (AKS) service. With Mocked true it spawns no sibling containers.
type AksConfig struct {
	Enabled bool
	Mocked  bool
}

func DefaultAksConfig() AksConfig { return AksConfig{Enabled: true, Mocked: false} }

func (c AksConfig) applyEnvVars(t *builder) {
	t.applyService("AKS", c.Enabled, setting{"MOCKED", boolStr(c.Mocked)})
}

// ApimConfig configures the API Management service.
type ApimConfig struct {
	Enabled bool
}

func DefaultApimConfig() ApimConfig { return ApimConfig{Enabled: true} }

func (c ApimConfig) applyEnvVars(t *builder) {
	t.applyService("APIM", c.Enabled)
}

// AppConfigConfig configures the App Configuration service.
type AppConfigConfig struct {
	Enabled bool
}

func DefaultAppConfigConfig() AppConfigConfig { return AppConfigConfig{Enabled: true} }

func (c AppConfigConfig) applyEnvVars(t *builder) {
	t.applyService("APP_CONFIG", c.Enabled)
}

// ArmConfig configures the Resource Manager service.
type ArmConfig struct {
	Enabled bool
}

func DefaultArmConfig() ArmConfig { return ArmConfig{Enabled: true} }

func (c ArmConfig) applyEnvVars(t *builder) {
	t.applyService("ARM", c.Enabled)
}

// BlobConfig configures the Blob Storage service.
type BlobConfig struct {
	Enabled bool
}

func DefaultBlobConfig() BlobConfig { return BlobConfig{Enabled: true} }

func (c BlobConfig) applyEnvVars(t *builder) {
	t.applyService("BLOB", c.Enabled)
}

// ContainerAppsConfig configures the Container Apps service. It is mocked by default; with Mocked false it runs in sibling containers.
type ContainerAppsConfig struct {
	Enabled bool
	Mocked  bool
}

func DefaultContainerAppsConfig() ContainerAppsConfig {
	return ContainerAppsConfig{Enabled: true, Mocked: true}
}

func (c ContainerAppsConfig) applyEnvVars(t *builder) {
	t.applyService("CONTAINER_APPS", c.Enabled, setting{"MOCKED", boolStr(c.Mocked)})
}

// CosmosApiConfig configures one Cosmos DB API engine.
type CosmosApiConfig struct {
	Enabled bool
}

// CosmosConfig configures the Cosmos DB service. EngineStartup is "on-demand" (the default when
// empty), "eager" or "disabled". Every API engine is off by default; the MongoDB, PostgreSQL,
// Cassandra and Gremlin engines run in sibling containers.
type CosmosConfig struct {
	Enabled       bool
	Mocked        bool
	EngineStartup string
	Nosql         CosmosApiConfig
	Mongodb       CosmosApiConfig
	Postgresql    CosmosApiConfig
	Cassandra     CosmosApiConfig
	Gremlin       CosmosApiConfig
	Table         CosmosApiConfig
}

func DefaultCosmosConfig() CosmosConfig {
	return CosmosConfig{Enabled: true, Mocked: false, EngineStartup: "on-demand"}
}

func (c CosmosConfig) applyEnvVars(t *builder) {
	t.applyService("COSMOS", c.Enabled,
		setting{"MOCKED", boolStr(c.Mocked)},
		setting{"ENGINES_STARTUP", c.EngineStartup},
		setting{"ENGINES_NOSQL_ENABLED", boolStr(c.Nosql.Enabled)},
		setting{"ENGINES_MONGODB_ENABLED", boolStr(c.Mongodb.Enabled)},
		setting{"ENGINES_POSTGRESQL_ENABLED", boolStr(c.Postgresql.Enabled)},
		setting{"ENGINES_CASSANDRA_ENABLED", boolStr(c.Cassandra.Enabled)},
		setting{"ENGINES_GREMLIN_ENABLED", boolStr(c.Gremlin.Enabled)},
		setting{"ENGINES_TABLE_ENABLED", boolStr(c.Table.Enabled)},
	)
}

// cosmosDockerEngines are the Cosmos DB API engines that run in sibling containers.
var cosmosDockerEngines = []string{"MONGODB", "POSTGRESQL", "CASSANDRA", "GREMLIN"}

// cosmosNeedsDocker mirrors CosmosConfig.requiresDockerSocket() in the Java module: engines start
// unless the startup mode is "disabled", and only the Docker-backed ones spawn a container.
func cosmosNeedsDocker(setting func(string) string) bool {
	if strings.EqualFold(setting("ENGINES_STARTUP"), "disabled") {
		return false
	}
	for _, api := range cosmosDockerEngines {
		if strings.EqualFold(setting("ENGINES_"+api+"_ENABLED"), "true") {
			return true
		}
	}
	return false
}

// EmailConfig configures the Communication Services Email service.
type EmailConfig struct {
	Enabled bool
}

func DefaultEmailConfig() EmailConfig { return EmailConfig{Enabled: true} }

func (c EmailConfig) applyEnvVars(t *builder) {
	t.applyService("EMAIL", c.Enabled)
}

// EntraConfig configures the Microsoft Entra ID service.
type EntraConfig struct {
	Enabled bool
}

func DefaultEntraConfig() EntraConfig { return EntraConfig{Enabled: true} }

func (c EntraConfig) applyEnvVars(t *builder) {
	t.applyService("ENTRA", c.Enabled)
}

// EventGridConfig configures the Event Grid service.
type EventGridConfig struct {
	Enabled bool
}

func DefaultEventGridConfig() EventGridConfig { return EventGridConfig{Enabled: true} }

func (c EventGridConfig) applyEnvVars(t *builder) {
	t.applyService("EVENT_GRID", c.Enabled)
}

// EventHubConfig configures the Event Hubs service. With Mocked true it spawns no sibling containers.
type EventHubConfig struct {
	Enabled bool
	Mocked  bool
}

func DefaultEventHubConfig() EventHubConfig { return EventHubConfig{Enabled: true, Mocked: false} }

func (c EventHubConfig) applyEnvVars(t *builder) {
	t.applyService("EVENT_HUB", c.Enabled, setting{"MOCKED", boolStr(c.Mocked)})
}

// FunctionsConfig configures the Functions service. With Mocked true it spawns no sibling containers.
type FunctionsConfig struct {
	Enabled bool
	Mocked  bool
}

func DefaultFunctionsConfig() FunctionsConfig { return FunctionsConfig{Enabled: true, Mocked: false} }

func (c FunctionsConfig) applyEnvVars(t *builder) {
	t.applyService("FUNCTIONS", c.Enabled, setting{"MOCKED", boolStr(c.Mocked)})
}

// GraphConfig configures the Microsoft Graph service.
type GraphConfig struct {
	Enabled bool
}

func DefaultGraphConfig() GraphConfig { return GraphConfig{Enabled: true} }

func (c GraphConfig) applyEnvVars(t *builder) {
	t.applyService("GRAPH", c.Enabled)
}

// KeyVaultConfig configures the Key Vault service.
type KeyVaultConfig struct {
	Enabled bool
}

func DefaultKeyVaultConfig() KeyVaultConfig { return KeyVaultConfig{Enabled: true} }

func (c KeyVaultConfig) applyEnvVars(t *builder) {
	t.applyService("KEY_VAULT", c.Enabled)
}

// ManagedIdentityConfig configures the Managed Identity service.
type ManagedIdentityConfig struct {
	Enabled bool
}

func DefaultManagedIdentityConfig() ManagedIdentityConfig {
	return ManagedIdentityConfig{Enabled: true}
}

func (c ManagedIdentityConfig) applyEnvVars(t *builder) {
	t.applyService("MANAGED_IDENTITY", c.Enabled)
}

// MariaDbConfig configures the Database for MariaDB service. With Mocked true it spawns no sibling containers.
type MariaDbConfig struct {
	Enabled bool
	Mocked  bool
}

func DefaultMariaDbConfig() MariaDbConfig { return MariaDbConfig{Enabled: true, Mocked: false} }

func (c MariaDbConfig) applyEnvVars(t *builder) {
	t.applyService("MARIA_DB", c.Enabled, setting{"MOCKED", boolStr(c.Mocked)})
}

// MonitorConfig configures the Monitor service.
type MonitorConfig struct {
	Enabled bool
}

func DefaultMonitorConfig() MonitorConfig { return MonitorConfig{Enabled: true} }

func (c MonitorConfig) applyEnvVars(t *builder) {
	t.applyService("MONITOR", c.Enabled)
}

// MySqlConfig configures the Database for MySQL service. With Mocked true it spawns no sibling containers.
type MySqlConfig struct {
	Enabled bool
	Mocked  bool
}

func DefaultMySqlConfig() MySqlConfig { return MySqlConfig{Enabled: true, Mocked: false} }

func (c MySqlConfig) applyEnvVars(t *builder) {
	t.applyService("MYSQL", c.Enabled, setting{"MOCKED", boolStr(c.Mocked)})
}

// NetworkConfig configures the Virtual Network service.
type NetworkConfig struct {
	Enabled bool
}

func DefaultNetworkConfig() NetworkConfig { return NetworkConfig{Enabled: true} }

func (c NetworkConfig) applyEnvVars(t *builder) {
	t.applyService("NETWORK", c.Enabled)
}

// PolicyConfig configures the Policy service.
type PolicyConfig struct {
	Enabled bool
}

func DefaultPolicyConfig() PolicyConfig { return PolicyConfig{Enabled: true} }

func (c PolicyConfig) applyEnvVars(t *builder) {
	t.applyService("POLICY", c.Enabled)
}

// PostgresConfig configures the Database for PostgreSQL service. With Mocked true it spawns no sibling containers.
type PostgresConfig struct {
	Enabled bool
	Mocked  bool
}

func DefaultPostgresConfig() PostgresConfig { return PostgresConfig{Enabled: true, Mocked: false} }

func (c PostgresConfig) applyEnvVars(t *builder) {
	t.applyService("POSTGRES", c.Enabled, setting{"MOCKED", boolStr(c.Mocked)})
}

// QueueConfig configures the Queue Storage service.
type QueueConfig struct {
	Enabled bool
}

func DefaultQueueConfig() QueueConfig { return QueueConfig{Enabled: true} }

func (c QueueConfig) applyEnvVars(t *builder) {
	t.applyService("QUEUE", c.Enabled)
}

// RedisConfig configures the Cache for Redis service. With Mocked true it spawns no sibling containers.
type RedisConfig struct {
	Enabled bool
	Mocked  bool
}

func DefaultRedisConfig() RedisConfig { return RedisConfig{Enabled: true, Mocked: false} }

func (c RedisConfig) applyEnvVars(t *builder) {
	t.applyService("REDIS", c.Enabled, setting{"MOCKED", boolStr(c.Mocked)})
}

// ServiceBusConfig configures the Service Bus service. It is mocked by default; with Mocked false it runs in sibling containers.
type ServiceBusConfig struct {
	Enabled bool
	Mocked  bool
}

func DefaultServiceBusConfig() ServiceBusConfig { return ServiceBusConfig{Enabled: true, Mocked: true} }

func (c ServiceBusConfig) applyEnvVars(t *builder) {
	t.applyService("SERVICE_BUS", c.Enabled, setting{"MOCKED", boolStr(c.Mocked)})
}

// SignalRConfig configures the SignalR Service service.
type SignalRConfig struct {
	Enabled bool
}

func DefaultSignalRConfig() SignalRConfig { return SignalRConfig{Enabled: true} }

func (c SignalRConfig) applyEnvVars(t *builder) {
	t.applyService("SIGNALR", c.Enabled)
}

// SqlDataPlaneProvider says what serves the data plane of the Azure SQL servers.
type SqlDataPlaneProvider string

const (
	// SqlDataPlaneNone serves the control plane only.
	SqlDataPlaneNone SqlDataPlaneProvider = "none"
	// SqlDataPlaneManaged runs SQL Server in sibling containers.
	SqlDataPlaneManaged SqlDataPlaneProvider = "managed"
	// SqlDataPlaneExternal points at a SQL Server the caller runs.
	SqlDataPlaneExternal SqlDataPlaneProvider = "external"
)

// SqlConfig configures the SQL Database service. With DataPlaneProvider empty the emulator picks
// SqlDataPlaneManaged when AcceptEula is "Y" and SqlDataPlaneNone otherwise ("N" is the default).
type SqlConfig struct {
	Enabled           bool
	AcceptEula        string
	DataPlaneProvider SqlDataPlaneProvider
}

func DefaultSqlConfig() SqlConfig { return SqlConfig{Enabled: true, AcceptEula: "N"} }

func (c SqlConfig) applyEnvVars(t *builder) {
	t.applyService("SQL", c.Enabled,
		setting{"ACCEPT_EULA", c.AcceptEula},
		setting{"DATA_PLANE_PROVIDER", string(c.DataPlaneProvider)},
	)
}

// sqlNeedsDocker reports whether the effective data plane provider is "managed", resolved as the
// emulator does: the explicit provider, then the deprecated MOCKED switch, then the EULA.
func sqlNeedsDocker(setting func(string) string) bool {
	if provider := setting("DATA_PLANE_PROVIDER"); provider != "" {
		return strings.EqualFold(provider, string(SqlDataPlaneManaged))
	}
	if mocked := setting("MOCKED"); mocked != "" {
		return !strings.EqualFold(mocked, "true")
	}
	return strings.EqualFold(setting("ACCEPT_EULA"), "Y")
}

// TableConfig configures the Table Storage service.
type TableConfig struct {
	Enabled bool
}

func DefaultTableConfig() TableConfig { return TableConfig{Enabled: true} }

func (c TableConfig) applyEnvVars(t *builder) {
	t.applyService("TABLE", c.Enabled)
}

// VmConfig configures the Virtual Machines service. It is mocked by default; with Mocked false it runs in sibling containers.
type VmConfig struct {
	Enabled bool
	Mocked  bool
}

func DefaultVmConfig() VmConfig { return VmConfig{Enabled: true, Mocked: true} }

func (c VmConfig) applyEnvVars(t *builder) {
	t.applyService("VM", c.Enabled, setting{"MOCKED", boolStr(c.Mocked)})
}

// applyAllConfigs applies the default config of every service.
func (c *builder) applyAllConfigs() {
	DefaultAciConfig().applyEnvVars(c)
	DefaultAcrConfig().applyEnvVars(c)
	DefaultAksConfig().applyEnvVars(c)
	DefaultApimConfig().applyEnvVars(c)
	DefaultAppConfigConfig().applyEnvVars(c)
	DefaultArmConfig().applyEnvVars(c)
	DefaultBlobConfig().applyEnvVars(c)
	DefaultContainerAppsConfig().applyEnvVars(c)
	DefaultCosmosConfig().applyEnvVars(c)
	DefaultEmailConfig().applyEnvVars(c)
	DefaultEntraConfig().applyEnvVars(c)
	DefaultEventGridConfig().applyEnvVars(c)
	DefaultEventHubConfig().applyEnvVars(c)
	DefaultFunctionsConfig().applyEnvVars(c)
	DefaultGraphConfig().applyEnvVars(c)
	DefaultKeyVaultConfig().applyEnvVars(c)
	DefaultManagedIdentityConfig().applyEnvVars(c)
	DefaultMariaDbConfig().applyEnvVars(c)
	DefaultMonitorConfig().applyEnvVars(c)
	DefaultMySqlConfig().applyEnvVars(c)
	DefaultNetworkConfig().applyEnvVars(c)
	DefaultPolicyConfig().applyEnvVars(c)
	DefaultPostgresConfig().applyEnvVars(c)
	DefaultQueueConfig().applyEnvVars(c)
	DefaultRedisConfig().applyEnvVars(c)
	DefaultServiceBusConfig().applyEnvVars(c)
	DefaultSignalRConfig().applyEnvVars(c)
	DefaultSqlConfig().applyEnvVars(c)
	DefaultTableConfig().applyEnvVars(c)
	DefaultVmConfig().applyEnvVars(c)
}
