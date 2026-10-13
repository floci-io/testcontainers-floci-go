package flociaz

// Package-level options for Run.

// WithDedicatedNetwork creates a dedicated Docker network for container-backed services
// (Functions, the databases, Redis, ...) to talk to Floci; Terminate removes it.
func WithDedicatedNetwork() Option {
	return func(c *builder) { c.dedicatedNetwork = true }
}

// WithDockerSocket overrides whether the host Docker socket is mounted. By default it is mounted
// only while an enabled, non-mocked service spawns sibling containers (ACR, AKS, Event Hubs,
// Functions, the databases, Redis, ...); pass false on hosts where the socket cannot be mounted,
// true to always mount it.
func WithDockerSocket(enabled bool) Option {
	return func(c *builder) { c.dockerSocketOverride = &enabled }
}

// WithAciConfig applies Container Instances service configuration.
func WithAciConfig(cfg AciConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithAcrConfig applies Container Registry service configuration.
func WithAcrConfig(cfg AcrConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithAksConfig applies Kubernetes Service (AKS) service configuration.
func WithAksConfig(cfg AksConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithApimConfig applies API Management service configuration.
func WithApimConfig(cfg ApimConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithAppConfigConfig applies App Configuration service configuration.
func WithAppConfigConfig(cfg AppConfigConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithArmConfig applies Resource Manager service configuration.
func WithArmConfig(cfg ArmConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithBlobConfig applies Blob Storage service configuration.
func WithBlobConfig(cfg BlobConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithContainerAppsConfig applies Container Apps service configuration.
func WithContainerAppsConfig(cfg ContainerAppsConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithCosmosConfig applies Cosmos DB service configuration.
func WithCosmosConfig(cfg CosmosConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithEmailConfig applies Communication Services Email service configuration.
func WithEmailConfig(cfg EmailConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithEntraConfig applies Microsoft Entra ID service configuration.
func WithEntraConfig(cfg EntraConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithEventGridConfig applies Event Grid service configuration.
func WithEventGridConfig(cfg EventGridConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithEventHubConfig applies Event Hubs service configuration.
func WithEventHubConfig(cfg EventHubConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithFunctionsConfig applies Functions service configuration.
func WithFunctionsConfig(cfg FunctionsConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithGraphConfig applies Microsoft Graph service configuration.
func WithGraphConfig(cfg GraphConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithKeyVaultConfig applies Key Vault service configuration.
func WithKeyVaultConfig(cfg KeyVaultConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithManagedIdentityConfig applies Managed Identity service configuration.
func WithManagedIdentityConfig(cfg ManagedIdentityConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithMariaDbConfig applies Database for MariaDB service configuration.
func WithMariaDbConfig(cfg MariaDbConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithMonitorConfig applies Monitor service configuration.
func WithMonitorConfig(cfg MonitorConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithMySqlConfig applies Database for MySQL service configuration.
func WithMySqlConfig(cfg MySqlConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithNetworkConfig applies Virtual Network service configuration.
func WithNetworkConfig(cfg NetworkConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithPolicyConfig applies Policy service configuration.
func WithPolicyConfig(cfg PolicyConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithPostgresConfig applies Database for PostgreSQL service configuration.
func WithPostgresConfig(cfg PostgresConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithQueueConfig applies Queue Storage service configuration.
func WithQueueConfig(cfg QueueConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithRedisConfig applies Cache for Redis service configuration.
func WithRedisConfig(cfg RedisConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithServiceBusConfig applies Service Bus service configuration.
func WithServiceBusConfig(cfg ServiceBusConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithSignalRConfig applies SignalR Service service configuration.
func WithSignalRConfig(cfg SignalRConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithSqlConfig applies SQL Database service configuration.
func WithSqlConfig(cfg SqlConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithTableConfig applies Table Storage service configuration.
func WithTableConfig(cfg TableConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithVmConfig applies Virtual Machines service configuration.
func WithVmConfig(cfg VmConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}
