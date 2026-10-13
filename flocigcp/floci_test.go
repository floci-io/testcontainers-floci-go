package flocigcp_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/client"
	"github.com/testcontainers/testcontainers-go"

	"github.com/floci-io/testcontainers-floci-go/flocigcp"
)

// run starts a container on the pinned image and registers its cleanup.
func run(t *testing.T, opts ...testcontainers.ContainerCustomizer) *flocigcp.Container {
	t.Helper()
	container, err := flocigcp.Run(context.Background(), testImage, opts...)
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("starting container: %v", err)
	}
	return container
}

// call sends one request and decodes the JSON answer into out (when out is non-nil), failing
// the test on anything but a 2xx status.
func call(t *testing.T, method, url, contentType string, body []byte, out any) []byte {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), method, url, bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer resp.Body.Close()
	answer, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("%s %s: reading the answer: %v", method, url, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		t.Fatalf("%s %s: HTTP %d: %s", method, url, resp.StatusCode, answer)
	}
	if out != nil {
		if err := json.Unmarshal(answer, out); err != nil {
			t.Fatalf("%s %s: decoding %q: %v", method, url, answer, err)
		}
	}
	return answer
}

func buckets(t *testing.T, storage, project string) []string {
	t.Helper()
	var list struct {
		Items []struct {
			Name string `json:"name"`
		} `json:"items"`
	}
	call(t, http.MethodGet, storage+"/storage/v1/b?project="+project, "", nil, &list)
	var names []string
	for _, item := range list.Items {
		names = append(names, item.Name)
	}
	return names
}

func createBucket(t *testing.T, storage, project, name string) {
	t.Helper()
	call(t, http.MethodPost, storage+"/storage/v1/b?project="+project, "application/json", []byte(`{"name": "`+name+`"}`), nil)
}

func TestRun_DefaultConfig(t *testing.T) {
	ctx := context.Background()
	container := run(t)

	endpoint := container.GetEndpoint()
	if !strings.HasPrefix(endpoint, "http://") {
		t.Errorf("endpoint = %q", endpoint)
	}
	if container.GetProjectID() != flocigcp.DefaultProjectID {
		t.Errorf("project = %q", container.GetProjectID())
	}
	if port, err := container.GetMappedPort(ctx, 4588); err != nil || port == 0 {
		t.Errorf("mapped port = %d, %v", port, err)
	}

	// The emulator variables: a bare host:port, except the Cloud Storage one, which is a URL.
	hostPort := strings.TrimPrefix(endpoint, "http://")
	for name, value := range container.GetEmulatorEnv() {
		want := hostPort
		switch name {
		case "STORAGE_EMULATOR_HOST":
			want = endpoint
		case "GOOGLE_CLOUD_PROJECT":
			want = flocigcp.DefaultProjectID
		}
		if value != want {
			t.Errorf("%s = %q, want %q", name, value, want)
		}
	}

	var health struct {
		Version string `json:"version"`
	}
	call(t, http.MethodGet, endpoint+"/_floci-gcp/health", "", nil, &health)
	if _, tag, _ := strings.Cut(testImage, ":"); health.Version != tag {
		t.Errorf("emulator version = %q, want the pinned %q", health.Version, tag)
	}

	var info struct {
		DefaultProject string `json:"defaultProject"`
	}
	call(t, http.MethodGet, endpoint+"/_floci-gcp/info", "", nil, &info)
	if info.DefaultProject != container.GetProjectID() {
		t.Errorf("the emulator's default project = %q, GetProjectID = %q", info.DefaultProject, container.GetProjectID())
	}
}

// GetProjectID follows the project the emulator was started with.
func TestRun_ProjectID(t *testing.T) {
	container := run(t, flocigcp.WithProjectID("my-project"))

	if container.GetProjectID() != "my-project" || container.GetEmulatorEnv()["GOOGLE_CLOUD_PROJECT"] != "my-project" {
		t.Errorf("project = %q, env = %v", container.GetProjectID(), container.GetEmulatorEnv())
	}
	var info struct {
		DefaultProject string `json:"defaultProject"`
	}
	call(t, http.MethodGet, container.GetEndpoint()+"/_floci-gcp/info", "", nil, &info)
	if info.DefaultProject != "my-project" {
		t.Errorf("the emulator's default project = %q", info.DefaultProject)
	}
}

// With TLS on, the emulator answers plain HTTP a moment before its HTTPS listener is up. Run waits
// for both, so the very first HTTPS request works: one attempt, no retry.
func TestRun_TLSReadyOnStart(t *testing.T) {
	container := run(t, testcontainers.WithEnv(map[string]string{"FLOCI_GCP_TLS_ENABLED": "true"}))

	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	defer client.CloseIdleConnections()
	resp, err := client.Get("https://" + strings.TrimPrefix(container.GetEndpoint(), "http://") + "/_floci-gcp/health")
	if err != nil {
		t.Fatalf("first HTTPS request after Run: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("health over HTTPS returned HTTP %d", resp.StatusCode)
	}
}

// Cloud Storage over its JSON API, at the address STORAGE_EMULATOR_HOST gives the client libraries.
func TestRun_StorageRoundTrip(t *testing.T) {
	container := run(t)
	storage := container.GetEmulatorEnv()["STORAGE_EMULATOR_HOST"]
	project := container.GetProjectID()

	createBucket(t, storage, project, "documents")
	call(t, http.MethodPost, storage+"/upload/storage/v1/b/documents/o?uploadType=media&name=hello.txt",
		"text/plain", []byte("hello from floci"), nil)
	if got := call(t, http.MethodGet, storage+"/storage/v1/b/documents/o/hello.txt?alt=media", "", nil, nil); string(got) != "hello from floci" {
		t.Fatalf("object = %q", got)
	}
	if got := buckets(t, storage, project); len(got) != 1 || got[0] != "documents" {
		t.Fatalf("buckets = %v", got)
	}
}

// Pub/Sub over its REST API, at the host:port PUBSUB_EMULATOR_HOST gives the client libraries.
func TestRun_PubSubRoundTrip(t *testing.T) {
	container := run(t)
	base := "http://" + container.GetEmulatorEnv()["PUBSUB_EMULATOR_HOST"] + "/v1/projects/" + container.GetProjectID()

	call(t, http.MethodPut, base+"/topics/orders", "application/json", []byte(`{}`), nil)
	call(t, http.MethodPut, base+"/subscriptions/orders-sub", "application/json",
		[]byte(`{"topic": "projects/`+container.GetProjectID()+`/topics/orders"}`), nil)

	var published struct {
		MessageIDs []string `json:"messageIds"`
	}
	call(t, http.MethodPost, base+"/topics/orders:publish", "application/json",
		[]byte(`{"messages": [{"data": "`+base64.StdEncoding.EncodeToString([]byte("order-1"))+`"}]}`), &published)
	if len(published.MessageIDs) != 1 {
		t.Fatalf("published = %+v", published)
	}

	var pulled struct {
		ReceivedMessages []struct {
			Message struct {
				Data      string `json:"data"`
				MessageID string `json:"messageId"`
			} `json:"message"`
		} `json:"receivedMessages"`
	}
	call(t, http.MethodPost, base+"/subscriptions/orders-sub:pull", "application/json", []byte(`{"maxMessages": 1}`), &pulled)
	if len(pulled.ReceivedMessages) != 1 {
		t.Fatalf("pulled = %+v", pulled)
	}
	data, err := base64.StdEncoding.DecodeString(pulled.ReceivedMessages[0].Message.Data)
	if err != nil || string(data) != "order-1" || pulled.ReceivedMessages[0].Message.MessageID != published.MessageIDs[0] {
		t.Fatalf("message = %+v (%q, %v), published %v", pulled.ReceivedMessages[0], data, err, published.MessageIDs)
	}
}

// Reset wipes the emulator's state without restarting it.
func TestRun_Reset(t *testing.T) {
	container := run(t)
	storage, project := container.GetEndpoint(), container.GetProjectID()

	createBucket(t, storage, project, "to-be-wiped")
	if got := buckets(t, storage, project); len(got) != 1 {
		t.Fatalf("buckets before reset = %v", got)
	}

	if err := container.Reset(context.Background()); err != nil {
		t.Fatalf("reset: %v", err)
	}
	if got := buckets(t, storage, project); len(got) != 0 {
		t.Fatalf("buckets after reset = %v", got)
	}
	// The emulator still works after a reset.
	createBucket(t, storage, project, "to-be-wiped")
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
	container := run(t, flocigcp.WithDedicatedNetwork())
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

// Terminate removes the sibling containers Floci spawned; a Cloud SQL instance is one (a
// PostgreSQL container), started when the instance is created through the Cloud SQL Admin API.
func TestRun_TerminateRemovesSiblings(t *testing.T) {
	ctx := context.Background()
	container := run(t)
	ns := container.GetResourceNamespace()

	call(t, http.MethodPost, container.GetEndpoint()+"/v1/projects/"+container.GetProjectID()+"/instances", "application/json",
		[]byte(`{"name": "cleanup-test", "databaseVersion": "POSTGRES_16", "region": "us-central1", "settings": {"tier": "db-f1-micro"}}`), nil)

	deadline := time.Now().Add(2 * time.Minute)
	for siblings(t, ns) == 0 {
		if time.Now().After(deadline) {
			t.Fatal("expected the Cloud SQL instance as a sibling container")
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
