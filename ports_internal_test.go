package floci

import "testing"

// Each Expose flag must publish exactly its service's port range in
// addition to the edge port.
func TestExposedPorts_FlagsExposeServiceRanges(t *testing.T) {
	tests := []struct {
		name        string
		apply       func(c *FlociContainer)
		base, count int
	}{
		{
			name: "ecr registry ports",
			apply: func(c *FlociContainer) {
				cfg := DefaultEcrConfig()
				cfg.ExposeRegistryPorts = true
				c.WithEcrConfig(cfg)
			},
			base:  DefaultEcrConfig().RegistryBasePort,
			count: DefaultEcrConfig().RegistryPortCount,
		},
		{
			name: "eks api server ports",
			apply: func(c *FlociContainer) {
				cfg := DefaultEksConfig()
				cfg.ExposeApiServerPorts = true
				c.WithEksConfig(cfg)
			},
			base:  DefaultEksConfig().ApiServerBasePort,
			count: DefaultEksConfig().ApiServerPortCount,
		},
		{
			name: "elasticache proxy ports",
			apply: func(c *FlociContainer) {
				cfg := DefaultElastiCacheConfig()
				cfg.ExposeProxyPorts = true
				c.WithElastiCacheConfig(cfg)
			},
			base:  DefaultElastiCacheConfig().ProxyBasePort,
			count: DefaultElastiCacheConfig().ProxyPortCount,
		},
		{
			name: "lambda runtime ports",
			apply: func(c *FlociContainer) {
				cfg := DefaultLambdaConfig()
				cfg.ExposeRuntimePorts = true
				c.WithLambdaConfig(cfg)
			},
			base:  DefaultLambdaConfig().RuntimeApiBasePort,
			count: DefaultLambdaConfig().RuntimeApiPortCount,
		},
		{
			name: "opensearch proxy ports",
			apply: func(c *FlociContainer) {
				cfg := DefaultOpenSearchConfig()
				cfg.ExposeProxyPorts = true
				c.WithOpenSearchConfig(cfg)
			},
			base:  DefaultOpenSearchConfig().ProxyBasePort,
			count: DefaultOpenSearchConfig().ProxyPortCount,
		},
		{
			name: "rds proxy ports",
			apply: func(c *FlociContainer) {
				cfg := DefaultRdsConfig()
				cfg.ExposeProxyPorts = true
				c.WithRdsConfig(cfg)
			},
			base:  DefaultRdsConfig().ProxyBasePort,
			count: DefaultRdsConfig().ProxyPortCount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewFlociContainer()
			tt.apply(c)

			if got, want := len(c.ports), 1+tt.count; got != want {
				t.Fatalf("expected %d exposed ports, got %d", want, got)
			}
			if _, ok := c.ports[flociPort]; !ok {
				t.Errorf("edge port %d missing from exposed ports", flociPort)
			}
			for p := tt.base; p < tt.base+tt.count; p++ {
				if _, ok := c.ports[p]; !ok {
					t.Errorf("port %d missing from exposed ports", p)
				}
			}
		})
	}
}

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
