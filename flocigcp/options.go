package flocigcp

// Package-level options for Run.

// WithProjectID sets the emulator's default project (FLOCI_GCP_DEFAULT_PROJECT_ID), the one
// GetProjectID and GOOGLE_CLOUD_PROJECT in GetEmulatorEnv then name. The emulator accepts any
// project ID in a request, whatever the default.
func WithProjectID(projectID string) Option {
	return func(c *builder) { c.withEnv(projectIDEnv, projectID) }
}

// WithDedicatedNetwork creates a dedicated Docker network for container-backed services
// (Cloud SQL, Cloud Run, Kafka, GKE, BigQuery) to talk to Floci; Terminate removes it.
func WithDedicatedNetwork() Option {
	return func(c *builder) { c.dedicatedNetwork = true }
}

// WithDockerSocket overrides whether the host Docker socket is mounted. By default it is mounted
// only while an enabled, non-mocked service spawns sibling containers (BigQuery, Cloud Run,
// Cloud SQL, GKE, Kafka); pass false on hosts where the socket cannot be mounted, true to always
// mount it.
func WithDockerSocket(enabled bool) Option {
	return func(c *builder) { c.dockerSocketOverride = &enabled }
}

// WithComputeConfig applies Compute Engine service configuration.
func WithComputeConfig(cfg ComputeConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithGcsConfig applies Cloud Storage service configuration.
func WithGcsConfig(cfg GcsConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithPubSubConfig applies Pub/Sub service configuration.
func WithPubSubConfig(cfg PubSubConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithFirestoreConfig applies Firestore service configuration.
func WithFirestoreConfig(cfg FirestoreConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithDatastoreConfig applies Datastore service configuration.
func WithDatastoreConfig(cfg DatastoreConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithIamConfig applies IAM service configuration.
func WithIamConfig(cfg IamConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithIamCredentialsConfig applies IAM Service Account Credentials service configuration.
func WithIamCredentialsConfig(cfg IamCredentialsConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithStsConfig applies Security Token Service service configuration.
func WithStsConfig(cfg StsConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithSecretManagerConfig applies Secret Manager service configuration.
func WithSecretManagerConfig(cfg SecretManagerConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithLoggingConfig applies Cloud Logging service configuration.
func WithLoggingConfig(cfg LoggingConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithKmsConfig applies Cloud KMS service configuration.
func WithKmsConfig(cfg KmsConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithKafkaConfig applies Managed Service for Apache Kafka service configuration.
func WithKafkaConfig(cfg KafkaConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithCloudSqlConfig applies Cloud SQL service configuration.
func WithCloudSqlConfig(cfg CloudSqlConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithCloudTasksConfig applies Cloud Tasks service configuration.
func WithCloudTasksConfig(cfg CloudTasksConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithCloudRunConfig applies Cloud Run service configuration.
func WithCloudRunConfig(cfg CloudRunConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithCloudFunctionsConfig applies Cloud Functions service configuration.
func WithCloudFunctionsConfig(cfg CloudFunctionsConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithMonitoringConfig applies Cloud Monitoring service configuration.
func WithMonitoringConfig(cfg MonitoringConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithSchedulerConfig applies Cloud Scheduler service configuration.
func WithSchedulerConfig(cfg SchedulerConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithEventarcConfig applies Eventarc service configuration.
func WithEventarcConfig(cfg EventarcConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithGkeConfig applies Kubernetes Engine (GKE) service configuration.
func WithGkeConfig(cfg GkeConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithServiceUsageConfig applies Service Usage service configuration.
func WithServiceUsageConfig(cfg ServiceUsageConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithResourceManagerConfig applies Resource Manager service configuration.
func WithResourceManagerConfig(cfg ResourceManagerConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithFirebaseAuthConfig applies Firebase Auth service configuration.
func WithFirebaseAuthConfig(cfg FirebaseAuthConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}

// WithBigQueryConfig applies BigQuery service configuration.
func WithBigQueryConfig(cfg BigQueryConfig) Option {
	return func(c *builder) { cfg.applyEnvVars(c) }
}
