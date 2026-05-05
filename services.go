package floci

import "strconv"

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func intStr(i int) string { return strconv.Itoa(i) }

// AcmConfig configures the ACM (AWS Certificate Manager) service.
type AcmConfig struct {
	Enabled               bool
	ValidationWaitSeconds int
}

func DefaultAcmConfig() AcmConfig {
	return AcmConfig{Enabled: true, ValidationWaitSeconds: 0}
}

func (c AcmConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_ACM_ENABLED", boolStr(c.Enabled))
	t.withEnv("FLOCI_SERVICES_ACM_VALIDATION_WAIT_SECONDS", intStr(c.ValidationWaitSeconds))
}

// ApiGatewayConfig configures the API Gateway service.
type ApiGatewayConfig struct {
	Enabled bool
}

func DefaultApiGatewayConfig() ApiGatewayConfig { return ApiGatewayConfig{Enabled: true} }

func (c ApiGatewayConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_APIGATEWAY_ENABLED", boolStr(c.Enabled))
}

// ApiGatewayV2Config configures the API Gateway V2 service.
type ApiGatewayV2Config struct {
	Enabled bool
}

func DefaultApiGatewayV2Config() ApiGatewayV2Config { return ApiGatewayV2Config{Enabled: true} }

func (c ApiGatewayV2Config) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_APIGATEWAYV2_ENABLED", boolStr(c.Enabled))
}

// AppConfigConfig configures the AppConfig service.
type AppConfigConfig struct {
	Enabled bool
}

func DefaultAppConfigConfig() AppConfigConfig { return AppConfigConfig{Enabled: true} }

func (c AppConfigConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_APPCONFIG_ENABLED", boolStr(c.Enabled))
}

// AppConfigDataConfig configures the AppConfig Data service.
type AppConfigDataConfig struct {
	Enabled bool
}

func DefaultAppConfigDataConfig() AppConfigDataConfig { return AppConfigDataConfig{Enabled: true} }

func (c AppConfigDataConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_APPCONFIGDATA_ENABLED", boolStr(c.Enabled))
}

// AthenaConfig configures the Athena service.
type AthenaConfig struct {
	Enabled      bool
	Mock         bool
	DefaultImage string
}

func DefaultAthenaConfig() AthenaConfig {
	return AthenaConfig{Enabled: true, Mock: false, DefaultImage: "floci/floci-duck:latest"}
}

func (c AthenaConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_ATHENA_ENABLED", boolStr(c.Enabled))
	t.withEnv("FLOCI_SERVICES_ATHENA_MOCK", boolStr(c.Mock))
	t.withEnv("FLOCI_SERVICES_ATHENA_DEFAULT_IMAGE", c.DefaultImage)
}

// BedrockRuntimeConfig configures the Bedrock Runtime service.
type BedrockRuntimeConfig struct {
	Enabled bool
}

func DefaultBedrockRuntimeConfig() BedrockRuntimeConfig {
	return BedrockRuntimeConfig{Enabled: true}
}

func (c BedrockRuntimeConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_BEDROCK_RUNTIME_ENABLED", boolStr(c.Enabled))
}

// CloudFormationConfig configures the CloudFormation service.
type CloudFormationConfig struct {
	Enabled bool
}

func DefaultCloudFormationConfig() CloudFormationConfig { return CloudFormationConfig{Enabled: true} }

func (c CloudFormationConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_CLOUDFORMATION_ENABLED", boolStr(c.Enabled))
}

// CloudWatchLogsConfig configures the CloudWatch Logs service.
type CloudWatchLogsConfig struct {
	Enabled           bool
	MaxEventsPerQuery int
}

func DefaultCloudWatchLogsConfig() CloudWatchLogsConfig {
	return CloudWatchLogsConfig{Enabled: true, MaxEventsPerQuery: 10000}
}

func (c CloudWatchLogsConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_CLOUDWATCH_LOGS_ENABLED", boolStr(c.Enabled))
	t.withEnv("FLOCI_SERVICES_CLOUDWATCH_LOGS_MAX_EVENTS_PER_QUERY", intStr(c.MaxEventsPerQuery))
}

// CloudWatchMetricsConfig configures the CloudWatch Metrics service.
type CloudWatchMetricsConfig struct {
	Enabled bool
}

func DefaultCloudWatchMetricsConfig() CloudWatchMetricsConfig {
	return CloudWatchMetricsConfig{Enabled: true}
}

func (c CloudWatchMetricsConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_CLOUDWATCH_METRICS_ENABLED", boolStr(c.Enabled))
}

// CodeBuildConfig configures the CodeBuild service.
type CodeBuildConfig struct {
	Enabled bool
}

func DefaultCodeBuildConfig() CodeBuildConfig { return CodeBuildConfig{Enabled: true} }

func (c CodeBuildConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_CODEBUILD_ENABLED", boolStr(c.Enabled))
}

// CodeDeployConfig configures the CodeDeploy service.
type CodeDeployConfig struct {
	Enabled bool
}

func DefaultCodeDeployConfig() CodeDeployConfig { return CodeDeployConfig{Enabled: true} }

func (c CodeDeployConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_CODEDEPLOY_ENABLED", boolStr(c.Enabled))
}

// CognitoConfig configures the Cognito service.
type CognitoConfig struct {
	Enabled bool
}

func DefaultCognitoConfig() CognitoConfig { return CognitoConfig{Enabled: true} }

func (c CognitoConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_COGNITO_ENABLED", boolStr(c.Enabled))
}

// DynamoDbConfig configures the DynamoDB service.
type DynamoDbConfig struct {
	Enabled bool
}

func DefaultDynamoDbConfig() DynamoDbConfig { return DynamoDbConfig{Enabled: true} }

func (c DynamoDbConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_DYNAMODB_ENABLED", boolStr(c.Enabled))
}

// Ec2Config configures the EC2 service.
type Ec2Config struct {
	Enabled  bool
	Mock     bool
	ImdsPort int
}

func DefaultEc2Config() Ec2Config {
	return Ec2Config{Enabled: true, Mock: false, ImdsPort: 9169}
}

func (c Ec2Config) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_EC2_ENABLED", boolStr(c.Enabled))
	t.withEnv("FLOCI_SERVICES_EC2_MOCK", boolStr(c.Mock))
	t.withEnv("FLOCI_SERVICES_EC2_IMDS_PORT", intStr(c.ImdsPort))
}

// EcrConfig configures the ECR service.
// When enabled, ports [RegistryBasePort, RegistryBasePort+RegistryPortCount) are exposed.
type EcrConfig struct {
	Enabled           bool
	RegistryImage     string
	RegistryBasePort  int
	RegistryPortCount int
}

func DefaultEcrConfig() EcrConfig {
	return EcrConfig{
		Enabled:           true,
		RegistryImage:     "registry:2",
		RegistryBasePort:  5100,
		RegistryPortCount: 100,
	}
}

func (c EcrConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_ECR_ENABLED", boolStr(c.Enabled))
	t.withEnv("FLOCI_SERVICES_ECR_REGISTRY_IMAGE", c.RegistryImage)
	t.withEnv("FLOCI_SERVICES_ECR_REGISTRY_BASE_PORT", intStr(c.RegistryBasePort))
}

func (c EcrConfig) applyExposedPorts(t *FlociContainer) {
	for i := range c.RegistryPortCount {
		t.withPort(c.RegistryBasePort + i)
	}
}

// EcsConfig configures the ECS service.
type EcsConfig struct {
	Enabled bool
	Mock    bool
}

func DefaultEcsConfig() EcsConfig { return EcsConfig{Enabled: true, Mock: false} }

func (c EcsConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_ECS_ENABLED", boolStr(c.Enabled))
	t.withEnv("FLOCI_SERVICES_ECS_MOCK", boolStr(c.Mock))
}

// EksConfig configures the EKS service.
// When enabled, ports [ApiServerBasePort, ApiServerBasePort+ApiServerPortCount) are exposed.
type EksConfig struct {
	Enabled            bool
	Mock               bool
	Provider           string
	DefaultImage       string
	ApiServerBasePort  int
	ApiServerPortCount int
}

func DefaultEksConfig() EksConfig {
	return EksConfig{
		Enabled:            true,
		Mock:               false,
		Provider:           "k3s",
		DefaultImage:       "rancher/k3s:latest",
		ApiServerBasePort:  6500,
		ApiServerPortCount: 100,
	}
}

func (c EksConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_EKS_ENABLED", boolStr(c.Enabled))
	t.withEnv("FLOCI_SERVICES_EKS_MOCK", boolStr(c.Mock))
	t.withEnv("FLOCI_SERVICES_EKS_PROVIDER", c.Provider)
	t.withEnv("FLOCI_SERVICES_EKS_DEFAULT_IMAGE", c.DefaultImage)
	t.withEnv("FLOCI_SERVICES_EKS_API_SERVER_BASE_PORT", intStr(c.ApiServerBasePort))
}

func (c EksConfig) applyExposedPorts(t *FlociContainer) {
	for i := range c.ApiServerPortCount {
		t.withPort(c.ApiServerBasePort + i)
	}
}

// ElastiCacheConfig configures the ElastiCache service.
// Ports [ProxyBasePort, ProxyBasePort+ProxyPortCount) are exposed.
type ElastiCacheConfig struct {
	Enabled        bool
	DefaultImage   string
	ProxyBasePort  int
	ProxyPortCount int
}

func DefaultElastiCacheConfig() ElastiCacheConfig {
	return ElastiCacheConfig{
		Enabled:        true,
		DefaultImage:   "valkey/valkey:8",
		ProxyBasePort:  6379,
		ProxyPortCount: 21,
	}
}

func (c ElastiCacheConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_ELASTICACHE_ENABLED", boolStr(c.Enabled))
	t.withEnv("FLOCI_SERVICES_ELASTICACHE_DEFAULT_IMAGE", c.DefaultImage)
	t.withEnv("FLOCI_SERVICES_ELASTICACHE_PROXY_BASE_PORT", intStr(c.ProxyBasePort))
}

func (c ElastiCacheConfig) applyExposedPorts(t *FlociContainer) {
	for i := range c.ProxyPortCount {
		t.withPort(c.ProxyBasePort + i)
	}
}

// ElbV2Config configures the ELBv2 service.
type ElbV2Config struct {
	Enabled bool
	Mock    bool
}

func DefaultElbV2Config() ElbV2Config { return ElbV2Config{Enabled: true, Mock: false} }

func (c ElbV2Config) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_ELBV2_ENABLED", boolStr(c.Enabled))
	t.withEnv("FLOCI_SERVICES_ELBV2_MOCK", boolStr(c.Mock))
}

// EventBridgeConfig configures the EventBridge service.
type EventBridgeConfig struct {
	Enabled bool
}

func DefaultEventBridgeConfig() EventBridgeConfig { return EventBridgeConfig{Enabled: true} }

func (c EventBridgeConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_EVENTBRIDGE_ENABLED", boolStr(c.Enabled))
}

// FirehoseConfig configures the Kinesis Data Firehose service.
type FirehoseConfig struct {
	Enabled bool
}

func DefaultFirehoseConfig() FirehoseConfig { return FirehoseConfig{Enabled: true} }

func (c FirehoseConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_FIREHOSE_ENABLED", boolStr(c.Enabled))
}

// GlueConfig configures the Glue service.
type GlueConfig struct {
	Enabled bool
}

func DefaultGlueConfig() GlueConfig { return GlueConfig{Enabled: true} }

func (c GlueConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_GLUE_ENABLED", boolStr(c.Enabled))
}

// IamConfig configures the IAM service.
type IamConfig struct {
	Enabled            bool
	EnforcementEnabled bool
}

func DefaultIamConfig() IamConfig {
	return IamConfig{Enabled: true, EnforcementEnabled: false}
}

func (c IamConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_IAM_ENABLED", boolStr(c.Enabled))
	t.withEnv("FLOCI_SERVICES_IAM_ENFORCEMENT_ENABLED", boolStr(c.EnforcementEnabled))
}

// KinesisConfig configures the Kinesis service.
type KinesisConfig struct {
	Enabled bool
}

func DefaultKinesisConfig() KinesisConfig { return KinesisConfig{Enabled: true} }

func (c KinesisConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_KINESIS_ENABLED", boolStr(c.Enabled))
}

// KmsConfig configures the KMS service.
type KmsConfig struct {
	Enabled bool
}

func DefaultKmsConfig() KmsConfig { return KmsConfig{Enabled: true} }

func (c KmsConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_KMS_ENABLED", boolStr(c.Enabled))
}

// LambdaConfig configures the Lambda service.
// Set ExposeRuntimePorts=true and WithDedicatedNetwork() to invoke Lambdas from the host.
type LambdaConfig struct {
	Enabled               bool
	DefaultMemoryMb       int
	DefaultTimeoutSeconds int
	Ephemeral             bool
	HotReloadEnabled      bool
	RuntimeApiBasePort    int
	RuntimeApiPortCount   int
	ExposeRuntimePorts    bool
	DockerNetwork         string
}

func DefaultLambdaConfig() LambdaConfig {
	return LambdaConfig{
		Enabled:               true,
		DefaultMemoryMb:       128,
		DefaultTimeoutSeconds: 3,
		Ephemeral:             false,
		HotReloadEnabled:      false,
		RuntimeApiBasePort:    9200,
		RuntimeApiPortCount:   100,
		ExposeRuntimePorts:    false,
	}
}

func (c LambdaConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_LAMBDA_ENABLED", boolStr(c.Enabled))
	t.withEnv("FLOCI_SERVICES_LAMBDA_DEFAULT_MEMORY_MB", intStr(c.DefaultMemoryMb))
	t.withEnv("FLOCI_SERVICES_LAMBDA_DEFAULT_TIMEOUT_SECONDS", intStr(c.DefaultTimeoutSeconds))
	t.withEnv("FLOCI_SERVICES_LAMBDA_EPHEMERAL", boolStr(c.Ephemeral))
	t.withEnv("FLOCI_SERVICES_LAMBDA_HOT_RELOAD_ENABLED", boolStr(c.HotReloadEnabled))
	t.withEnv("FLOCI_SERVICES_LAMBDA_RUNTIME_API_BASE_PORT", intStr(c.RuntimeApiBasePort))
	if c.DockerNetwork != "" {
		t.withEnv("FLOCI_SERVICES_LAMBDA_DOCKER_NETWORK", c.DockerNetwork)
	}
}

func (c LambdaConfig) applyExposedPorts(t *FlociContainer) {
	if c.ExposeRuntimePorts {
		for i := range c.RuntimeApiPortCount {
			t.withPort(c.RuntimeApiBasePort + i)
		}
	}
}

// MskConfig configures the MSK (Managed Streaming for Kafka) service.
type MskConfig struct {
	Enabled      bool
	Mock         bool
	DefaultImage string
}

func DefaultMskConfig() MskConfig {
	return MskConfig{Enabled: true, Mock: false, DefaultImage: "redpandadata/redpanda:latest"}
}

func (c MskConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_MSK_ENABLED", boolStr(c.Enabled))
	t.withEnv("FLOCI_SERVICES_MSK_MOCK", boolStr(c.Mock))
	t.withEnv("FLOCI_SERVICES_MSK_DEFAULT_IMAGE", c.DefaultImage)
}

// OpenSearchConfig configures the OpenSearch service.
// Ports [ProxyBasePort, ProxyBasePort+ProxyPortCount) are exposed.
type OpenSearchConfig struct {
	Enabled        bool
	Mock           bool
	DefaultImage   string
	ProxyBasePort  int
	ProxyPortCount int
}

func DefaultOpenSearchConfig() OpenSearchConfig {
	return OpenSearchConfig{
		Enabled:        true,
		Mock:           false,
		DefaultImage:   "opensearchproject/opensearch:2",
		ProxyBasePort:  9400,
		ProxyPortCount: 100,
	}
}

func (c OpenSearchConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_OPENSEARCH_ENABLED", boolStr(c.Enabled))
	t.withEnv("FLOCI_SERVICES_OPENSEARCH_MOCK", boolStr(c.Mock))
	t.withEnv("FLOCI_SERVICES_OPENSEARCH_DEFAULT_IMAGE", c.DefaultImage)
	t.withEnv("FLOCI_SERVICES_OPENSEARCH_PROXY_BASE_PORT", intStr(c.ProxyBasePort))
}

func (c OpenSearchConfig) applyExposedPorts(t *FlociContainer) {
	for i := range c.ProxyPortCount {
		t.withPort(c.ProxyBasePort + i)
	}
}

// PipesConfig configures the Pipes service.
type PipesConfig struct {
	Enabled bool
}

func DefaultPipesConfig() PipesConfig { return PipesConfig{Enabled: true} }

func (c PipesConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_PIPES_ENABLED", boolStr(c.Enabled))
}

// RdsConfig configures the RDS service.
// Ports [ProxyBasePort, ProxyBasePort+ProxyPortCount) are exposed for direct DB access.
type RdsConfig struct {
	Enabled              bool
	ProxyBasePort        int
	ProxyPortCount       int
	DefaultPostgresImage string
	DefaultMysqlImage    string
	DefaultMariadbImage  string
}

func DefaultRdsConfig() RdsConfig {
	return RdsConfig{
		Enabled:              true,
		ProxyBasePort:        7001,
		ProxyPortCount:       99,
		DefaultPostgresImage: "postgres:16-alpine",
		DefaultMysqlImage:    "mysql:8.0",
		DefaultMariadbImage:  "mariadb:11",
	}
}

func (c RdsConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_RDS_ENABLED", boolStr(c.Enabled))
	t.withEnv("FLOCI_SERVICES_RDS_PROXY_BASE_PORT", intStr(c.ProxyBasePort))
	t.withEnv("FLOCI_SERVICES_RDS_DEFAULT_POSTGRES_IMAGE", c.DefaultPostgresImage)
	t.withEnv("FLOCI_SERVICES_RDS_DEFAULT_MYSQL_IMAGE", c.DefaultMysqlImage)
	t.withEnv("FLOCI_SERVICES_RDS_DEFAULT_MARIADB_IMAGE", c.DefaultMariadbImage)
}

func (c RdsConfig) applyExposedPorts(t *FlociContainer) {
	for i := range c.ProxyPortCount {
		t.withPort(c.ProxyBasePort + i)
	}
}

// ResourceGroupsTaggingConfig configures the Resource Groups Tagging service.
type ResourceGroupsTaggingConfig struct {
	Enabled bool
}

func DefaultResourceGroupsTaggingConfig() ResourceGroupsTaggingConfig {
	return ResourceGroupsTaggingConfig{Enabled: true}
}

func (c ResourceGroupsTaggingConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_RESOURCEGROUPSTAGGING_ENABLED", boolStr(c.Enabled))
}

// S3Config configures the S3 service.
type S3Config struct {
	Enabled                     bool
	DefaultPresignExpirySeconds int
}

func DefaultS3Config() S3Config {
	return S3Config{Enabled: true, DefaultPresignExpirySeconds: 3600}
}

func (c S3Config) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_S3_ENABLED", boolStr(c.Enabled))
	t.withEnv("FLOCI_SERVICES_S3_DEFAULT_PRESIGN_EXPIRY_SECONDS", intStr(c.DefaultPresignExpirySeconds))
}

// SchedulerConfig configures the Scheduler service.
type SchedulerConfig struct {
	Enabled bool
}

func DefaultSchedulerConfig() SchedulerConfig { return SchedulerConfig{Enabled: true} }

func (c SchedulerConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_SCHEDULER_ENABLED", boolStr(c.Enabled))
}

// SecretsManagerConfig configures the Secrets Manager service.
type SecretsManagerConfig struct {
	Enabled                   bool
	DefaultRecoveryWindowDays int
}

func DefaultSecretsManagerConfig() SecretsManagerConfig {
	return SecretsManagerConfig{Enabled: true, DefaultRecoveryWindowDays: 30}
}

func (c SecretsManagerConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_SECRETS_MANAGER_ENABLED", boolStr(c.Enabled))
	t.withEnv("FLOCI_SERVICES_SECRETS_MANAGER_DEFAULT_RECOVERY_WINDOW_DAYS", intStr(c.DefaultRecoveryWindowDays))
}

// SesConfig configures the SES service.
type SesConfig struct {
	Enabled bool
}

func DefaultSesConfig() SesConfig { return SesConfig{Enabled: true} }

func (c SesConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_SES_ENABLED", boolStr(c.Enabled))
}

// SesV2Config configures the SES V2 service.
type SesV2Config struct {
	Enabled bool
}

func DefaultSesV2Config() SesV2Config { return SesV2Config{Enabled: true} }

func (c SesV2Config) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_SES_V2_ENABLED", boolStr(c.Enabled))
}

// SnsConfig configures the SNS service.
type SnsConfig struct {
	Enabled bool
}

func DefaultSnsConfig() SnsConfig { return SnsConfig{Enabled: true} }

func (c SnsConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_SNS_ENABLED", boolStr(c.Enabled))
}

// SqsConfig configures the SQS service.
type SqsConfig struct {
	Enabled                  bool
	DefaultVisibilityTimeout int
	MaxMessageSize           int
}

func DefaultSqsConfig() SqsConfig {
	return SqsConfig{Enabled: true, DefaultVisibilityTimeout: 30, MaxMessageSize: 262144}
}

func (c SqsConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_SQS_ENABLED", boolStr(c.Enabled))
	t.withEnv("FLOCI_SERVICES_SQS_DEFAULT_VISIBILITY_TIMEOUT", intStr(c.DefaultVisibilityTimeout))
	t.withEnv("FLOCI_SERVICES_SQS_MAX_MESSAGE_SIZE", intStr(c.MaxMessageSize))
}

// SsmConfig configures the SSM (Systems Manager / Parameter Store) service.
type SsmConfig struct {
	Enabled             bool
	MaxParameterHistory int
}

func DefaultSsmConfig() SsmConfig {
	return SsmConfig{Enabled: true, MaxParameterHistory: 5}
}

func (c SsmConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_SSM_ENABLED", boolStr(c.Enabled))
	t.withEnv("FLOCI_SERVICES_SSM_MAX_PARAMETER_HISTORY", intStr(c.MaxParameterHistory))
}

// StepFunctionsConfig configures the Step Functions service.
type StepFunctionsConfig struct {
	Enabled bool
}

func DefaultStepFunctionsConfig() StepFunctionsConfig { return StepFunctionsConfig{Enabled: true} }

func (c StepFunctionsConfig) applyEnvVars(t *FlociContainer) {
	t.withEnv("FLOCI_SERVICES_STEPFUNCTIONS_ENABLED", boolStr(c.Enabled))
}
