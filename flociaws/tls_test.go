package flociaws_test

import (
	"context"
	"crypto/tls"
	"net/http"
	"strings"
	"testing"

	floci "github.com/floci-io/testcontainers-floci-go/flociaws"
	"github.com/testcontainers/testcontainers-go"
)

// With TLS on, the emulator answers plain HTTP a moment before its HTTPS listener is up. Run waits
// for both, so the very first HTTPS request works: one attempt, no retry.
func TestRun_TLSReadyOnStart(t *testing.T) {
	container, err := floci.Run(context.Background(), testImage,
		testcontainers.WithEnv(map[string]string{"FLOCI_TLS_ENABLED": "true"}))
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("starting container: %v", err)
	}

	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	defer client.CloseIdleConnections()
	resp, err := client.Get("https://" + strings.TrimPrefix(container.GetEndpoint(), "http://") + "/_floci/health")
	if err != nil {
		t.Fatalf("first HTTPS request after Run: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("health over HTTPS returned HTTP %d", resp.StatusCode)
	}
}
