package core

import (
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testDescriptor = Descriptor{
	Name:           "test",
	DefaultImage:   "floci/floci-test:latest",
	Port:           4500,
	EnvPrefix:      "FLOCI_TEST_",
	HealthPath:     "/_floci-test/health",
	StartupTimeout: time.Minute,
}

// healthWaits applies the request's options to an empty container request, as Run does, and
// returns the health probes it ends up waiting for, flattened in order.
func healthWaits(t *testing.T, env map[string]string, customizers ...testcontainers.ContainerCustomizer) []*wait.HTTPStrategy {
	t.Helper()
	off := false
	opts, _ := Request{Descriptor: testDescriptor, Env: env, Ports: []int{4500}, DockerSocket: &off}.Options(nil, customizers)
	req := &testcontainers.GenericContainerRequest{}
	for _, opt := range opts {
		if err := opt.Customize(req); err != nil {
			t.Fatalf("customize: %v", err)
		}
	}
	var flatten func(s wait.Strategy) []*wait.HTTPStrategy
	flatten = func(s wait.Strategy) []*wait.HTTPStrategy {
		switch s := s.(type) {
		case *wait.HTTPStrategy:
			return []*wait.HTTPStrategy{s}
		case *wait.MultiStrategy:
			var out []*wait.HTTPStrategy
			for _, inner := range s.Strategies {
				out = append(out, flatten(inner)...)
			}
			return out
		case nil:
			return nil
		default:
			t.Fatalf("unexpected wait strategy %T", s)
			return nil
		}
	}
	return flatten(req.WaitingFor)
}

func TestDescriptor_TLSEnabled(t *testing.T) {
	if got := testDescriptor.TLSEnabledEnv(); got != "FLOCI_TEST_TLS_ENABLED" {
		t.Errorf("TLS switch = %q", got)
	}
	if got := (Descriptor{EnvPrefix: "FLOCI_"}).TLSEnabledEnv(); got != "FLOCI_TLS_ENABLED" {
		t.Errorf("AWS TLS switch = %q", got)
	}
	for value, want := range map[string]bool{"true": true, "TRUE": true, "false": false, "": false, "1": false} {
		if got := testDescriptor.TLSEnabled(map[string]string{"FLOCI_TEST_TLS_ENABLED": value}); got != want {
			t.Errorf("TLSEnabled(%q) = %v, want %v", value, got, want)
		}
	}
	if testDescriptor.TLSEnabled(nil) || testDescriptor.TLSEnabled(map[string]string{"FLOCI_OTHER_TLS_ENABLED": "true"}) {
		t.Error("TLS is off unless this emulator's switch is on")
	}
}

// Readiness is the plain-HTTP health probe, plus an HTTPS one exactly when the final environment
// turns TLS on, whether the module or a generic testcontainers.WithEnv set it.
func TestOptions_WaitSelection(t *testing.T) {
	tlsOn := map[string]string{"FLOCI_TEST_TLS_ENABLED": "true"}
	tests := []struct {
		name        string
		env         map[string]string
		customizers []testcontainers.ContainerCustomizer
		wantTLS     []bool // UseTLS of each health probe, in order
	}{
		{"default", map[string]string{}, nil, []bool{false}},
		{"TLS off explicitly", map[string]string{"FLOCI_TEST_TLS_ENABLED": "false"}, nil, []bool{false}},
		{"TLS on in the module env", tlsOn, nil, []bool{false, true}},
		{"TLS on through a generic WithEnv", map[string]string{},
			[]testcontainers.ContainerCustomizer{testcontainers.WithEnv(tlsOn)}, []bool{false, true}},
		{"TLS turned off again by a generic WithEnv", tlsOn,
			[]testcontainers.ContainerCustomizer{testcontainers.WithEnv(map[string]string{"FLOCI_TEST_TLS_ENABLED": "false"})}, []bool{false}},
		{"another emulator's switch", map[string]string{"FLOCI_OTHER_TLS_ENABLED": "true"}, nil, []bool{false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			probes := healthWaits(t, tt.env, tt.customizers...)
			if len(probes) != len(tt.wantTLS) {
				t.Fatalf("%d health probes, want %d", len(probes), len(tt.wantTLS))
			}
			for i, probe := range probes {
				if probe.UseTLS != tt.wantTLS[i] {
					t.Errorf("probe %d: UseTLS = %v, want %v", i, probe.UseTLS, tt.wantTLS[i])
				}
				if probe.Path != "/_floci-test/health" || probe.Port.Port() != "4500" {
					t.Errorf("probe %d: path %q port %q", i, probe.Path, probe.Port.Port())
				}
				if probe.UseTLS && (probe.TLSConfig == nil || !probe.TLSConfig.InsecureSkipVerify || !probe.AllowInsecure) {
					t.Errorf("probe %d: the HTTPS probe must not validate the certificate", i)
				}
				if !probe.StatusCodeMatcher(200) || probe.StatusCodeMatcher(503) {
					t.Errorf("probe %d: only HTTP 200 is ready", i)
				}
			}
		})
	}
}

// A wait strategy the caller sets replaces the module's; the HTTPS probe is still added after it.
func TestOptions_HTTPSWaitFollowsACallersStrategy(t *testing.T) {
	custom := wait.ForHTTP("/custom").WithPort("4500/tcp")
	probes := healthWaits(t, map[string]string{"FLOCI_TEST_TLS_ENABLED": "true"}, testcontainers.WithWaitStrategy(custom))
	if len(probes) != 2 || probes[0].Path != "/custom" || probes[0].UseTLS || !probes[1].UseTLS || probes[1].Path != "/_floci-test/health" {
		t.Fatalf("probes = %+v", probes)
	}
}
