package flocioci_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/client"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/objectstorage"
	"github.com/testcontainers/testcontainers-go"

	"github.com/floci-io/testcontainers-floci-go/flocioci"
)

// run starts a container on the pinned image and registers its cleanup.
func run(t *testing.T, opts ...testcontainers.ContainerCustomizer) *flocioci.Container {
	t.Helper()
	container, err := flocioci.Run(context.Background(), testImage, opts...)
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("starting container: %v", err)
	}
	return container
}

// call sends one unsigned request that must succeed; it decodes the JSON answer into out when
// out is non-nil.
func call(t *testing.T, method, url string, body []byte, out any) []byte {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), method, url, bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
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

func bucketsURL(c *flocioci.Container) string {
	return c.GetEndpoint() + "/n/" + c.GetNamespace() + "/b"
}

func buckets(t *testing.T, c *flocioci.Container) []string {
	t.Helper()
	var list []struct {
		Name string `json:"name"`
	}
	call(t, http.MethodGet, bucketsURL(c)+"?compartmentId="+c.GetTenancyID(), nil, &list)
	var names []string
	for _, item := range list {
		names = append(names, item.Name)
	}
	return names
}

func createBucket(t *testing.T, c *flocioci.Container, name string) {
	t.Helper()
	call(t, http.MethodPost, bucketsURL(c), []byte(`{"name": "`+name+`", "compartmentId": "`+c.GetTenancyID()+`"}`), nil)
}

func TestRun_DefaultConfig(t *testing.T) {
	ctx := context.Background()
	container := run(t)

	endpoint := container.GetEndpoint()
	if !strings.HasPrefix(endpoint, "http://") {
		t.Errorf("endpoint = %q", endpoint)
	}
	if container.GetRegion() != flocioci.DefaultRegion || container.GetTenancyID() != flocioci.DefaultTenancyID ||
		container.GetNamespace() != flocioci.DefaultNamespace {
		t.Errorf("identity = %q %q %q", container.GetRegion(), container.GetTenancyID(), container.GetNamespace())
	}
	if container.GetUserID() != flocioci.DefaultUserID || container.GetFingerprint() != flocioci.DefaultFingerprint {
		t.Errorf("credentials = %q %q", container.GetUserID(), container.GetFingerprint())
	}
	if port, err := container.GetMappedPort(ctx, 4599); err != nil || port == 0 {
		t.Errorf("mapped port = %d, %v", port, err)
	}

	block, _ := pem.Decode([]byte(container.GetPrivateKeyPEM()))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		t.Fatalf("private key is not an RSA PEM block: %q", container.GetPrivateKeyPEM())
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err != nil || key.N.BitLen() != 2048 {
		t.Fatalf("private key: %v", err)
	}

	var health struct {
		Version string `json:"version"`
	}
	call(t, http.MethodGet, endpoint+"/_floci-oci/health", nil, &health)
	if _, tag, _ := strings.Cut(testImage, ":"); health.Version != tag {
		t.Errorf("emulator version = %q, want the pinned %q", health.Version, tag)
	}

	// The emulator reports the namespace GetNamespace promises.
	var namespace string
	call(t, http.MethodGet, endpoint+"/n", nil, &namespace)
	if namespace != container.GetNamespace() {
		t.Errorf("the emulator's namespace = %q, GetNamespace = %q", namespace, container.GetNamespace())
	}
}

// The identity helpers follow the identity the emulator was started with.
func TestRun_Identity(t *testing.T) {
	const tenancy = "ocid1.tenancy.oc1..mytenancy"
	container := run(t, flocioci.WithRegion("eu-frankfurt-1"), flocioci.WithTenancyID(tenancy), flocioci.WithNamespace("my-namespace"))

	if container.GetRegion() != "eu-frankfurt-1" || container.GetTenancyID() != tenancy || container.GetNamespace() != "my-namespace" {
		t.Errorf("identity = %q %q %q", container.GetRegion(), container.GetTenancyID(), container.GetNamespace())
	}
	var namespace string
	call(t, http.MethodGet, container.GetEndpoint()+"/n", nil, &namespace)
	if namespace != "my-namespace" {
		t.Errorf("the emulator's namespace = %q", namespace)
	}
	// Unsigned requests land in the tenancy the module reports.
	createBucket(t, container, "documents")
	if got := buckets(t, container); len(got) != 1 || got[0] != "documents" {
		t.Errorf("buckets in %s = %v", tenancy, got)
	}
}

// With TLS on, the emulator answers plain HTTP a moment before its HTTPS listener is up. Run waits
// for both, so the very first HTTPS request works: one attempt, no retry.
func TestRun_TLSReadyOnStart(t *testing.T) {
	container := run(t, testcontainers.WithEnv(map[string]string{"FLOCI_OCI_TLS_ENABLED": "true"}))

	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	defer client.CloseIdleConnections()
	resp, err := client.Get("https://" + strings.TrimPrefix(container.GetEndpoint(), "http://") + "/_floci-oci/health")
	if err != nil {
		t.Fatalf("first HTTPS request after Run: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("health over HTTPS returned HTTP %d", resp.StatusCode)
	}
}

// Object Storage over plain, unsigned HTTP: the emulator does not ask for a signature by default.
func TestRun_ObjectStorageRoundTrip(t *testing.T) {
	container := run(t)

	createBucket(t, container, "documents")
	call(t, http.MethodPut, bucketsURL(container)+"/documents/o/hello.txt", []byte("hello from floci"), nil)
	if got := call(t, http.MethodGet, bucketsURL(container)+"/documents/o/hello.txt", nil, nil); string(got) != "hello from floci" {
		t.Fatalf("object = %q", got)
	}
	if got := buckets(t, container); len(got) != 1 || got[0] != "documents" {
		t.Fatalf("buckets = %v", got)
	}
}

// storageClient is an OCI SDK Object Storage client for the container, signing as provider.
func storageClient(t *testing.T, c *flocioci.Container, provider common.ConfigurationProvider) objectstorage.ObjectStorageClient {
	t.Helper()
	storage, err := objectstorage.NewObjectStorageClientWithConfigurationProvider(provider)
	if err != nil {
		t.Fatalf("object storage client: %v", err)
	}
	storage.Host = c.GetEndpoint()
	return storage
}

func sdkBuckets(t *testing.T, storage objectstorage.ObjectStorageClient, namespace, compartment string) []string {
	t.Helper()
	list, err := storage.ListBuckets(context.Background(), objectstorage.ListBucketsRequest{
		NamespaceName: common.String(namespace),
		CompartmentId: common.String(compartment),
	})
	if err != nil {
		t.Fatalf("list buckets: %v", err)
	}
	var names []string
	for _, item := range list.Items {
		names = append(names, *item.Name)
	}
	return names
}

// The credentials are ones the OCI SDK accepts, from the configuration file and passed directly:
// the SDK loads the profile and the key and signs with them, and the emulator, which keeps each
// tenancy's resources apart by the tenancy in the signature, serves those requests from the
// tenancy GetTenancyID names.
func TestRun_SDKCredentials(t *testing.T) {
	ctx := context.Background()
	container := run(t)

	dir := t.TempDir()
	keyFile, configFile := filepath.Join(dir, "floci_key.pem"), filepath.Join(dir, "config")
	if err := os.WriteFile(keyFile, []byte(container.GetPrivateKeyPEM()), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configFile, []byte(container.GetConfigFileContents(keyFile)), 0o600); err != nil {
		t.Fatal(err)
	}
	fromFile, err := common.ConfigurationProviderFromFile(configFile, "")
	if err != nil {
		t.Fatalf("configuration provider: %v", err)
	}
	if tenancy, err := fromFile.TenancyOCID(); err != nil || tenancy != container.GetTenancyID() {
		t.Errorf("tenancy from the configuration file = %q, %v", tenancy, err)
	}
	if keyID, err := fromFile.KeyID(); err != nil || keyID != container.GetTenancyID()+"/"+container.GetUserID()+"/"+container.GetFingerprint() {
		t.Errorf("key ID from the configuration file = %q, %v", keyID, err)
	}
	if region, err := fromFile.Region(); err != nil || region != container.GetRegion() {
		t.Errorf("region from the configuration file = %q, %v", region, err)
	}
	if key, err := fromFile.PrivateRSAKey(); err != nil || key == nil {
		t.Fatalf("private key from the configuration file: %v", err)
	}

	// The configuration file's client writes.
	storage := storageClient(t, container, fromFile)
	ns, err := storage.GetNamespace(ctx, objectstorage.GetNamespaceRequest{})
	if err != nil {
		t.Fatalf("get namespace: %v", err)
	}
	if *ns.Value != container.GetNamespace() {
		t.Errorf("namespace = %q, want %q", *ns.Value, container.GetNamespace())
	}
	if _, err := storage.CreateBucket(ctx, objectstorage.CreateBucketRequest{
		NamespaceName: ns.Value,
		CreateBucketDetails: objectstorage.CreateBucketDetails{
			Name:          common.String("documents"),
			CompartmentId: common.String(container.GetTenancyID()),
		},
	}); err != nil {
		t.Fatalf("create bucket: %v", err)
	}
	content := []byte("hello from floci")
	if _, err := storage.PutObject(ctx, objectstorage.PutObjectRequest{
		NamespaceName: ns.Value,
		BucketName:    common.String("documents"),
		ObjectName:    common.String("hello.txt"),
		ContentLength: common.Int64(int64(len(content))),
		PutObjectBody: io.NopCloser(bytes.NewReader(content)),
	}); err != nil {
		t.Fatalf("put object: %v", err)
	}

	// A client given the same credentials directly reads it back.
	raw := common.NewRawConfigurationProvider(container.GetTenancyID(), container.GetUserID(), container.GetRegion(),
		container.GetFingerprint(), container.GetPrivateKeyPEM(), nil)
	object, err := storageClient(t, container, raw).GetObject(ctx, objectstorage.GetObjectRequest{
		NamespaceName: ns.Value,
		BucketName:    common.String("documents"),
		ObjectName:    common.String("hello.txt"),
	})
	if err != nil {
		t.Fatalf("get object: %v", err)
	}
	defer object.Content.Close()
	if got, err := io.ReadAll(object.Content); err != nil || string(got) != "hello from floci" {
		t.Fatalf("object = %q, %v", got, err)
	}

	// Unsigned requests land in the same tenancy; another tenancy's credentials see none of it.
	if got := buckets(t, container); len(got) != 1 || got[0] != "documents" {
		t.Errorf("buckets seen without a signature = %v", got)
	}
	const otherTenancy = "ocid1.tenancy.oc1..someoneelse"
	other := common.NewRawConfigurationProvider(otherTenancy, container.GetUserID(), container.GetRegion(),
		container.GetFingerprint(), container.GetPrivateKeyPEM(), nil)
	// Each tenancy has its own namespace, and the emulator answers 404 for anyone else's.
	otherStorage := storageClient(t, container, other)
	otherNs, err := otherStorage.GetNamespace(ctx, objectstorage.GetNamespaceRequest{})
	if err != nil {
		t.Fatalf("get namespace as another tenancy: %v", err)
	}
	if *otherNs.Value == *ns.Value {
		t.Errorf("another tenancy shares namespace %q", *ns.Value)
	}
	if got := sdkBuckets(t, otherStorage, *otherNs.Value, otherTenancy); len(got) != 0 {
		t.Errorf("buckets seen by another tenancy = %v", got)
	}
	if got := sdkBuckets(t, storageClient(t, container, raw), *ns.Value, container.GetTenancyID()); len(got) != 1 {
		t.Errorf("buckets seen by the container's tenancy = %v", got)
	}
}

// A key the caller supplies is the one handed back, and a named profile loads in the SDK.
func TestRun_OwnKeyAndNamedProfile(t *testing.T) {
	donor := run(t)
	container := run(t, flocioci.WithPrivateKeyPEM(donor.GetPrivateKeyPEM()))
	if container.GetPrivateKeyPEM() != donor.GetPrivateKeyPEM() {
		t.Fatal("GetPrivateKeyPEM is not the supplied key")
	}

	dir := t.TempDir()
	keyFile, configFile := filepath.Join(dir, "floci_key.pem"), filepath.Join(dir, "config")
	if err := os.WriteFile(keyFile, []byte(container.GetPrivateKeyPEM()), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configFile, []byte(container.GetConfigFileContents(keyFile, "FLOCI")), 0o600); err != nil {
		t.Fatal(err)
	}
	provider, err := common.ConfigurationProviderFromFileWithProfile(configFile, "FLOCI", "")
	if err != nil {
		t.Fatalf("configuration provider: %v", err)
	}
	ns, err := storageClient(t, container, provider).GetNamespace(context.Background(), objectstorage.GetNamespaceRequest{})
	if err != nil {
		t.Fatalf("get namespace: %v", err)
	}
	if *ns.Value != container.GetNamespace() {
		t.Errorf("namespace = %q, want %q", *ns.Value, container.GetNamespace())
	}
}

// Reset wipes the emulator's state without restarting it.
func TestRun_Reset(t *testing.T) {
	container := run(t)

	createBucket(t, container, "to-be-wiped")
	if got := buckets(t, container); len(got) != 1 {
		t.Fatalf("buckets before reset = %v", got)
	}

	if err := container.Reset(context.Background()); err != nil {
		t.Fatalf("reset: %v", err)
	}
	if got := buckets(t, container); len(got) != 0 {
		t.Fatalf("buckets after reset = %v", got)
	}
	// The emulator still works after a reset.
	createBucket(t, container, "to-be-wiped")
}

// Every start gets its own resource namespace, so parallel runs never share sibling containers,
// and its own signing key.
func TestRun_NamespaceAndKeyAreUniquePerStart(t *testing.T) {
	first, second := run(t), run(t)
	a, b := first.GetResourceNamespace(), second.GetResourceNamespace()
	if !strings.HasPrefix(a, "tc-") || !strings.HasPrefix(b, "tc-") || a == b {
		t.Fatalf("namespaces = %q and %q, want two distinct tc- values", a, b)
	}
	if first.GetPrivateKeyPEM() == second.GetPrivateKeyPEM() {
		t.Fatal("two containers share a private key")
	}
}

func TestRun_DedicatedNetwork(t *testing.T) {
	container := run(t, flocioci.WithDedicatedNetwork())
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

// Terminate removes the sibling containers Floci spawned; an OKE cluster is one (a k3s
// container), started when the cluster is created through the Container Engine API.
func TestRun_TerminateRemovesSiblings(t *testing.T) {
	ctx := context.Background()
	container := run(t)
	ns := container.GetResourceNamespace()

	call(t, http.MethodPost, container.GetEndpoint()+"/20180222/clusters", []byte(`{"name": "cleanup-test", "compartmentId": "`+
		container.GetTenancyID()+`", "vcnId": "ocid1.vcn.oc1.iad.cleanuptest", "kubernetesVersion": "v1.30.1"}`), nil)

	deadline := time.Now().Add(3 * time.Minute)
	for siblings(t, ns) == 0 {
		if time.Now().After(deadline) {
			t.Fatal("expected the OKE cluster as a sibling container")
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
