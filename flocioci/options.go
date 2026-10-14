package flocioci

// Package-level options for Run.

// WithRegion sets the region the emulator serves (FLOCI_OCI_DEFAULT_REGION), which GetRegion then returns.
func WithRegion(region string) Option {
	return func(c *builder) { c.withEnv(regionEnv, region) }
}

// WithTenancyID sets the emulator's default tenancy (FLOCI_OCI_DEFAULT_TENANCY_ID), the tenancy of
// unsigned requests, which GetTenancyID then returns and the credentials sign as.
func WithTenancyID(tenancyID string) Option {
	return func(c *builder) { c.withEnv(tenancyIDEnv, tenancyID) }
}

// WithNamespace sets the Object Storage namespace (FLOCI_OCI_DEFAULT_NAMESPACE), which GetNamespace
// then returns.
func WithNamespace(namespace string) Option {
	return func(c *builder) { c.withEnv(namespaceEnv, namespace) }
}

// WithPrivateKeyPEM makes GetPrivateKeyPEM return the given API signing key instead of one
// generated for the container, for tests that already have a key on disk or in a fixture. The
// emulator never verifies signatures, so any RSA key works.
func WithPrivateKeyPEM(privateKeyPEM string) Option {
	return func(c *builder) { c.privateKeyPEM = privateKeyPEM }
}

// WithDedicatedNetwork creates a dedicated Docker network for container-backed services
// (Functions, OKE) to talk to Floci; Terminate removes it.
func WithDedicatedNetwork() Option {
	return func(c *builder) { c.dedicatedNetwork = true }
}

// WithDockerSocket overrides whether the host Docker socket is mounted. By default it is mounted
// only while an enabled, non-mocked service spawns sibling containers (Functions, OKE); pass
// false on hosts where the socket cannot be mounted, true to always mount it.
func WithDockerSocket(enabled bool) Option {
	return func(c *builder) { c.dockerSocketOverride = &enabled }
}

// WithIdentityConfig applies Identity (IAM) service configuration.
func WithIdentityConfig(cfg IdentityConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithObjectStorageConfig applies Object Storage service configuration.
func WithObjectStorageConfig(cfg ObjectStorageConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithQueueConfig applies Queue service configuration.
func WithQueueConfig(cfg QueueConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithKmsConfig applies Key Management service configuration.
func WithKmsConfig(cfg KmsConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithVaultConfig applies Vault (secrets) service configuration.
func WithVaultConfig(cfg VaultConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithStreamingConfig applies Streaming service configuration.
func WithStreamingConfig(cfg StreamingConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithFunctionsConfig applies Functions service configuration.
func WithFunctionsConfig(cfg FunctionsConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithOkeConfig applies Kubernetes Engine (OKE) service configuration.
func WithOkeConfig(cfg OkeConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}
