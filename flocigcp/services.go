package flocigcp

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
	c.withEnv(gcpDescriptor.ServiceEnv(token, "ENABLED"), boolStr(enabled))
	for _, s := range settings {
		key := gcpDescriptor.ServiceEnv(token, s.name)
		if enabled && s.value != "" {
			c.withEnv(key, s.value)
		} else {
			delete(c.envVars, key)
		}
	}
}

// service is one Floci GCP service: its token in FLOCI_GCP_SERVICES_<token>_ENABLED and whether it
// has a mock switch (FLOCI_GCP_SERVICES_<token>_MOCK, false by default). The list mirrors ServicesConfig
// in the emulator's EmulatorConfig.
type service struct {
	token    string
	mockable bool
}

var services = []service{
	{token: "COMPUTE"},
	{token: "GCS"},
	{token: "PUBSUB"},
	{token: "FIRESTORE"},
	{token: "DATASTORE"},
	{token: "IAM"},
	{token: "IAMCREDENTIALS"},
	{token: "STS"},
	{token: "SECRETMANAGER"},
	{token: "LOGGING"},
	{token: "KMS"},
	{token: "KAFKA", mockable: true},
	{token: "CLOUDSQL", mockable: true},
	{token: "CLOUDTASKS"},
	{token: "CLOUDRUN", mockable: true},
	{token: "CLOUDFUNCTIONS"},
	{token: "MONITORING"},
	{token: "SCHEDULER"},
	{token: "EVENTARC"},
	{token: "GKE", mockable: true},
	{token: "SERVICEUSAGE"},
	{token: "RESOURCEMANAGER"},
	{token: "FIREBASEAUTH"},
	{token: "BIGQUERY", mockable: true},
}

// ComputeConfig configures the Compute Engine service.
type ComputeConfig struct {
	Enabled bool
}

func DefaultComputeConfig() ComputeConfig { return ComputeConfig{Enabled: true} }

func (c ComputeConfig) applyEnvVars(t *builder) {
	t.applyService("COMPUTE", c.Enabled)
}

// GcsConfig configures the Cloud Storage service.
type GcsConfig struct {
	Enabled bool
}

func DefaultGcsConfig() GcsConfig { return GcsConfig{Enabled: true} }

func (c GcsConfig) applyEnvVars(t *builder) {
	t.applyService("GCS", c.Enabled)
}

// PubSubConfig configures the Pub/Sub service.
type PubSubConfig struct {
	Enabled bool
}

func DefaultPubSubConfig() PubSubConfig { return PubSubConfig{Enabled: true} }

func (c PubSubConfig) applyEnvVars(t *builder) {
	t.applyService("PUBSUB", c.Enabled)
}

// FirestoreConfig configures the Firestore service.
type FirestoreConfig struct {
	Enabled bool
}

func DefaultFirestoreConfig() FirestoreConfig { return FirestoreConfig{Enabled: true} }

func (c FirestoreConfig) applyEnvVars(t *builder) {
	t.applyService("FIRESTORE", c.Enabled)
}

// DatastoreConfig configures the Datastore service.
type DatastoreConfig struct {
	Enabled bool
}

func DefaultDatastoreConfig() DatastoreConfig { return DatastoreConfig{Enabled: true} }

func (c DatastoreConfig) applyEnvVars(t *builder) {
	t.applyService("DATASTORE", c.Enabled)
}

// IamConfig configures the IAM service.
type IamConfig struct {
	Enabled bool
}

func DefaultIamConfig() IamConfig { return IamConfig{Enabled: true} }

func (c IamConfig) applyEnvVars(t *builder) {
	t.applyService("IAM", c.Enabled)
}

// IamCredentialsConfig configures the IAM Service Account Credentials service.
type IamCredentialsConfig struct {
	Enabled bool
}

func DefaultIamCredentialsConfig() IamCredentialsConfig { return IamCredentialsConfig{Enabled: true} }

func (c IamCredentialsConfig) applyEnvVars(t *builder) {
	t.applyService("IAMCREDENTIALS", c.Enabled)
}

// StsConfig configures the Security Token Service service.
type StsConfig struct {
	Enabled bool
}

func DefaultStsConfig() StsConfig { return StsConfig{Enabled: true} }

func (c StsConfig) applyEnvVars(t *builder) {
	t.applyService("STS", c.Enabled)
}

// SecretManagerConfig configures the Secret Manager service.
type SecretManagerConfig struct {
	Enabled bool
}

func DefaultSecretManagerConfig() SecretManagerConfig { return SecretManagerConfig{Enabled: true} }

func (c SecretManagerConfig) applyEnvVars(t *builder) {
	t.applyService("SECRETMANAGER", c.Enabled)
}

// LoggingConfig configures the Cloud Logging service.
type LoggingConfig struct {
	Enabled bool
}

func DefaultLoggingConfig() LoggingConfig { return LoggingConfig{Enabled: true} }

func (c LoggingConfig) applyEnvVars(t *builder) {
	t.applyService("LOGGING", c.Enabled)
}

// KmsConfig configures the Cloud KMS service.
type KmsConfig struct {
	Enabled bool
}

func DefaultKmsConfig() KmsConfig { return KmsConfig{Enabled: true} }

func (c KmsConfig) applyEnvVars(t *builder) {
	t.applyService("KMS", c.Enabled)
}

// KafkaConfig configures the Managed Service for Apache Kafka service. With Mock true it spawns no sibling containers.
type KafkaConfig struct {
	Enabled bool
	Mock    bool
}

func DefaultKafkaConfig() KafkaConfig { return KafkaConfig{Enabled: true, Mock: false} }

func (c KafkaConfig) applyEnvVars(t *builder) {
	t.applyService("KAFKA", c.Enabled, setting{"MOCK", boolStr(c.Mock)})
}

// CloudSqlConfig configures the Cloud SQL service. With Mock true it spawns no sibling containers.
type CloudSqlConfig struct {
	Enabled bool
	Mock    bool
}

func DefaultCloudSqlConfig() CloudSqlConfig { return CloudSqlConfig{Enabled: true, Mock: false} }

func (c CloudSqlConfig) applyEnvVars(t *builder) {
	t.applyService("CLOUDSQL", c.Enabled, setting{"MOCK", boolStr(c.Mock)})
}

// CloudTasksConfig configures the Cloud Tasks service.
type CloudTasksConfig struct {
	Enabled bool
}

func DefaultCloudTasksConfig() CloudTasksConfig { return CloudTasksConfig{Enabled: true} }

func (c CloudTasksConfig) applyEnvVars(t *builder) {
	t.applyService("CLOUDTASKS", c.Enabled)
}

// CloudRunConfig configures the Cloud Run service. With Mock true it spawns no sibling containers.
type CloudRunConfig struct {
	Enabled bool
	Mock    bool
}

func DefaultCloudRunConfig() CloudRunConfig { return CloudRunConfig{Enabled: true, Mock: false} }

func (c CloudRunConfig) applyEnvVars(t *builder) {
	t.applyService("CLOUDRUN", c.Enabled, setting{"MOCK", boolStr(c.Mock)})
}

// CloudFunctionsConfig configures the Cloud Functions service.
type CloudFunctionsConfig struct {
	Enabled bool
}

func DefaultCloudFunctionsConfig() CloudFunctionsConfig { return CloudFunctionsConfig{Enabled: true} }

func (c CloudFunctionsConfig) applyEnvVars(t *builder) {
	t.applyService("CLOUDFUNCTIONS", c.Enabled)
}

// MonitoringConfig configures the Cloud Monitoring service.
type MonitoringConfig struct {
	Enabled bool
}

func DefaultMonitoringConfig() MonitoringConfig { return MonitoringConfig{Enabled: true} }

func (c MonitoringConfig) applyEnvVars(t *builder) {
	t.applyService("MONITORING", c.Enabled)
}

// SchedulerConfig configures the Cloud Scheduler service.
type SchedulerConfig struct {
	Enabled bool
}

func DefaultSchedulerConfig() SchedulerConfig { return SchedulerConfig{Enabled: true} }

func (c SchedulerConfig) applyEnvVars(t *builder) {
	t.applyService("SCHEDULER", c.Enabled)
}

// EventarcConfig configures the Eventarc service.
type EventarcConfig struct {
	Enabled bool
}

func DefaultEventarcConfig() EventarcConfig { return EventarcConfig{Enabled: true} }

func (c EventarcConfig) applyEnvVars(t *builder) {
	t.applyService("EVENTARC", c.Enabled)
}

// GkeConfig configures the Kubernetes Engine (GKE) service. With Mock true it spawns no sibling containers.
type GkeConfig struct {
	Enabled bool
	Mock    bool
}

func DefaultGkeConfig() GkeConfig { return GkeConfig{Enabled: true, Mock: false} }

func (c GkeConfig) applyEnvVars(t *builder) {
	t.applyService("GKE", c.Enabled, setting{"MOCK", boolStr(c.Mock)})
}

// ServiceUsageConfig configures the Service Usage service.
type ServiceUsageConfig struct {
	Enabled bool
}

func DefaultServiceUsageConfig() ServiceUsageConfig { return ServiceUsageConfig{Enabled: true} }

func (c ServiceUsageConfig) applyEnvVars(t *builder) {
	t.applyService("SERVICEUSAGE", c.Enabled)
}

// ResourceManagerConfig configures the Resource Manager service.
type ResourceManagerConfig struct {
	Enabled bool
}

func DefaultResourceManagerConfig() ResourceManagerConfig {
	return ResourceManagerConfig{Enabled: true}
}

func (c ResourceManagerConfig) applyEnvVars(t *builder) {
	t.applyService("RESOURCEMANAGER", c.Enabled)
}

// FirebaseAuthConfig configures the Firebase Auth service.
type FirebaseAuthConfig struct {
	Enabled bool
}

func DefaultFirebaseAuthConfig() FirebaseAuthConfig { return FirebaseAuthConfig{Enabled: true} }

func (c FirebaseAuthConfig) applyEnvVars(t *builder) {
	t.applyService("FIREBASEAUTH", c.Enabled)
}

// BigQueryConfig configures the BigQuery service. Queries run in a floci-duck sibling container,
// unless Mock is true (a built-in SQL subset) or DuckURL names a floci-duck that is already running.
type BigQueryConfig struct {
	Enabled bool
	Mock    bool
	DuckURL string
}

func DefaultBigQueryConfig() BigQueryConfig { return BigQueryConfig{Enabled: true, Mock: false} }

func (c BigQueryConfig) applyEnvVars(t *builder) {
	t.applyService("BIGQUERY", c.Enabled, setting{"MOCK", boolStr(c.Mock)}, setting{"DUCK_URL", c.DuckURL})
}

// bigQueryNeedsDocker reports whether BigQuery starts its own floci-duck container, which it does
// unless one is configured.
func bigQueryNeedsDocker(setting func(string) string) bool { return setting("DUCK_URL") == "" }

// applyAllConfigs applies the default config of every service.
func (c *builder) applyAllConfigs() {
	DefaultComputeConfig().applyEnvVars(c)
	DefaultGcsConfig().applyEnvVars(c)
	DefaultPubSubConfig().applyEnvVars(c)
	DefaultFirestoreConfig().applyEnvVars(c)
	DefaultDatastoreConfig().applyEnvVars(c)
	DefaultIamConfig().applyEnvVars(c)
	DefaultIamCredentialsConfig().applyEnvVars(c)
	DefaultStsConfig().applyEnvVars(c)
	DefaultSecretManagerConfig().applyEnvVars(c)
	DefaultLoggingConfig().applyEnvVars(c)
	DefaultKmsConfig().applyEnvVars(c)
	DefaultKafkaConfig().applyEnvVars(c)
	DefaultCloudSqlConfig().applyEnvVars(c)
	DefaultCloudTasksConfig().applyEnvVars(c)
	DefaultCloudRunConfig().applyEnvVars(c)
	DefaultCloudFunctionsConfig().applyEnvVars(c)
	DefaultMonitoringConfig().applyEnvVars(c)
	DefaultSchedulerConfig().applyEnvVars(c)
	DefaultEventarcConfig().applyEnvVars(c)
	DefaultGkeConfig().applyEnvVars(c)
	DefaultServiceUsageConfig().applyEnvVars(c)
	DefaultResourceManagerConfig().applyEnvVars(c)
	DefaultFirebaseAuthConfig().applyEnvVars(c)
	DefaultBigQueryConfig().applyEnvVars(c)
}
