// Package floci is the original import path of the Floci AWS module, kept so existing code
// keeps compiling. Every name here is an alias of the same name in package flociaws, the
// module's home for AWS alongside the other Floci clouds:
//
//	import "github.com/floci-io/testcontainers-floci-go/flociaws"
//
//	fc, err := flociaws.Run(ctx, "floci/floci:latest", flociaws.WithRegion("eu-west-1"))
//
// Deprecated: use package github.com/floci-io/testcontainers-floci-go/flociaws.
package floci

import (
	"context"

	"github.com/testcontainers/testcontainers-go"

	"github.com/floci-io/testcontainers-floci-go/flociaws"
)

// Identity defaults of the AWS emulator.
//
// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
const (
	DefaultRegion           = flociaws.DefaultRegion
	DefaultAvailabilityZone = flociaws.DefaultAvailabilityZone
	DefaultAccountID        = flociaws.DefaultAccountID
	DefaultAccessKey        = flociaws.DefaultAccessKey
	DefaultSecretKey        = flociaws.DefaultSecretKey
)

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type AcmConfig = flociaws.AcmConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type ApiGatewayConfig = flociaws.ApiGatewayConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type ApiGatewayV2Config = flociaws.ApiGatewayV2Config

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type AppConfigConfig = flociaws.AppConfigConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type AppConfigDataConfig = flociaws.AppConfigDataConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type AthenaConfig = flociaws.AthenaConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type BedrockRuntimeConfig = flociaws.BedrockRuntimeConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type CloudFormationConfig = flociaws.CloudFormationConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type CloudWatchLogsConfig = flociaws.CloudWatchLogsConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type CloudWatchMetricsConfig = flociaws.CloudWatchMetricsConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type CodeBuildConfig = flociaws.CodeBuildConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type CodeDeployConfig = flociaws.CodeDeployConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type CognitoConfig = flociaws.CognitoConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type Container = flociaws.Container

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type DynamoDbConfig = flociaws.DynamoDbConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type Ec2Config = flociaws.Ec2Config

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type EcrConfig = flociaws.EcrConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type EcsConfig = flociaws.EcsConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type EksConfig = flociaws.EksConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type ElastiCacheConfig = flociaws.ElastiCacheConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type ElbV2Config = flociaws.ElbV2Config

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type EventBridgeConfig = flociaws.EventBridgeConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type FirehoseConfig = flociaws.FirehoseConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type FlociContainer = flociaws.FlociContainer

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type GlueConfig = flociaws.GlueConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type IamConfig = flociaws.IamConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type KinesisConfig = flociaws.KinesisConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type KmsConfig = flociaws.KmsConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type LambdaConfig = flociaws.LambdaConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type MskConfig = flociaws.MskConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type OpenSearchConfig = flociaws.OpenSearchConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type Option = flociaws.Option

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type PipesConfig = flociaws.PipesConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type RdsConfig = flociaws.RdsConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type ResourceGroupsTaggingConfig = flociaws.ResourceGroupsTaggingConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type S3Config = flociaws.S3Config

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type SchedulerConfig = flociaws.SchedulerConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type SecretsManagerConfig = flociaws.SecretsManagerConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type SesConfig = flociaws.SesConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type SesV2Config = flociaws.SesV2Config

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type SnsConfig = flociaws.SnsConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type SqsConfig = flociaws.SqsConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type SsmConfig = flociaws.SsmConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type StartedFlociContainer = flociaws.StartedFlociContainer

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
type StepFunctionsConfig = flociaws.StepFunctionsConfig

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultAcmConfig() AcmConfig { return flociaws.DefaultAcmConfig() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultApiGatewayConfig() ApiGatewayConfig { return flociaws.DefaultApiGatewayConfig() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultApiGatewayV2Config() ApiGatewayV2Config { return flociaws.DefaultApiGatewayV2Config() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultAppConfigConfig() AppConfigConfig { return flociaws.DefaultAppConfigConfig() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultAppConfigDataConfig() AppConfigDataConfig { return flociaws.DefaultAppConfigDataConfig() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultAthenaConfig() AthenaConfig { return flociaws.DefaultAthenaConfig() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultBedrockRuntimeConfig() BedrockRuntimeConfig {
	return flociaws.DefaultBedrockRuntimeConfig()
}

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultCloudFormationConfig() CloudFormationConfig {
	return flociaws.DefaultCloudFormationConfig()
}

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultCloudWatchLogsConfig() CloudWatchLogsConfig {
	return flociaws.DefaultCloudWatchLogsConfig()
}

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultCloudWatchMetricsConfig() CloudWatchMetricsConfig {
	return flociaws.DefaultCloudWatchMetricsConfig()
}

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultCodeBuildConfig() CodeBuildConfig { return flociaws.DefaultCodeBuildConfig() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultCodeDeployConfig() CodeDeployConfig { return flociaws.DefaultCodeDeployConfig() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultCognitoConfig() CognitoConfig { return flociaws.DefaultCognitoConfig() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	return flociaws.Run(ctx, img, opts...)
}

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultDynamoDbConfig() DynamoDbConfig { return flociaws.DefaultDynamoDbConfig() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultEc2Config() Ec2Config { return flociaws.DefaultEc2Config() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultEcrConfig() EcrConfig { return flociaws.DefaultEcrConfig() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultEcsConfig() EcsConfig { return flociaws.DefaultEcsConfig() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultEksConfig() EksConfig { return flociaws.DefaultEksConfig() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultElastiCacheConfig() ElastiCacheConfig { return flociaws.DefaultElastiCacheConfig() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultElbV2Config() ElbV2Config { return flociaws.DefaultElbV2Config() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultEventBridgeConfig() EventBridgeConfig { return flociaws.DefaultEventBridgeConfig() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultFirehoseConfig() FirehoseConfig { return flociaws.DefaultFirehoseConfig() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func NewFlociContainer() *FlociContainer { return flociaws.NewFlociContainer() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultGlueConfig() GlueConfig { return flociaws.DefaultGlueConfig() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultIamConfig() IamConfig { return flociaws.DefaultIamConfig() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultKinesisConfig() KinesisConfig { return flociaws.DefaultKinesisConfig() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultKmsConfig() KmsConfig { return flociaws.DefaultKmsConfig() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultLambdaConfig() LambdaConfig { return flociaws.DefaultLambdaConfig() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultMskConfig() MskConfig { return flociaws.DefaultMskConfig() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultOpenSearchConfig() OpenSearchConfig { return flociaws.DefaultOpenSearchConfig() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithAccountID(accountID string) Option { return flociaws.WithAccountID(accountID) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithAcmConfig(cfg AcmConfig) Option { return flociaws.WithAcmConfig(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithApiGatewayConfig(cfg ApiGatewayConfig) Option { return flociaws.WithApiGatewayConfig(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithApiGatewayV2Config(cfg ApiGatewayV2Config) Option {
	return flociaws.WithApiGatewayV2Config(cfg)
}

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithAppConfigConfig(cfg AppConfigConfig) Option { return flociaws.WithAppConfigConfig(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithAppConfigDataConfig(cfg AppConfigDataConfig) Option {
	return flociaws.WithAppConfigDataConfig(cfg)
}

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithAthenaConfig(cfg AthenaConfig) Option { return flociaws.WithAthenaConfig(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithAvailabilityZone(zone string) Option { return flociaws.WithAvailabilityZone(zone) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithBedrockRuntimeConfig(cfg BedrockRuntimeConfig) Option {
	return flociaws.WithBedrockRuntimeConfig(cfg)
}

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithCloudFormationConfig(cfg CloudFormationConfig) Option {
	return flociaws.WithCloudFormationConfig(cfg)
}

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithCloudWatchLogsConfig(cfg CloudWatchLogsConfig) Option {
	return flociaws.WithCloudWatchLogsConfig(cfg)
}

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithCloudWatchMetricsConfig(cfg CloudWatchMetricsConfig) Option {
	return flociaws.WithCloudWatchMetricsConfig(cfg)
}

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithCodeBuildConfig(cfg CodeBuildConfig) Option { return flociaws.WithCodeBuildConfig(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithCodeDeployConfig(cfg CodeDeployConfig) Option { return flociaws.WithCodeDeployConfig(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithCognitoConfig(cfg CognitoConfig) Option { return flociaws.WithCognitoConfig(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithDedicatedNetwork() Option { return flociaws.WithDedicatedNetwork() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithDockerSocket(enabled bool) Option { return flociaws.WithDockerSocket(enabled) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithDynamoDbConfig(cfg DynamoDbConfig) Option { return flociaws.WithDynamoDbConfig(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithEc2Config(cfg Ec2Config) Option { return flociaws.WithEc2Config(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithEcrConfig(cfg EcrConfig) Option { return flociaws.WithEcrConfig(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithEcsConfig(cfg EcsConfig) Option { return flociaws.WithEcsConfig(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithEksConfig(cfg EksConfig) Option { return flociaws.WithEksConfig(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithElastiCacheConfig(cfg ElastiCacheConfig) Option { return flociaws.WithElastiCacheConfig(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithElbV2Config(cfg ElbV2Config) Option { return flociaws.WithElbV2Config(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithEventBridgeConfig(cfg EventBridgeConfig) Option { return flociaws.WithEventBridgeConfig(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithFirehoseConfig(cfg FirehoseConfig) Option { return flociaws.WithFirehoseConfig(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithGlueConfig(cfg GlueConfig) Option { return flociaws.WithGlueConfig(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithIamConfig(cfg IamConfig) Option { return flociaws.WithIamConfig(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithKinesisConfig(cfg KinesisConfig) Option { return flociaws.WithKinesisConfig(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithKmsConfig(cfg KmsConfig) Option { return flociaws.WithKmsConfig(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithLambdaConfig(cfg LambdaConfig) Option { return flociaws.WithLambdaConfig(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithMskConfig(cfg MskConfig) Option { return flociaws.WithMskConfig(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithOpenSearchConfig(cfg OpenSearchConfig) Option { return flociaws.WithOpenSearchConfig(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithPipesConfig(cfg PipesConfig) Option { return flociaws.WithPipesConfig(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithRdsConfig(cfg RdsConfig) Option { return flociaws.WithRdsConfig(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithRegion(region string) Option { return flociaws.WithRegion(region) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithResourceGroupsTaggingConfig(cfg ResourceGroupsTaggingConfig) Option {
	return flociaws.WithResourceGroupsTaggingConfig(cfg)
}

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithS3Config(cfg S3Config) Option { return flociaws.WithS3Config(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithSchedulerConfig(cfg SchedulerConfig) Option { return flociaws.WithSchedulerConfig(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithSecretsManagerConfig(cfg SecretsManagerConfig) Option {
	return flociaws.WithSecretsManagerConfig(cfg)
}

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithSesConfig(cfg SesConfig) Option { return flociaws.WithSesConfig(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithSesV2Config(cfg SesV2Config) Option { return flociaws.WithSesV2Config(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithSnsConfig(cfg SnsConfig) Option { return flociaws.WithSnsConfig(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithSqsConfig(cfg SqsConfig) Option { return flociaws.WithSqsConfig(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithSsmConfig(cfg SsmConfig) Option { return flociaws.WithSsmConfig(cfg) }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func WithStepFunctionsConfig(cfg StepFunctionsConfig) Option {
	return flociaws.WithStepFunctionsConfig(cfg)
}

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultPipesConfig() PipesConfig { return flociaws.DefaultPipesConfig() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultRdsConfig() RdsConfig { return flociaws.DefaultRdsConfig() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultResourceGroupsTaggingConfig() ResourceGroupsTaggingConfig {
	return flociaws.DefaultResourceGroupsTaggingConfig()
}

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultS3Config() S3Config { return flociaws.DefaultS3Config() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultSchedulerConfig() SchedulerConfig { return flociaws.DefaultSchedulerConfig() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultSecretsManagerConfig() SecretsManagerConfig {
	return flociaws.DefaultSecretsManagerConfig()
}

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultSesConfig() SesConfig { return flociaws.DefaultSesConfig() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultSesV2Config() SesV2Config { return flociaws.DefaultSesV2Config() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultSnsConfig() SnsConfig { return flociaws.DefaultSnsConfig() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultSqsConfig() SqsConfig { return flociaws.DefaultSqsConfig() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultSsmConfig() SsmConfig { return flociaws.DefaultSsmConfig() }

// Deprecated: use the same name in package github.com/floci-io/testcontainers-floci-go/flociaws.
func DefaultStepFunctionsConfig() StepFunctionsConfig { return flociaws.DefaultStepFunctionsConfig() }
