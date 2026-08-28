package floci

import "testing"

// Exposure flags must be declarative: re-applying a config with the flag
// turned off has to remove the ports a previous application added.
func TestExposedPorts_OptOutRemovesPreviouslyAddedPorts(t *testing.T) {
	c := NewFlociContainer()

	if got := len(c.ports); got != 1 {
		t.Fatalf("expected only the edge port exposed by default, got %d ports", got)
	}

	enabled := DefaultRdsConfig()
	enabled.ExposeProxyPorts = true
	c.WithRdsConfig(enabled)

	if got, want := len(c.ports), 1+enabled.ProxyPortCount; got != want {
		t.Fatalf("expected %d ports after enabling ExposeProxyPorts, got %d", want, got)
	}

	c.WithRdsConfig(DefaultRdsConfig())

	if got := len(c.ports); got != 1 {
		t.Fatalf("expected only the edge port after disabling ExposeProxyPorts, got %d ports", got)
	}
	if _, ok := c.ports[flociPort]; !ok {
		t.Fatalf("edge port %d missing from exposed ports", flociPort)
	}
}
