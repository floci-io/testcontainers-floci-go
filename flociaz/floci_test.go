package flociaz_test

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
	"github.com/moby/moby/client"
	"github.com/testcontainers/testcontainers-go"

	"github.com/floci-io/testcontainers-floci-go/flociaz"
)

// run starts a container on the pinned image and registers its cleanup.
func run(t *testing.T, opts ...testcontainers.ContainerCustomizer) *flociaz.Container {
	t.Helper()
	container, err := flociaz.Run(context.Background(), testImage, opts...)
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("starting container: %v", err)
	}
	return container
}

func blobContainers(t *testing.T, c *azblob.Client) []string {
	t.Helper()
	var names []string
	pager := c.NewListContainersPager(nil)
	for pager.More() {
		page, err := pager.NextPage(context.Background())
		if err != nil {
			t.Fatalf("listing blob containers: %v", err)
		}
		for _, item := range page.ContainerItems {
			names = append(names, *item.Name)
		}
	}
	return names
}

func TestRun_DefaultConfig(t *testing.T) {
	ctx := context.Background()
	container := run(t)

	if !strings.HasPrefix(container.GetEndpoint(), "http://") {
		t.Errorf("endpoint = %q", container.GetEndpoint())
	}
	if want := "https://" + strings.TrimPrefix(container.GetEndpoint(), "http://"); container.GetHTTPSEndpoint() != want {
		t.Errorf("https endpoint = %q, want %q", container.GetHTTPSEndpoint(), want)
	}
	if container.GetSubscriptionID() != flociaz.DefaultSubscriptionID {
		t.Errorf("subscription = %q", container.GetSubscriptionID())
	}
	if container.GetTenantID() != flociaz.DefaultTenantID {
		t.Errorf("tenant = %q", container.GetTenantID())
	}
	if container.GetAccountName() != flociaz.DefaultAccountName || container.GetAccountKey() != flociaz.DefaultAccountKey {
		t.Errorf("storage account = %q / %q", container.GetAccountName(), container.GetAccountKey())
	}
	if port, err := container.GetMappedPort(ctx, 4577); err != nil || port == 0 {
		t.Errorf("mapped port = %d, %v", port, err)
	}

	resp, err := http.Get(container.GetEndpoint() + "/_floci/health")
	if err != nil {
		t.Fatalf("health: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("health returned HTTP %d", resp.StatusCode)
	}

	// TLS is off by default, so there is no certificate to hand out.
	if _, err := container.GetTLSCertificate(ctx); err == nil {
		t.Error("expected GetTLSCertificate to fail while TLS is disabled")
	}
}

// With TLS enabled the emulator serves its certificate, on the same port as plain HTTP.
func TestRun_TLSCertificate(t *testing.T) {
	container := run(t, testcontainers.WithEnv(map[string]string{"FLOCI_AZ_TLS_ENABLED": "true"}))

	pem, err := container.GetTLSCertificate(context.Background())
	if err != nil {
		t.Fatalf("tls certificate: %v", err)
	}
	if !strings.Contains(pem, "BEGIN CERTIFICATE") {
		t.Fatalf("not a PEM certificate: %q", pem)
	}
}

// With TLS on, the emulator answers plain HTTP a moment before its HTTPS listener is up. Run waits
// for both, so the very first HTTPS request works: one attempt, no retry.
func TestRun_TLSReadyOnStart(t *testing.T) {
	container := run(t, testcontainers.WithEnv(map[string]string{"FLOCI_AZ_TLS_ENABLED": "true"}))

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

func TestRun_BlobRoundTrip(t *testing.T) {
	ctx := context.Background()
	container := run(t)

	blobs, err := azblob.NewClientFromConnectionString(container.GetStorageConnectionString(), nil)
	if err != nil {
		t.Fatalf("blob client: %v", err)
	}
	if !strings.HasPrefix(blobs.URL(), container.GetBlobEndpoint()) {
		t.Errorf("blob client URL = %q, want it under %q", blobs.URL(), container.GetBlobEndpoint())
	}
	if _, err := blobs.CreateContainer(ctx, "documents", nil); err != nil {
		t.Fatalf("create container: %v", err)
	}
	if _, err := blobs.UploadBuffer(ctx, "documents", "hello.txt", []byte("hello from floci"), nil); err != nil {
		t.Fatalf("upload: %v", err)
	}
	download, err := blobs.DownloadStream(ctx, "documents", "hello.txt", nil)
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	defer download.Body.Close()
	body, err := io.ReadAll(download.Body)
	if err != nil {
		t.Fatalf("reading the blob: %v", err)
	}
	if string(body) != "hello from floci" {
		t.Fatalf("blob = %q", body)
	}
}

func TestRun_QueueRoundTrip(t *testing.T) {
	ctx := context.Background()
	container := run(t)

	service, err := azqueue.NewServiceClientFromConnectionString(container.GetStorageConnectionString(), nil)
	if err != nil {
		t.Fatalf("queue client: %v", err)
	}
	queue := service.NewQueueClient("orders")
	if _, err := queue.Create(ctx, nil); err != nil {
		t.Fatalf("create queue: %v", err)
	}
	if _, err := queue.EnqueueMessage(ctx, "order-1", nil); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	received, err := queue.DequeueMessage(ctx, nil)
	if err != nil {
		t.Fatalf("dequeue: %v", err)
	}
	if len(received.Messages) != 1 || *received.Messages[0].MessageText != "order-1" {
		t.Fatalf("messages = %+v", received.Messages)
	}
}

// Reset wipes the emulator's state without restarting it.
func TestRun_Reset(t *testing.T) {
	ctx := context.Background()
	container := run(t)

	blobs, err := azblob.NewClientFromConnectionString(container.GetStorageConnectionString(), nil)
	if err != nil {
		t.Fatalf("blob client: %v", err)
	}
	if _, err := blobs.CreateContainer(ctx, "to-be-wiped", nil); err != nil {
		t.Fatalf("create container: %v", err)
	}
	if got := blobContainers(t, blobs); len(got) != 1 {
		t.Fatalf("containers before reset = %v", got)
	}

	if err := container.Reset(ctx); err != nil {
		t.Fatalf("reset: %v", err)
	}
	if got := blobContainers(t, blobs); len(got) != 0 {
		t.Fatalf("containers after reset = %v", got)
	}
	// The emulator still works after a reset.
	if _, err := blobs.CreateContainer(ctx, "to-be-wiped", nil); err != nil {
		t.Fatalf("create container after reset: %v", err)
	}
}

// Every start gets its own resource namespace, so parallel runs never share sibling containers.
func TestRun_NamespaceIsUniquePerStart(t *testing.T) {
	first, second := run(t), run(t)
	a, b := first.GetResourceNamespace(), second.GetResourceNamespace()
	if !strings.HasPrefix(a, "tc-") || !strings.HasPrefix(b, "tc-") || a == b {
		t.Fatalf("namespaces = %q and %q, want two distinct tc- values", a, b)
	}
}

func TestRun_DedicatedNetwork(t *testing.T) {
	container := run(t, flociaz.WithDedicatedNetwork())
	if container.GetDedicatedNetworkName() == "" {
		t.Error("expected non-empty dedicated network name")
	}
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

// Terminate removes the sibling containers Floci spawned; a Redis cache is one (a Valkey
// container), started when the cache is created through the Resource Manager API.
func TestRun_TerminateRemovesSiblings(t *testing.T) {
	ctx := context.Background()
	container := run(t)
	ns := container.GetResourceNamespace()

	url := container.GetEndpoint() + "/subscriptions/" + container.GetSubscriptionID() +
		"/resourceGroups/rg/providers/Microsoft.Cache/redis/cleanup-test?api-version=2024-03-01"
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url,
		strings.NewReader(`{"location": "eastus", "properties": {"sku": {"name": "Basic", "family": "C", "capacity": 0}}}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create cache: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		t.Fatalf("create cache: HTTP %d: %s", resp.StatusCode, body)
	}

	deadline := time.Now().Add(2 * time.Minute)
	for siblings(t, ns) == 0 {
		if time.Now().After(deadline) {
			t.Fatal("expected the Redis cache as a sibling container")
		}
		time.Sleep(time.Second)
	}

	if err := container.Terminate(ctx); err != nil {
		t.Fatalf("terminate: %v", err)
	}
	if n := siblings(t, ns); n != 0 {
		t.Fatalf("%d sibling containers left after Terminate", n)
	}
}
