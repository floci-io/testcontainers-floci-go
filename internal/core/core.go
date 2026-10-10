// Package core is the cloud-independent part of the Floci Testcontainers modules: one
// cloud descriptor and one container lifecycle, shared by flociaws (and later flociaz,
// flocigcp and flocioci).
package core

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"strconv"
	"strings"
	"time"

	dockercontainer "github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/testcontainers/testcontainers-go"
	tcnetwork "github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

// DockerSocket is the host Docker socket path, mounted at the same path in the emulator.
const DockerSocket = "/var/run/docker.sock"

// SocketService is a service that spawns sibling containers and therefore needs the host
// Docker socket. Token is the SCREAMING_SNAKE name in <prefix>SERVICES_<Token>_ENABLED; a
// Mockable service spawns nothing while <prefix>SERVICES_<Token>_MOCK is true.
type SocketService struct {
	Token    string
	Mockable bool
}

// HostSetting is a service setting that must name a host the test process can reach, such
// as the AWS RDS endpoint host. When the service is enabled and the setting is unset, the
// core sets it to the Docker host.
type HostSetting struct {
	Token   string
	Setting string
}

// NamespaceLabel is the label every Floci emulator puts on the sibling containers it spawns,
// holding its DOCKER_RESOURCE_NAMESPACE.
const NamespaceLabel = "floci_namespace"

// Descriptor holds the facts that tell one Floci emulator apart from another.
type Descriptor struct {
	Name           string // short cloud name, e.g. "aws"
	DefaultImage   string // e.g. "floci/floci:latest"
	Port           int    // the emulator's single edge port
	EnvPrefix      string // e.g. "FLOCI_" or "FLOCI_AZ_"
	HealthPath     string // answers 200 once the emulator is ready
	ResetPath      string // wipes all state on POST; empty if the cloud has none
	LogLevelEnv    string // env var that sets the emulator's log level
	StartupTimeout time.Duration
	SocketServices []SocketService
	HostSettings   []HostSetting
}

// PortSpec is the edge port in testcontainers' "4566/tcp" form.
func (d Descriptor) PortSpec() string { return fmt.Sprintf("%d/tcp", d.Port) }

// NetworkEnv names the Docker network sibling containers join.
func (d Descriptor) NetworkEnv() string { return d.EnvPrefix + "SERVICES_DOCKER_NETWORK" }

// ResourceNamespaceEnv prefixes sibling container names, keeping parallel runs apart.
func (d Descriptor) ResourceNamespaceEnv() string { return d.EnvPrefix + "DOCKER_RESOURCE_NAMESPACE" }

// ServiceEnv is the env var of one service setting, e.g. ServiceEnv("SQS", "ENABLED").
func (d Descriptor) ServiceEnv(token, setting string) string {
	return d.EnvPrefix + "SERVICES_" + token + "_" + setting
}

// DockerSocketRequired reports whether any enabled, non-mocked socket service in env spawns
// sibling containers. A missing _ENABLED key means enabled, which is Floci's default.
func (d Descriptor) DockerSocketRequired(env map[string]string) bool {
	for _, svc := range d.SocketServices {
		enabled, ok := env[d.ServiceEnv(svc.Token, "ENABLED")]
		if ok && !strings.EqualFold(enabled, "true") {
			continue
		}
		if svc.Mockable && strings.EqualFold(env[d.ServiceEnv(svc.Token, "MOCK")], "true") {
			continue
		}
		return true
	}
	return false
}

// serviceEnabled reports whether a service is enabled in env; a missing key means enabled.
func (d Descriptor) serviceEnabled(env map[string]string, token string) bool {
	enabled, ok := env[d.ServiceEnv(token, "ENABLED")]
	return !ok || strings.EqualFold(enabled, "true")
}

// NewNamespace returns a unique resource namespace such as "tc-1a2b3c4d".
func NewNamespace() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return "tc-" + hex.EncodeToString(b)
}

// Request is what a cloud module hands the core to start one emulator.
type Request struct {
	Descriptor       Descriptor
	Image            string
	Env              map[string]string
	Ports            []int
	DockerSocket     *bool // nil: detect from the final env; otherwise forced on or off
	DedicatedNetwork bool
}

// NeedsDockerSocket reports whether the socket is mounted for the given final env.
func (r Request) NeedsDockerSocket(env map[string]string) bool {
	if r.DockerSocket != nil {
		return *r.DockerSocket
	}
	return r.Descriptor.DockerSocketRequired(env)
}

// Options returns the customizers for testcontainers.Run: the module defaults, then the
// caller's customizers, then a final step that sees the fully customized request. That last
// step records the final environment and decides the Docker socket mount from it, so generic
// options such as testcontainers.WithEnv cannot leave either out of sync with the emulator.
func (r Request) Options(network *testcontainers.DockerNetwork, customizers []testcontainers.ContainerCustomizer) ([]testcontainers.ContainerCustomizer, *map[string]string) {
	d := r.Descriptor
	exposedPorts := make([]string, 0, len(r.Ports))
	for _, port := range r.Ports {
		exposedPorts = append(exposedPorts, fmt.Sprintf("%d/tcp", port))
	}

	opts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(exposedPorts...),
		testcontainers.WithEnv(maps.Clone(r.Env)),
		testcontainers.WithWaitStrategy(wait.ForHTTP(d.HealthPath).
			WithPort(d.PortSpec()).
			WithStatusCodeMatcher(func(status int) bool { return status == http.StatusOK }).
			WithStartupTimeout(d.StartupTimeout)),
	}
	if network != nil {
		opts = append(opts, tcnetwork.WithNetwork(nil, network))
	}
	opts = append(opts, customizers...)

	finalEnv := map[string]string{}
	opts = append(opts, testcontainers.CustomizeRequestOption(func(req *testcontainers.GenericContainerRequest) error {
		if err := r.setHostSettings(req); err != nil {
			return err
		}
		finalEnv = maps.Clone(req.Env)
		if r.NeedsDockerSocket(req.Env) {
			// Chain rather than replace, so a caller's WithHostConfigModifier still runs.
			prev := req.HostConfigModifier
			req.HostConfigModifier = func(hc *dockercontainer.HostConfig) {
				if prev != nil {
					prev(hc)
				}
				hc.Binds = append(hc.Binds, DockerSocket+":"+DockerSocket)
			}
		}
		return nil
	}))
	return opts, &finalEnv
}

// setHostSettings points every unset host setting of an enabled service at the Docker host,
// so the emulator advertises addresses the test process can reach. Without it Floci advertises
// sibling containers' bridge addresses, which only a Linux host can reach.
func (r Request) setHostSettings(req *testcontainers.GenericContainerRequest) error {
	host := ""
	for _, hs := range r.Descriptor.HostSettings {
		key := r.Descriptor.ServiceEnv(hs.Token, hs.Setting)
		if _, set := req.Env[key]; set || !r.Descriptor.serviceEnabled(req.Env, hs.Token) {
			continue
		}
		if host == "" {
			h, err := dockerHost()
			if err != nil {
				// No Docker to ask: the start fails on its own, so leave Floci's default.
				return nil
			}
			host = h
		}
		if req.Env == nil {
			req.Env = map[string]string{}
		}
		req.Env[key] = host
	}
	return nil
}

// dockerHost is the host the test process reaches published ports on, as testcontainers
// resolves it (TESTCONTAINERS_HOST_OVERRIDE, a remote daemon, or the gateway inside a container).
func dockerHost() (string, error) {
	provider, err := testcontainers.NewDockerProvider()
	if err != nil {
		return "", err
	}
	defer provider.Close()
	return provider.DaemonHost(context.Background())
}

// Run starts the emulator described by r. The returned container is non-nil whenever the
// container was created, even if Run also returns an error, so it can always be cleaned up.
// finalEnv is the environment the emulator actually started with.
func Run(ctx context.Context, r Request, customizers ...testcontainers.ContainerCustomizer) (c *Container, finalEnv map[string]string, err error) {
	d := r.Descriptor
	var network *testcontainers.DockerNetwork
	if r.DedicatedNetwork {
		network, err = tcnetwork.New(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("creating dedicated network: %w", err)
		}
		r.Env = maps.Clone(r.Env)
		r.Env[d.NetworkEnv()] = network.Name
	}

	opts, env := r.Options(network, customizers)
	ctr, err := testcontainers.Run(ctx, r.Image, opts...)
	if ctr != nil {
		c = &Container{Container: ctr, network: network, descriptor: d, resourceNamespace: (*env)[d.ResourceNamespaceEnv()]}
	} else if network != nil {
		// No container owns the network yet, so nothing else would remove it.
		_ = network.Remove(ctx)
	}
	if err != nil {
		return c, *env, fmt.Errorf("run floci %s: %w", d.Name, err)
	}

	c.endpoint, err = ctr.PortEndpoint(ctx, d.PortSpec(), "http")
	if err != nil {
		return c, *env, fmt.Errorf("getting floci %s endpoint: %w", d.Name, err)
	}
	return c, *env, nil
}

// Container is a running Floci emulator. It embeds testcontainers.Container, so Logs, Exec,
// Inspect and testcontainers.CleanupContainer work on it.
type Container struct {
	testcontainers.Container
	network           *testcontainers.DockerNetwork
	descriptor        Descriptor
	endpoint          string
	resourceNamespace string
}

// GetEndpoint returns the emulator's base URL, e.g. "http://localhost:32768".
func (c *Container) GetEndpoint() string { return c.endpoint }

// GetDedicatedNetworkName returns the dedicated Docker network name, or "" if none.
func (c *Container) GetDedicatedNetworkName() string {
	if c.network == nil {
		return ""
	}
	return c.network.Name
}

// GetResourceNamespace returns the prefix of the sibling containers' names.
func (c *Container) GetResourceNamespace() string { return c.resourceNamespace }

// GetMappedPort returns the host-mapped port for a given container port.
func (c *Container) GetMappedPort(ctx context.Context, port int) (int, error) {
	mapped, err := c.MappedPort(ctx, fmt.Sprintf("%d/tcp", port))
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(mapped.Port())
}

// Reset wipes all emulator state (buckets, queues, tables, ...) without restarting.
func (c *Container) Reset(ctx context.Context) error {
	if c.descriptor.ResetPath == "" {
		return fmt.Errorf("the %s emulator has no state reset", c.descriptor.Name)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+c.descriptor.ResetPath, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("state reset: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("state reset failed with HTTP %d", resp.StatusCode)
	}
	return nil
}

// Terminate stops and removes the container, then cleans up after it (see Cleanup).
func (c *Container) Terminate(ctx context.Context, opts ...testcontainers.TerminateOption) error {
	return errors.Join(c.Container.Terminate(ctx, opts...), c.Cleanup(ctx))
}

// Cleanup removes what outlives the emulator container: the sibling containers it spawned (those
// labelled with its resource namespace), then the dedicated network, if any. Siblings are managed
// by Floci, not by the testcontainers reaper, so without this they outlive the test run; a leaked
// one with a fixed host port, such as the AWS ECR registry, blocks the next run. Containers sharing
// a namespace (set through the env) lose their siblings too. Call it after the emulator container
// is gone; Terminate does both.
func (c *Container) Cleanup(ctx context.Context) error {
	if c.resourceNamespace != "" {
		removeSiblings(ctx, c.resourceNamespace)
	}
	if c.network != nil {
		if err := c.network.Remove(ctx); err != nil {
			return fmt.Errorf("removing network: %w", err)
		}
	}
	return nil
}

// removeSiblings removes every container labelled with the given resource namespace. It is
// best effort: teardown never fails over leftover siblings.
func removeSiblings(ctx context.Context, resourceNamespace string) {
	cli, err := testcontainers.NewDockerClientWithOpts(ctx)
	if err != nil {
		return
	}
	defer cli.Close()
	list, err := cli.ContainerList(ctx, client.ContainerListOptions{
		All:     true,
		Filters: client.Filters{}.Add("label", NamespaceLabel+"="+resourceNamespace),
	})
	if err != nil {
		return
	}
	for _, sibling := range list.Items {
		_, _ = cli.ContainerRemove(ctx, sibling.ID, client.ContainerRemoveOptions{Force: true, RemoveVolumes: true})
	}
}
