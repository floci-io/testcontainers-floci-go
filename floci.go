// Package floci provides a Testcontainers module for Floci — a free, open-source local AWS emulator.
//
// Example:
//
//	floci, err := floci.NewFlociContainer().Start(ctx)
//	if err != nil { ... }
//	defer floci.Stop(ctx)
//
//	cfg, _ := config.LoadDefaultConfig(ctx,
//	    config.WithRegion(floci.GetRegion()),
//	    config.WithBaseEndpoint(floci.GetEndpoint()),
//	    config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
//	        floci.GetAccessKey(), floci.GetSecretKey(), "",
//	    )),
//	)
package floci

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"strconv"
	"time"

	dockercontainer "github.com/moby/moby/api/types/container"
	"github.com/testcontainers/testcontainers-go"
	tcnetwork "github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultImage   = "floci/floci:latest"
	dockerSocket   = "/var/run/docker.sock"
	flociPort      = 4566
	startupTimeout = 120 * time.Second

	DefaultRegion           = "us-east-1"
	DefaultAvailabilityZone = "us-east-1a"
	DefaultAccountID        = "000000000000"
	DefaultAccessKey        = "test"
	DefaultSecretKey        = "test"
)

// FlociContainer is a builder for a Floci testcontainer.
type FlociContainer struct {
	image            string
	envVars          map[string]string
	ports            map[int]struct{}
	dedicatedNetwork bool

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
func NewFlociContainer() *FlociContainer {
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
	c.applyAllConfigs()
	return c
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
func (c *FlociContainer) Start(ctx context.Context) (*StartedFlociContainer, error) {
	var dockerNetwork *testcontainers.DockerNetwork
	if c.dedicatedNetwork {
		var err error
		dockerNetwork, err = tcnetwork.New(ctx)
		if err != nil {
			return nil, fmt.Errorf("creating dedicated network: %w", err)
		}
		c.withEnv("FLOCI_SERVICES_DOCKER_NETWORK", dockerNetwork.Name)
	}

	exposedPorts := make([]string, 0, len(c.ports))
	for port := range c.ports {
		exposedPorts = append(exposedPorts, fmt.Sprintf("%d/tcp", port))
	}

	envCopy := maps.Clone(c.envVars)

	req := testcontainers.ContainerRequest{
		Image:        c.image,
		ExposedPorts: exposedPorts,
		Env:          envCopy,
		HostConfigModifier: func(hc *dockercontainer.HostConfig) {
			hc.Binds = append(hc.Binds, dockerSocket+":"+dockerSocket)
		},
		WaitingFor: wait.ForHTTP("/_floci/health").
			WithPort(fmt.Sprintf("%d/tcp", flociPort)).
			WithStatusCodeMatcher(func(status int) bool { return status == http.StatusOK }).
			WithStartupTimeout(startupTimeout),
	}

	if dockerNetwork != nil {
		req.Networks = []string{dockerNetwork.Name}
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		if dockerNetwork != nil {
			_ = dockerNetwork.Remove(ctx)
		}
		return nil, fmt.Errorf("starting floci container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		if dockerNetwork != nil {
			_ = dockerNetwork.Remove(ctx)
		}
		return nil, fmt.Errorf("getting container host: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, strconv.Itoa(flociPort))
	if err != nil {
		_ = container.Terminate(ctx)
		if dockerNetwork != nil {
			_ = dockerNetwork.Remove(ctx)
		}
		return nil, fmt.Errorf("getting mapped port: %w", err)
	}

	region := c.envVars["FLOCI_DEFAULT_REGION"]
	if region == "" {
		region = DefaultRegion
	}
	az := c.envVars["FLOCI_DEFAULT_AVAILABILITY_ZONE"]
	if az == "" {
		az = DefaultAvailabilityZone
	}
	accountID := c.envVars["FLOCI_DEFAULT_ACCOUNT_ID"]
	if accountID == "" {
		accountID = DefaultAccountID
	}

	dedicatedNetworkName := ""
	if dockerNetwork != nil {
		dedicatedNetworkName = dockerNetwork.Name
	}

	return &StartedFlociContainer{
		container:            container,
		network:              dockerNetwork,
		endpoint:             fmt.Sprintf("http://%s:%s", host, mappedPort.Port()),
		region:               region,
		availabilityZone:     az,
		accountID:            accountID,
		dedicatedNetworkName: dedicatedNetworkName,
	}, nil
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
	c.ecrConfig.applyExposedPorts(c)
	c.eksConfig.applyExposedPorts(c)
	c.elastiCacheConfig.applyExposedPorts(c)
	c.lambdaConfig.applyExposedPorts(c)
	c.openSearchConfig.applyExposedPorts(c)
	c.rdsConfig.applyExposedPorts(c)
}

// StartedFlociContainer is a running Floci container instance.
type StartedFlociContainer struct {
	container            testcontainers.Container
	network              *testcontainers.DockerNetwork
	endpoint             string
	region               string
	availabilityZone     string
	accountID            string
	dedicatedNetworkName string
}

// GetEndpoint returns the HTTP endpoint for Floci (e.g. "http://localhost:32768").
func (s *StartedFlociContainer) GetEndpoint() string { return s.endpoint }

// GetRegion returns the configured AWS region.
func (s *StartedFlociContainer) GetRegion() string { return s.region }

// GetAccessKey returns the AWS access key (always "test").
func (s *StartedFlociContainer) GetAccessKey() string { return DefaultAccessKey }

// GetSecretKey returns the AWS secret key (always "test").
func (s *StartedFlociContainer) GetSecretKey() string { return DefaultSecretKey }

// GetAccountID returns the configured AWS account ID.
func (s *StartedFlociContainer) GetAccountID() string { return s.accountID }

// GetAvailabilityZone returns the configured availability zone.
func (s *StartedFlociContainer) GetAvailabilityZone() string { return s.availabilityZone }

// GetDedicatedNetworkName returns the dedicated Docker network name, or empty string if none.
func (s *StartedFlociContainer) GetDedicatedNetworkName() string { return s.dedicatedNetworkName }

// GetMappedPort returns the host-mapped port for a given container port.
func (s *StartedFlociContainer) GetMappedPort(ctx context.Context, port int) (int, error) {
	mapped, err := s.container.MappedPort(ctx, strconv.Itoa(port))
	if err != nil {
		return 0, err
	}
	p, err := strconv.Atoi(mapped.Port())
	if err != nil {
		return 0, err
	}
	return p, nil
}

// Stop terminates the Floci container and removes any dedicated network.
func (s *StartedFlociContainer) Stop(ctx context.Context) error {
	if err := s.container.Terminate(ctx); err != nil {
		return fmt.Errorf("terminating floci container: %w", err)
	}
	if s.network != nil {
		if err := s.network.Remove(ctx); err != nil {
			return fmt.Errorf("removing network: %w", err)
		}
	}
	return nil
}
