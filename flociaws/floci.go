// Package flociaws provides the Testcontainers module for the Floci local AWS emulator (floci/floci).
//
// Example:
//
//	fc, err := floci.Run(ctx, "floci/floci:latest")
//	testcontainers.CleanupContainer(t, fc)
//	if err != nil { ... }
//
//	cfg, _ := config.LoadDefaultConfig(ctx,
//	    config.WithRegion(fc.GetRegion()),
//	    config.WithBaseEndpoint(fc.GetEndpoint()),
//	    config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
//	        fc.GetAccessKey(), fc.GetSecretKey(), "",
//	    )),
//	)
package flociaws

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strconv"
	"time"

	"github.com/testcontainers/testcontainers-go"

	"github.com/floci-io/testcontainers-floci-go/internal/core"
)

const (
	defaultImage = "floci/floci:latest"
	flociPort    = 4566

	DefaultRegion           = "us-east-1"
	DefaultAvailabilityZone = "us-east-1a"
	DefaultAccountID        = "000000000000"
	DefaultAccessKey        = "test"
	DefaultSecretKey        = "test"
)

// awsDescriptor is the AWS emulator (floci/floci) as data, for the shared core.
var awsDescriptor = core.Descriptor{
	Name:           "aws",
	DefaultImage:   defaultImage,
	Port:           flociPort,
	EnvPrefix:      "FLOCI_",
	HealthPath:     "/_floci/health",
	ResetPath:      "/_floci/state/reset",
	LogLevelEnv:    "QUARKUS_LOG_CATEGORY__IO_GITHUB_HECTORVENT__LEVEL",
	StartupTimeout: 120 * time.Second,
	// RDS clients connect to the endpoint the API returns, so it must be reachable from the host.
	HostSettings: []core.HostSetting{{Token: "RDS", Setting: "ENDPOINT_HOST"}},
	// Services that spawn sibling containers, mirroring requiresDockerSocket() in the Java module.
	SocketServices: []core.SocketService{
		{Token: "ATHENA", Mockable: true},
		{Token: "CODEBUILD"},
		{Token: "EC2", Mockable: true},
		{Token: "ECR"},
		{Token: "ECS", Mockable: true},
		{Token: "EKS", Mockable: true},
		{Token: "ELASTICACHE"},
		{Token: "LAMBDA"},
		{Token: "MSK", Mockable: true},
		{Token: "NEPTUNE"},
		{Token: "OPENSEARCH", Mockable: true},
		{Token: "RDS"},
	},
}

// FlociContainer is a builder for a Floci testcontainer.
type FlociContainer struct {
	image            string
	envVars          map[string]string
	ports            map[int]struct{}
	dedicatedNetwork bool
	// dockerSocket overrides socket auto-detection when non-nil (see WithDockerSocket).
	dockerSocketOverride *bool
	// generatedNamespace is the namespace newBuilder set; each start replaces it with a fresh
	// one unless the caller chose their own, so two containers from one builder never share it.
	generatedNamespace string

	acmConfig                   AcmConfig
	apiGatewayConfig            ApiGatewayConfig
	apiGatewayV2Config          ApiGatewayV2Config
	appConfigConfig             AppConfigConfig
	appConfigDataConfig         AppConfigDataConfig
	athenaConfig                AthenaConfig
	bedrockRuntimeConfig        BedrockRuntimeConfig
	cloudFormationConfig        CloudFormationConfig
	cloudWatchLogsConfig        CloudWatchLogsConfig
	cloudWatchMetricsConfig     CloudWatchMetricsConfig
	codeBuildConfig             CodeBuildConfig
	codeDeployConfig            CodeDeployConfig
	cognitoConfig               CognitoConfig
	dynamoDbConfig              DynamoDbConfig
	ec2Config                   Ec2Config
	ecrConfig                   EcrConfig
	ecsConfig                   EcsConfig
	eksConfig                   EksConfig
	elastiCacheConfig           ElastiCacheConfig
	elbV2Config                 ElbV2Config
	eventBridgeConfig           EventBridgeConfig
	firehoseConfig              FirehoseConfig
	glueConfig                  GlueConfig
	iamConfig                   IamConfig
	kinesisConfig               KinesisConfig
	kmsConfig                   KmsConfig
	lambdaConfig                LambdaConfig
	mskConfig                   MskConfig
	openSearchConfig            OpenSearchConfig
	pipesConfig                 PipesConfig
	rdsConfig                   RdsConfig
	resourceGroupsTaggingConfig ResourceGroupsTaggingConfig
	s3Config                    S3Config
	schedulerConfig             SchedulerConfig
	secretsManagerConfig        SecretsManagerConfig
	sesConfig                   SesConfig
	sesV2Config                 SesV2Config
	snsConfig                   SnsConfig
	sqsConfig                   SqsConfig
	ssmConfig                   SsmConfig
	stepFunctionsConfig         StepFunctionsConfig
}

// NewFlociContainer creates a new FlociContainer builder with default configuration.
//
// Deprecated: use Run, which follows the testcontainers-go module convention:
// floci.Run(ctx, "floci/floci:latest", floci.WithRegion("eu-west-1")).
func NewFlociContainer() *FlociContainer {
	return newBuilder()
}

func newBuilder() *FlociContainer {
	c := &FlociContainer{
		image:   defaultImage,
		envVars: make(map[string]string),
		ports:   map[int]struct{}{flociPort: {}},

		acmConfig:                   DefaultAcmConfig(),
		apiGatewayConfig:            DefaultApiGatewayConfig(),
		apiGatewayV2Config:          DefaultApiGatewayV2Config(),
		appConfigConfig:             DefaultAppConfigConfig(),
		appConfigDataConfig:         DefaultAppConfigDataConfig(),
		athenaConfig:                DefaultAthenaConfig(),
		bedrockRuntimeConfig:        DefaultBedrockRuntimeConfig(),
		cloudFormationConfig:        DefaultCloudFormationConfig(),
		cloudWatchLogsConfig:        DefaultCloudWatchLogsConfig(),
		cloudWatchMetricsConfig:     DefaultCloudWatchMetricsConfig(),
		codeBuildConfig:             DefaultCodeBuildConfig(),
		codeDeployConfig:            DefaultCodeDeployConfig(),
		cognitoConfig:               DefaultCognitoConfig(),
		dynamoDbConfig:              DefaultDynamoDbConfig(),
		ec2Config:                   DefaultEc2Config(),
		ecrConfig:                   DefaultEcrConfig(),
		ecsConfig:                   DefaultEcsConfig(),
		eksConfig:                   DefaultEksConfig(),
		elastiCacheConfig:           DefaultElastiCacheConfig(),
		elbV2Config:                 DefaultElbV2Config(),
		eventBridgeConfig:           DefaultEventBridgeConfig(),
		firehoseConfig:              DefaultFirehoseConfig(),
		glueConfig:                  DefaultGlueConfig(),
		iamConfig:                   DefaultIamConfig(),
		kinesisConfig:               DefaultKinesisConfig(),
		kmsConfig:                   DefaultKmsConfig(),
		lambdaConfig:                DefaultLambdaConfig(),
		mskConfig:                   DefaultMskConfig(),
		openSearchConfig:            DefaultOpenSearchConfig(),
		pipesConfig:                 DefaultPipesConfig(),
		rdsConfig:                   DefaultRdsConfig(),
		resourceGroupsTaggingConfig: DefaultResourceGroupsTaggingConfig(),
		s3Config:                    DefaultS3Config(),
		schedulerConfig:             DefaultSchedulerConfig(),
		secretsManagerConfig:        DefaultSecretsManagerConfig(),
		sesConfig:                   DefaultSesConfig(),
		sesV2Config:                 DefaultSesV2Config(),
		snsConfig:                   DefaultSnsConfig(),
		sqsConfig:                   DefaultSqsConfig(),
		ssmConfig:                   DefaultSsmConfig(),
		stepFunctionsConfig:         DefaultStepFunctionsConfig(),
	}

	c.withEnv("FLOCI_DEFAULT_REGION", DefaultRegion)
	c.withEnv("FLOCI_DEFAULT_ACCOUNT_ID", DefaultAccountID)
	c.withEnv("FLOCI_DEFAULT_AVAILABILITY_ZONE", DefaultAvailabilityZone)
	// Sibling containers are named after the resource; a unique namespace per container
	// keeps parallel test runs from colliding and makes leftovers attributable.
	c.generatedNamespace = core.NewNamespace()
	c.withEnv(awsDescriptor.ResourceNamespaceEnv(), c.generatedNamespace)
	c.applyAllConfigs()
	return c
}

// Option configures Floci-specific settings. It implements
// testcontainers.ContainerCustomizer, so Floci options and generic testcontainers
// options can be passed to Run together:
//
//	floci.Run(ctx, "floci/floci:latest",
//	    floci.WithRegion("eu-west-1"),
//	    testcontainers.WithEnv(map[string]string{"FLOCI_LOG_LEVEL": "debug"}),
//	)
type Option func(*FlociContainer)

// Customize is a no-op: Floci options are applied to the module's own settings
// before the container request is built.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error { return nil }

// Run creates and starts a Floci container from the given image, following the
// testcontainers-go module convention. Floci options (WithRegion, WithS3Config, ...)
// configure the emulator; any other testcontainers.ContainerCustomizer is applied to
// the container request after the module's defaults, so it can override them.
//
// The returned container is non-nil whenever the container was created, even if
// Run also returns an error, so it can always be cleaned up:
//
//	fc, err := floci.Run(ctx, "floci/floci:latest")
//	testcontainers.CleanupContainer(t, fc)
//	require.NoError(t, err)
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	c := newBuilder()
	if img != "" {
		c.image = img
	}
	var customizers []testcontainers.ContainerCustomizer
	for _, opt := range opts {
		if o, ok := opt.(Option); ok {
			o(c)
			continue
		}
		customizers = append(customizers, opt)
	}
	return c.run(ctx, customizers...)
}

func (c *FlociContainer) withEnv(key, value string) *FlociContainer {
	c.envVars[key] = value
	return c
}

func (c *FlociContainer) withPort(port int) *FlociContainer {
	c.ports[port] = struct{}{}
	return c
}

// WithImage overrides the Docker image used for the container.
func (c *FlociContainer) WithImage(image string) *FlociContainer {
	c.image = image
	return c
}

// WithRegion sets the default AWS region.
func (c *FlociContainer) WithRegion(region string) *FlociContainer {
	return c.withEnv("FLOCI_DEFAULT_REGION", region)
}

// WithAccountID sets the default AWS account ID.
func (c *FlociContainer) WithAccountID(accountID string) *FlociContainer {
	return c.withEnv("FLOCI_DEFAULT_ACCOUNT_ID", accountID)
}

// WithAvailabilityZone sets the default availability zone.
func (c *FlociContainer) WithAvailabilityZone(zone string) *FlociContainer {
	return c.withEnv("FLOCI_DEFAULT_AVAILABILITY_ZONE", zone)
}

// WithDedicatedNetwork creates a dedicated Docker network for container-based services
// (Lambda, RDS, ElastiCache, etc.) to communicate with Floci. The network name is
// generated at Start() time and automatically passed via FLOCI_SERVICES_DOCKER_NETWORK.
func (c *FlociContainer) WithDedicatedNetwork() *FlociContainer {
	c.dedicatedNetwork = true
	return c
}

// WithDockerSocket overrides whether the host Docker socket is mounted. By default it
// is mounted only while an enabled service spawns sibling containers (Lambda, RDS,
// ElastiCache, ECS, EC2, EKS, ECR, MSK, OpenSearch, Athena, CodeBuild); pass false on
// hosts where the socket cannot be mounted, true to always mount it.
func (c *FlociContainer) WithDockerSocket(enabled bool) *FlociContainer {
	c.dockerSocketOverride = &enabled
	return c
}

// WithAcmConfig applies ACM service configuration.
func (c *FlociContainer) WithAcmConfig(cfg AcmConfig) *FlociContainer {
	c.acmConfig = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithApiGatewayConfig applies API Gateway service configuration.
func (c *FlociContainer) WithApiGatewayConfig(cfg ApiGatewayConfig) *FlociContainer {
	c.apiGatewayConfig = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithApiGatewayV2Config applies API Gateway V2 service configuration.
func (c *FlociContainer) WithApiGatewayV2Config(cfg ApiGatewayV2Config) *FlociContainer {
	c.apiGatewayV2Config = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithAppConfigConfig applies AppConfig service configuration.
func (c *FlociContainer) WithAppConfigConfig(cfg AppConfigConfig) *FlociContainer {
	c.appConfigConfig = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithAppConfigDataConfig applies AppConfig Data service configuration.
func (c *FlociContainer) WithAppConfigDataConfig(cfg AppConfigDataConfig) *FlociContainer {
	c.appConfigDataConfig = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithAthenaConfig applies Athena service configuration.
func (c *FlociContainer) WithAthenaConfig(cfg AthenaConfig) *FlociContainer {
	c.athenaConfig = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithBedrockRuntimeConfig applies Bedrock Runtime service configuration.
func (c *FlociContainer) WithBedrockRuntimeConfig(cfg BedrockRuntimeConfig) *FlociContainer {
	c.bedrockRuntimeConfig = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithCloudFormationConfig applies CloudFormation service configuration.
func (c *FlociContainer) WithCloudFormationConfig(cfg CloudFormationConfig) *FlociContainer {
	c.cloudFormationConfig = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithCloudWatchLogsConfig applies CloudWatch Logs service configuration.
func (c *FlociContainer) WithCloudWatchLogsConfig(cfg CloudWatchLogsConfig) *FlociContainer {
	c.cloudWatchLogsConfig = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithCloudWatchMetricsConfig applies CloudWatch Metrics service configuration.
func (c *FlociContainer) WithCloudWatchMetricsConfig(cfg CloudWatchMetricsConfig) *FlociContainer {
	c.cloudWatchMetricsConfig = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithCodeBuildConfig applies CodeBuild service configuration.
func (c *FlociContainer) WithCodeBuildConfig(cfg CodeBuildConfig) *FlociContainer {
	c.codeBuildConfig = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithCodeDeployConfig applies CodeDeploy service configuration.
func (c *FlociContainer) WithCodeDeployConfig(cfg CodeDeployConfig) *FlociContainer {
	c.codeDeployConfig = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithCognitoConfig applies Cognito service configuration.
func (c *FlociContainer) WithCognitoConfig(cfg CognitoConfig) *FlociContainer {
	c.cognitoConfig = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithDynamoDbConfig applies DynamoDB service configuration.
func (c *FlociContainer) WithDynamoDbConfig(cfg DynamoDbConfig) *FlociContainer {
	c.dynamoDbConfig = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithEc2Config applies EC2 service configuration.
func (c *FlociContainer) WithEc2Config(cfg Ec2Config) *FlociContainer {
	c.ec2Config = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithEcrConfig applies ECR service configuration.
func (c *FlociContainer) WithEcrConfig(cfg EcrConfig) *FlociContainer {
	c.ecrConfig = cfg
	cfg.applyEnvVars(c)
	c.refreshExposedPorts()
	return c
}

// WithEcsConfig applies ECS service configuration.
func (c *FlociContainer) WithEcsConfig(cfg EcsConfig) *FlociContainer {
	c.ecsConfig = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithEksConfig applies EKS service configuration.
func (c *FlociContainer) WithEksConfig(cfg EksConfig) *FlociContainer {
	c.eksConfig = cfg
	cfg.applyEnvVars(c)
	c.refreshExposedPorts()
	return c
}

// WithElastiCacheConfig applies ElastiCache service configuration.
func (c *FlociContainer) WithElastiCacheConfig(cfg ElastiCacheConfig) *FlociContainer {
	c.elastiCacheConfig = cfg
	cfg.applyEnvVars(c)
	c.refreshExposedPorts()
	return c
}

// WithElbV2Config applies ELBv2 service configuration.
func (c *FlociContainer) WithElbV2Config(cfg ElbV2Config) *FlociContainer {
	c.elbV2Config = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithEventBridgeConfig applies EventBridge service configuration.
func (c *FlociContainer) WithEventBridgeConfig(cfg EventBridgeConfig) *FlociContainer {
	c.eventBridgeConfig = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithFirehoseConfig applies Firehose service configuration.
func (c *FlociContainer) WithFirehoseConfig(cfg FirehoseConfig) *FlociContainer {
	c.firehoseConfig = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithGlueConfig applies Glue service configuration.
func (c *FlociContainer) WithGlueConfig(cfg GlueConfig) *FlociContainer {
	c.glueConfig = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithIamConfig applies IAM service configuration.
func (c *FlociContainer) WithIamConfig(cfg IamConfig) *FlociContainer {
	c.iamConfig = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithKinesisConfig applies Kinesis service configuration.
func (c *FlociContainer) WithKinesisConfig(cfg KinesisConfig) *FlociContainer {
	c.kinesisConfig = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithKmsConfig applies KMS service configuration.
func (c *FlociContainer) WithKmsConfig(cfg KmsConfig) *FlociContainer {
	c.kmsConfig = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithLambdaConfig applies Lambda service configuration.
func (c *FlociContainer) WithLambdaConfig(cfg LambdaConfig) *FlociContainer {
	c.lambdaConfig = cfg
	cfg.applyEnvVars(c)
	c.refreshExposedPorts()
	return c
}

// WithMskConfig applies MSK service configuration.
func (c *FlociContainer) WithMskConfig(cfg MskConfig) *FlociContainer {
	c.mskConfig = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithOpenSearchConfig applies OpenSearch service configuration.
func (c *FlociContainer) WithOpenSearchConfig(cfg OpenSearchConfig) *FlociContainer {
	c.openSearchConfig = cfg
	cfg.applyEnvVars(c)
	c.refreshExposedPorts()
	return c
}

// WithPipesConfig applies Pipes service configuration.
func (c *FlociContainer) WithPipesConfig(cfg PipesConfig) *FlociContainer {
	c.pipesConfig = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithRdsConfig applies RDS service configuration.
func (c *FlociContainer) WithRdsConfig(cfg RdsConfig) *FlociContainer {
	c.rdsConfig = cfg
	cfg.applyEnvVars(c)
	c.refreshExposedPorts()
	return c
}

// WithResourceGroupsTaggingConfig applies Resource Groups Tagging service configuration.
func (c *FlociContainer) WithResourceGroupsTaggingConfig(cfg ResourceGroupsTaggingConfig) *FlociContainer {
	c.resourceGroupsTaggingConfig = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithS3Config applies S3 service configuration.
func (c *FlociContainer) WithS3Config(cfg S3Config) *FlociContainer {
	c.s3Config = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithSchedulerConfig applies Scheduler service configuration.
func (c *FlociContainer) WithSchedulerConfig(cfg SchedulerConfig) *FlociContainer {
	c.schedulerConfig = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithSecretsManagerConfig applies Secrets Manager service configuration.
func (c *FlociContainer) WithSecretsManagerConfig(cfg SecretsManagerConfig) *FlociContainer {
	c.secretsManagerConfig = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithSesConfig applies SES service configuration.
func (c *FlociContainer) WithSesConfig(cfg SesConfig) *FlociContainer {
	c.sesConfig = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithSesV2Config applies SES V2 service configuration.
func (c *FlociContainer) WithSesV2Config(cfg SesV2Config) *FlociContainer {
	c.sesV2Config = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithSnsConfig applies SNS service configuration.
func (c *FlociContainer) WithSnsConfig(cfg SnsConfig) *FlociContainer {
	c.snsConfig = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithSqsConfig applies SQS service configuration.
func (c *FlociContainer) WithSqsConfig(cfg SqsConfig) *FlociContainer {
	c.sqsConfig = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithSsmConfig applies SSM service configuration.
func (c *FlociContainer) WithSsmConfig(cfg SsmConfig) *FlociContainer {
	c.ssmConfig = cfg
	cfg.applyEnvVars(c)
	return c
}

// WithStepFunctionsConfig applies Step Functions service configuration.
func (c *FlociContainer) WithStepFunctionsConfig(cfg StepFunctionsConfig) *FlociContainer {
	c.stepFunctionsConfig = cfg
	cfg.applyEnvVars(c)
	return c
}

// Start launches the Floci container and waits for it to be ready.
//
// Deprecated: use Run.
func (c *FlociContainer) Start(ctx context.Context) (*StartedFlociContainer, error) {
	return c.run(ctx)
}

// needsDockerSocket reports whether the host Docker socket should be mounted for the
// given final container environment: an explicit WithDockerSocket override wins,
// otherwise any enabled service that spawns sibling containers requires it.
func (c *FlociContainer) needsDockerSocket(env map[string]string) bool {
	return c.request().NeedsDockerSocket(env)
}

// request is the builder's state as the core's start request.
func (c *FlociContainer) request() core.Request {
	ports := make([]int, 0, len(c.ports))
	for port := range c.ports {
		ports = append(ports, port)
	}
	env := maps.Clone(c.envVars)
	if env[awsDescriptor.ResourceNamespaceEnv()] == c.generatedNamespace {
		env[awsDescriptor.ResourceNamespaceEnv()] = core.NewNamespace()
	}
	return core.Request{
		Descriptor:       awsDescriptor,
		Image:            c.image,
		Env:              env,
		Ports:            ports,
		DockerSocket:     c.dockerSocketOverride,
		DedicatedNetwork: c.dedicatedNetwork,
	}
}

func (c *FlociContainer) run(ctx context.Context, customizers ...testcontainers.ContainerCustomizer) (*Container, error) {
	cc, env, err := core.Run(ctx, c.request(), customizers...)
	if cc == nil {
		return nil, err
	}
	return &Container{
		Container:        cc.Container,
		core:             cc,
		region:           envOr(env, "FLOCI_DEFAULT_REGION", DefaultRegion),
		availabilityZone: envOr(env, "FLOCI_DEFAULT_AVAILABILITY_ZONE", DefaultAvailabilityZone),
		accountID:        envOr(env, "FLOCI_DEFAULT_ACCOUNT_ID", DefaultAccountID),
	}, err
}

// requestOptions returns the customizers Run passes to testcontainers.Run (see core.Request.Options).
func (c *FlociContainer) requestOptions(network *testcontainers.DockerNetwork, customizers []testcontainers.ContainerCustomizer) ([]testcontainers.ContainerCustomizer, *map[string]string) {
	return c.request().Options(network, customizers)
}

func envOr(env map[string]string, key, fallback string) string {
	if v := env[key]; v != "" {
		return v
	}
	return fallback
}

func (c *FlociContainer) applyAllConfigs() {
	c.acmConfig.applyEnvVars(c)
	c.apiGatewayConfig.applyEnvVars(c)
	c.apiGatewayV2Config.applyEnvVars(c)
	c.appConfigConfig.applyEnvVars(c)
	c.appConfigDataConfig.applyEnvVars(c)
	c.athenaConfig.applyEnvVars(c)
	c.bedrockRuntimeConfig.applyEnvVars(c)
	c.cloudFormationConfig.applyEnvVars(c)
	c.cloudWatchLogsConfig.applyEnvVars(c)
	c.cloudWatchMetricsConfig.applyEnvVars(c)
	c.codeBuildConfig.applyEnvVars(c)
	c.codeDeployConfig.applyEnvVars(c)
	c.cognitoConfig.applyEnvVars(c)
	c.dynamoDbConfig.applyEnvVars(c)
	c.ec2Config.applyEnvVars(c)
	c.ecrConfig.applyEnvVars(c)
	c.ecsConfig.applyEnvVars(c)
	c.eksConfig.applyEnvVars(c)
	c.elastiCacheConfig.applyEnvVars(c)
	c.elbV2Config.applyEnvVars(c)
	c.eventBridgeConfig.applyEnvVars(c)
	c.firehoseConfig.applyEnvVars(c)
	c.glueConfig.applyEnvVars(c)
	c.iamConfig.applyEnvVars(c)
	c.kinesisConfig.applyEnvVars(c)
	c.kmsConfig.applyEnvVars(c)
	c.lambdaConfig.applyEnvVars(c)
	c.mskConfig.applyEnvVars(c)
	c.openSearchConfig.applyEnvVars(c)
	c.pipesConfig.applyEnvVars(c)
	c.rdsConfig.applyEnvVars(c)
	c.resourceGroupsTaggingConfig.applyEnvVars(c)
	c.s3Config.applyEnvVars(c)
	c.schedulerConfig.applyEnvVars(c)
	c.secretsManagerConfig.applyEnvVars(c)
	c.sesConfig.applyEnvVars(c)
	c.sesV2Config.applyEnvVars(c)
	c.snsConfig.applyEnvVars(c)
	c.sqsConfig.applyEnvVars(c)
	c.ssmConfig.applyEnvVars(c)
	c.stepFunctionsConfig.applyEnvVars(c)
	c.refreshExposedPorts()
}

func (c *FlociContainer) refreshExposedPorts() {
	// Rebuild from scratch so re-applying a config with an Expose flag turned
	// off also removes the ports it previously added.
	c.ports = map[int]struct{}{flociPort: {}}
	c.ecrConfig.applyExposedPorts(c)
	c.eksConfig.applyExposedPorts(c)
	c.elastiCacheConfig.applyExposedPorts(c)
	c.lambdaConfig.applyExposedPorts(c)
	c.openSearchConfig.applyExposedPorts(c)
	c.rdsConfig.applyExposedPorts(c)
}

// Container is a running Floci container. It embeds testcontainers.Container, as it always has,
// so Logs, Exec, Inspect and testcontainers.CleanupContainer work on it; the shared core behind
// GetEndpoint, GetMappedPort, Reset and Terminate (which also removes the dedicated network and
// Floci's sibling containers) is kept in an unexported field.
type Container struct {
	testcontainers.Container
	core             *core.Container
	region           string
	availabilityZone string
	accountID        string
}

// Container must stay a drop-in testcontainers.Container (CleanupContainer, Logs, Exec).
var _ testcontainers.Container = (*Container)(nil)

// StartedFlociContainer is the type returned by the deprecated Start.
//
// Deprecated: use Container.
type StartedFlociContainer = Container

// GetEndpoint returns the base URL, e.g. "http://localhost:32768".
func (s *Container) GetEndpoint() string {
	if s.core == nil {
		return ""
	}
	return s.core.GetEndpoint()
}

// GetDedicatedNetworkName returns the dedicated Docker network name, or "" if none.
func (s *Container) GetDedicatedNetworkName() string {
	if s.core == nil {
		return ""
	}
	return s.core.GetDedicatedNetworkName()
}

// GetResourceNamespace returns the prefix of the sibling containers' names.
func (s *Container) GetResourceNamespace() string {
	if s.core == nil {
		return ""
	}
	return s.core.GetResourceNamespace()
}

// GetMappedPort returns the host-mapped port for a given container port, asking the embedded
// Container (so a wrapper assigned to it is honoured).
func (s *Container) GetMappedPort(ctx context.Context, port int) (int, error) {
	mapped, err := s.MappedPort(ctx, fmt.Sprintf("%d/tcp", port))
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(mapped.Port())
}

// Reset wipes all emulator state (buckets, queues, tables, ...) without restarting.
func (s *Container) Reset(ctx context.Context) error {
	if s.core == nil {
		return errors.New("floci: Reset needs a container started by Run")
	}
	return s.core.Reset(ctx)
}

// Terminate stops and removes the container through the embedded Container (so a wrapper
// assigned to it runs its own cleanup), then removes Floci's sibling containers and the dedicated
// network, if any.
func (s *Container) Terminate(ctx context.Context, opts ...testcontainers.TerminateOption) error {
	err := s.Container.Terminate(ctx, opts...)
	if s.core != nil {
		err = errors.Join(err, s.core.Cleanup(ctx))
	}
	return err
}

// GetRegion returns the configured AWS region.
func (s *Container) GetRegion() string { return s.region }

// GetAccessKey returns the AWS access key (always "test").
func (s *Container) GetAccessKey() string { return DefaultAccessKey }

// GetSecretKey returns the AWS secret key (always "test").
func (s *Container) GetSecretKey() string { return DefaultSecretKey }

// GetAccountID returns the configured AWS account ID.
func (s *Container) GetAccountID() string { return s.accountID }

// GetAvailabilityZone returns the configured availability zone.
func (s *Container) GetAvailabilityZone() string { return s.availabilityZone }
