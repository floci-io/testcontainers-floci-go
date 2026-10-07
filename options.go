package floci

// Package-level options for Run. Each one mirrors the FlociContainer builder method of the
// same name, so a test written against the builder ports by replacing c.WithX(v) with
// floci.WithX(v) in the Run call.

// WithRegion sets the default AWS region.
func WithRegion(region string) Option {
	return func(c *FlociContainer) { c.WithRegion(region) }
}

// WithAccountID sets the default AWS account ID.
func WithAccountID(accountID string) Option {
	return func(c *FlociContainer) { c.WithAccountID(accountID) }
}

// WithAvailabilityZone sets the default availability zone.
func WithAvailabilityZone(zone string) Option {
	return func(c *FlociContainer) { c.WithAvailabilityZone(zone) }
}

// WithDedicatedNetwork creates a dedicated Docker network for container-based services
// (Lambda, RDS, ElastiCache, ...) to talk to Floci; Terminate removes it.
func WithDedicatedNetwork() Option {
	return func(c *FlociContainer) { c.WithDedicatedNetwork() }
}

// WithDockerSocket overrides Docker socket auto-detection; see FlociContainer.WithDockerSocket.
func WithDockerSocket(enabled bool) Option {
	return func(c *FlociContainer) { c.WithDockerSocket(enabled) }
}

// WithAcmConfig applies ACM service configuration.
func WithAcmConfig(cfg AcmConfig) Option {
	return func(c *FlociContainer) { c.WithAcmConfig(cfg) }
}

// WithApiGatewayConfig applies API Gateway service configuration.
func WithApiGatewayConfig(cfg ApiGatewayConfig) Option {
	return func(c *FlociContainer) { c.WithApiGatewayConfig(cfg) }
}

// WithApiGatewayV2Config applies API Gateway V2 service configuration.
func WithApiGatewayV2Config(cfg ApiGatewayV2Config) Option {
	return func(c *FlociContainer) { c.WithApiGatewayV2Config(cfg) }
}

// WithAppConfigConfig applies AppConfig service configuration.
func WithAppConfigConfig(cfg AppConfigConfig) Option {
	return func(c *FlociContainer) { c.WithAppConfigConfig(cfg) }
}

// WithAppConfigDataConfig applies AppConfig Data service configuration.
func WithAppConfigDataConfig(cfg AppConfigDataConfig) Option {
	return func(c *FlociContainer) { c.WithAppConfigDataConfig(cfg) }
}

// WithAthenaConfig applies Athena service configuration.
func WithAthenaConfig(cfg AthenaConfig) Option {
	return func(c *FlociContainer) { c.WithAthenaConfig(cfg) }
}

// WithBedrockRuntimeConfig applies Bedrock Runtime service configuration.
func WithBedrockRuntimeConfig(cfg BedrockRuntimeConfig) Option {
	return func(c *FlociContainer) { c.WithBedrockRuntimeConfig(cfg) }
}

// WithCloudFormationConfig applies CloudFormation service configuration.
func WithCloudFormationConfig(cfg CloudFormationConfig) Option {
	return func(c *FlociContainer) { c.WithCloudFormationConfig(cfg) }
}

// WithCloudWatchLogsConfig applies CloudWatch Logs service configuration.
func WithCloudWatchLogsConfig(cfg CloudWatchLogsConfig) Option {
	return func(c *FlociContainer) { c.WithCloudWatchLogsConfig(cfg) }
}

// WithCloudWatchMetricsConfig applies CloudWatch Metrics service configuration.
func WithCloudWatchMetricsConfig(cfg CloudWatchMetricsConfig) Option {
	return func(c *FlociContainer) { c.WithCloudWatchMetricsConfig(cfg) }
}

// WithCodeBuildConfig applies CodeBuild service configuration.
func WithCodeBuildConfig(cfg CodeBuildConfig) Option {
	return func(c *FlociContainer) { c.WithCodeBuildConfig(cfg) }
}

// WithCodeDeployConfig applies CodeDeploy service configuration.
func WithCodeDeployConfig(cfg CodeDeployConfig) Option {
	return func(c *FlociContainer) { c.WithCodeDeployConfig(cfg) }
}

// WithCognitoConfig applies Cognito service configuration.
func WithCognitoConfig(cfg CognitoConfig) Option {
	return func(c *FlociContainer) { c.WithCognitoConfig(cfg) }
}

// WithDynamoDbConfig applies DynamoDB service configuration.
func WithDynamoDbConfig(cfg DynamoDbConfig) Option {
	return func(c *FlociContainer) { c.WithDynamoDbConfig(cfg) }
}

// WithEc2Config applies EC2 service configuration.
func WithEc2Config(cfg Ec2Config) Option {
	return func(c *FlociContainer) { c.WithEc2Config(cfg) }
}

// WithEcrConfig applies ECR service configuration.
func WithEcrConfig(cfg EcrConfig) Option {
	return func(c *FlociContainer) { c.WithEcrConfig(cfg) }
}

// WithEcsConfig applies ECS service configuration.
func WithEcsConfig(cfg EcsConfig) Option {
	return func(c *FlociContainer) { c.WithEcsConfig(cfg) }
}

// WithEksConfig applies EKS service configuration.
func WithEksConfig(cfg EksConfig) Option {
	return func(c *FlociContainer) { c.WithEksConfig(cfg) }
}

// WithElastiCacheConfig applies ElastiCache service configuration.
func WithElastiCacheConfig(cfg ElastiCacheConfig) Option {
	return func(c *FlociContainer) { c.WithElastiCacheConfig(cfg) }
}

// WithElbV2Config applies ELBv2 service configuration.
func WithElbV2Config(cfg ElbV2Config) Option {
	return func(c *FlociContainer) { c.WithElbV2Config(cfg) }
}

// WithEventBridgeConfig applies EventBridge service configuration.
func WithEventBridgeConfig(cfg EventBridgeConfig) Option {
	return func(c *FlociContainer) { c.WithEventBridgeConfig(cfg) }
}

// WithFirehoseConfig applies Firehose service configuration.
func WithFirehoseConfig(cfg FirehoseConfig) Option {
	return func(c *FlociContainer) { c.WithFirehoseConfig(cfg) }
}

// WithGlueConfig applies Glue service configuration.
func WithGlueConfig(cfg GlueConfig) Option {
	return func(c *FlociContainer) { c.WithGlueConfig(cfg) }
}

// WithIamConfig applies IAM service configuration.
func WithIamConfig(cfg IamConfig) Option {
	return func(c *FlociContainer) { c.WithIamConfig(cfg) }
}

// WithKinesisConfig applies Kinesis service configuration.
func WithKinesisConfig(cfg KinesisConfig) Option {
	return func(c *FlociContainer) { c.WithKinesisConfig(cfg) }
}

// WithKmsConfig applies KMS service configuration.
func WithKmsConfig(cfg KmsConfig) Option {
	return func(c *FlociContainer) { c.WithKmsConfig(cfg) }
}

// WithLambdaConfig applies Lambda service configuration.
func WithLambdaConfig(cfg LambdaConfig) Option {
	return func(c *FlociContainer) { c.WithLambdaConfig(cfg) }
}

// WithMskConfig applies MSK service configuration.
func WithMskConfig(cfg MskConfig) Option {
	return func(c *FlociContainer) { c.WithMskConfig(cfg) }
}

// WithOpenSearchConfig applies OpenSearch service configuration.
func WithOpenSearchConfig(cfg OpenSearchConfig) Option {
	return func(c *FlociContainer) { c.WithOpenSearchConfig(cfg) }
}

// WithPipesConfig applies Pipes service configuration.
func WithPipesConfig(cfg PipesConfig) Option {
	return func(c *FlociContainer) { c.WithPipesConfig(cfg) }
}

// WithRdsConfig applies RDS service configuration.
func WithRdsConfig(cfg RdsConfig) Option {
	return func(c *FlociContainer) { c.WithRdsConfig(cfg) }
}

// WithResourceGroupsTaggingConfig applies Resource Groups Tagging service configuration.
func WithResourceGroupsTaggingConfig(cfg ResourceGroupsTaggingConfig) Option {
	return func(c *FlociContainer) { c.WithResourceGroupsTaggingConfig(cfg) }
}

// WithS3Config applies S3 service configuration.
func WithS3Config(cfg S3Config) Option {
	return func(c *FlociContainer) { c.WithS3Config(cfg) }
}

// WithSchedulerConfig applies Scheduler service configuration.
func WithSchedulerConfig(cfg SchedulerConfig) Option {
	return func(c *FlociContainer) { c.WithSchedulerConfig(cfg) }
}

// WithSecretsManagerConfig applies Secrets Manager service configuration.
func WithSecretsManagerConfig(cfg SecretsManagerConfig) Option {
	return func(c *FlociContainer) { c.WithSecretsManagerConfig(cfg) }
}

// WithSesConfig applies SES service configuration.
func WithSesConfig(cfg SesConfig) Option {
	return func(c *FlociContainer) { c.WithSesConfig(cfg) }
}

// WithSesV2Config applies SES V2 service configuration.
func WithSesV2Config(cfg SesV2Config) Option {
	return func(c *FlociContainer) { c.WithSesV2Config(cfg) }
}

// WithSnsConfig applies SNS service configuration.
func WithSnsConfig(cfg SnsConfig) Option {
	return func(c *FlociContainer) { c.WithSnsConfig(cfg) }
}

// WithSqsConfig applies SQS service configuration.
func WithSqsConfig(cfg SqsConfig) Option {
	return func(c *FlociContainer) { c.WithSqsConfig(cfg) }
}

// WithSsmConfig applies SSM service configuration.
func WithSsmConfig(cfg SsmConfig) Option {
	return func(c *FlociContainer) { c.WithSsmConfig(cfg) }
}

// WithStepFunctionsConfig applies Step Functions service configuration.
func WithStepFunctionsConfig(cfg StepFunctionsConfig) Option {
	return func(c *FlociContainer) { c.WithStepFunctionsConfig(cfg) }
}
