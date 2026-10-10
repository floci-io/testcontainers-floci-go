package floci_test

import (
	"testing"

	"github.com/testcontainers/testcontainers-go"

	floci "github.com/floci-io/testcontainers-floci-go"
	"github.com/floci-io/testcontainers-floci-go/flociaws"
)

// The root package is a set of aliases: its types are the flociaws types, so values pass
// between the two import paths freely and old code keeps compiling.
var (
	_ *flociaws.Container      = (*floci.Container)(nil)
	_ *flociaws.FlociContainer = (*floci.FlociContainer)(nil)
	_ flociaws.Option          = floci.WithRegion("eu-west-1")
	_ flociaws.S3Config        = floci.DefaultS3Config()
)

// Container still embeds testcontainers.Container under its usual field name, so code that sets
// or reads that field keeps compiling through both import paths.
var (
	_ = floci.Container{Container: testcontainers.Container(nil)}
	_ = func(c *floci.Container) testcontainers.Container { return c.Container }
)

func TestRootAliasesMatchFlociaws(t *testing.T) {
	if floci.DefaultRegion != flociaws.DefaultRegion {
		t.Fatalf("DefaultRegion = %q, want %q", floci.DefaultRegion, flociaws.DefaultRegion)
	}
	if got, want := floci.DefaultSqsConfig(), flociaws.DefaultSqsConfig(); got != want {
		t.Fatalf("DefaultSqsConfig() = %+v, want %+v", got, want)
	}
}
