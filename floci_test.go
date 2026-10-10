package floci_test

import (
	"testing"

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

func TestRootAliasesMatchFlociaws(t *testing.T) {
	if floci.DefaultRegion != flociaws.DefaultRegion {
		t.Fatalf("DefaultRegion = %q, want %q", floci.DefaultRegion, flociaws.DefaultRegion)
	}
	if got, want := floci.DefaultSqsConfig(), flociaws.DefaultSqsConfig(); got != want {
		t.Fatalf("DefaultSqsConfig() = %+v, want %+v", got, want)
	}
}
