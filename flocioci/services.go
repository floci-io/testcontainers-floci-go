package flocioci

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
	c.withEnv(ociDescriptor.ServiceEnv(token, "ENABLED"), boolStr(enabled))
	for _, s := range settings {
		key := ociDescriptor.ServiceEnv(token, s.name)
		if enabled && s.value != "" {
			c.withEnv(key, s.value)
		} else {
			delete(c.envVars, key)
		}
	}
}

// service is one Floci OCI service: its token in FLOCI_OCI_SERVICES_<token>_ENABLED and whether it
// has a mock switch (FLOCI_OCI_SERVICES_<token>_MOCK, false by default). The list mirrors ServicesConfig
// in the emulator's EmulatorConfig.
type service struct {
	token    string
	mockable bool
}

var services = []service{
	{token: "IDENTITY"},
	{token: "OBJECTSTORAGE"},
	{token: "QUEUE"},
	{token: "KMS"},
	{token: "VAULT"},
	{token: "STREAMING"},
	{token: "FUNCTIONS", mockable: true},
	{token: "OKE", mockable: true},
}

// IdentityConfig configures the Identity (IAM) service.
type IdentityConfig struct {
	Enabled bool
}

func DefaultIdentityConfig() IdentityConfig { return IdentityConfig{Enabled: true} }

func (c IdentityConfig) applyEnvVars(t *builder) {
	t.applyService("IDENTITY", c.Enabled)
}

// ObjectStorageConfig configures the Object Storage service.
type ObjectStorageConfig struct {
	Enabled bool
}

func DefaultObjectStorageConfig() ObjectStorageConfig { return ObjectStorageConfig{Enabled: true} }

func (c ObjectStorageConfig) applyEnvVars(t *builder) {
	t.applyService("OBJECTSTORAGE", c.Enabled)
}

// QueueConfig configures the Queue service.
type QueueConfig struct {
	Enabled bool
}

func DefaultQueueConfig() QueueConfig { return QueueConfig{Enabled: true} }

func (c QueueConfig) applyEnvVars(t *builder) {
	t.applyService("QUEUE", c.Enabled)
}

// KmsConfig configures the Key Management service.
type KmsConfig struct {
	Enabled bool
}

func DefaultKmsConfig() KmsConfig { return KmsConfig{Enabled: true} }

func (c KmsConfig) applyEnvVars(t *builder) {
	t.applyService("KMS", c.Enabled)
}

// VaultConfig configures the Vault (secrets) service.
type VaultConfig struct {
	Enabled bool
}

func DefaultVaultConfig() VaultConfig { return VaultConfig{Enabled: true} }

func (c VaultConfig) applyEnvVars(t *builder) {
	t.applyService("VAULT", c.Enabled)
}

// StreamingConfig configures the Streaming service.
type StreamingConfig struct {
	Enabled bool
}

func DefaultStreamingConfig() StreamingConfig { return StreamingConfig{Enabled: true} }

func (c StreamingConfig) applyEnvVars(t *builder) {
	t.applyService("STREAMING", c.Enabled)
}

// FunctionsConfig configures the Functions service. With Mock true it spawns no sibling containers.
type FunctionsConfig struct {
	Enabled bool
	Mock    bool
}

func DefaultFunctionsConfig() FunctionsConfig { return FunctionsConfig{Enabled: true, Mock: false} }

func (c FunctionsConfig) applyEnvVars(t *builder) {
	t.applyService("FUNCTIONS", c.Enabled, setting{"MOCK", boolStr(c.Mock)})
}

// OkeConfig configures the Kubernetes Engine (OKE) service. With Mock true it spawns no sibling containers.
type OkeConfig struct {
	Enabled bool
	Mock    bool
}

func DefaultOkeConfig() OkeConfig { return OkeConfig{Enabled: true, Mock: false} }

func (c OkeConfig) applyEnvVars(t *builder) {
	t.applyService("OKE", c.Enabled, setting{"MOCK", boolStr(c.Mock)})
}

// applyAllConfigs applies the default config of every service.
func (c *builder) applyAllConfigs() {
	DefaultIdentityConfig().applyEnvVars(c)
	DefaultObjectStorageConfig().applyEnvVars(c)
	DefaultQueueConfig().applyEnvVars(c)
	DefaultKmsConfig().applyEnvVars(c)
	DefaultVaultConfig().applyEnvVars(c)
	DefaultStreamingConfig().applyEnvVars(c)
	DefaultFunctionsConfig().applyEnvVars(c)
	DefaultOkeConfig().applyEnvVars(c)
}
