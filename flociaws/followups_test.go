package flociaws_test

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	floci "github.com/floci-io/testcontainers-floci-go/flociaws"
	"github.com/moby/moby/client"
	"github.com/testcontainers/testcontainers-go"
)

func awsConfig(t *testing.T, c *floci.Container) aws.Config {
	t.Helper()
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(c.GetRegion()),
		config.WithBaseEndpoint(c.GetEndpoint()),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(c.GetAccessKey(), c.GetSecretKey(), "")))
	if err != nil {
		t.Fatalf("aws config: %v", err)
	}
	return cfg
}

func siblings(t *testing.T, resourceNamespace string) int {
	t.Helper()
	cli, err := testcontainers.NewDockerClientWithOpts(context.Background())
	if err != nil {
		t.Fatalf("docker client: %v", err)
	}
	defer cli.Close()
	list, err := cli.ContainerList(context.Background(), client.ContainerListOptions{
		All:     true,
		Filters: client.Filters{}.Add("label", "floci_namespace="+resourceNamespace),
	})
	if err != nil {
		t.Fatalf("listing containers: %v", err)
	}
	return len(list.Items)
}

// Terminate removes the sibling containers Floci spawned; the ECR registry is one, started by
// CreateRepository.
func TestRun_TerminateRemovesSiblings(t *testing.T) {
	ctx := context.Background()

	container, err := floci.Run(ctx, testImage)
	if err != nil {
		testcontainers.CleanupContainer(t, container)
		t.Fatalf("starting container: %v", err)
	}
	ns := container.GetResourceNamespace()

	_, err = ecr.NewFromConfig(awsConfig(t, container)).CreateRepository(ctx, &ecr.CreateRepositoryInput{
		RepositoryName: aws.String("cleanup-test"),
	})
	if err != nil {
		testcontainers.CleanupContainer(t, container)
		t.Fatalf("create repository: %v", err)
	}
	if siblings(t, ns) == 0 {
		testcontainers.CleanupContainer(t, container)
		t.Fatal("expected the ECR registry as a sibling container")
	}

	if err := container.Terminate(ctx); err != nil {
		t.Fatalf("terminate: %v", err)
	}
	if n := siblings(t, ns); n != 0 {
		t.Fatalf("%d sibling containers left after Terminate", n)
	}
}

// RDS endpoints point at the Docker host and its published proxy port, so a host-side client
// can connect on any OS (Floci otherwise advertises a bridge address only Linux can reach).
func TestRun_RdsEndpointIsReachable(t *testing.T) {
	ctx := context.Background()

	rdsCfg := floci.DefaultRdsConfig()
	rdsCfg.ProxyBasePort = 7010
	rdsCfg.ProxyPortCount = 3
	rdsCfg.ExposeProxyPorts = true
	container, err := floci.Run(ctx, testImage, floci.WithRdsConfig(rdsCfg))
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("starting container: %v", err)
	}

	client := rds.NewFromConfig(awsConfig(t, container))
	_, err = client.CreateDBInstance(ctx, &rds.CreateDBInstanceInput{
		DBInstanceIdentifier: aws.String("endpoint-test"),
		DBInstanceClass:      aws.String("db.t3.micro"),
		Engine:               aws.String("postgres"),
		MasterUsername:       aws.String("admin"),
		MasterUserPassword:   aws.String("secret123"),
		AllocatedStorage:     aws.Int32(20),
	})
	if err != nil {
		t.Fatalf("create db instance: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("host: %v", err)
	}
	deadline := time.Now().Add(2 * time.Minute)
	for {
		out, err := client.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{DBInstanceIdentifier: aws.String("endpoint-test")})
		if err == nil && len(out.DBInstances) == 1 && out.DBInstances[0].Endpoint != nil &&
			aws.ToString(out.DBInstances[0].DBInstanceStatus) == "available" {
			ep := out.DBInstances[0].Endpoint
			if got := aws.ToString(ep.Address); got != host {
				t.Fatalf("endpoint address = %q, want the Docker host %q", got, host)
			}
			conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, strconv.Itoa(int(aws.ToInt32(ep.Port)))), 5*time.Second)
			if err != nil {
				t.Fatalf("connecting to the advertised endpoint: %v", err)
			}
			conn.Close()
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("db instance not available in time (last error: %v)", err)
		}
		time.Sleep(2 * time.Second)
	}
}
