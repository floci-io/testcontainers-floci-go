package lambda_test

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	ltypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"

	floci "github.com/floci-io/testcontainers-floci-go"
)

func TestLambdaExample(t *testing.T) {
	ctx := context.Background()

	fc, err := floci.NewFlociContainer().
		WithDedicatedNetwork().
		Start(ctx)
	if err != nil {
		t.Fatalf("starting floci: %v", err)
	}
	t.Cleanup(func() { _ = fc.Stop(ctx) })

	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "bootstrap")

	cmd := exec.CommandContext(ctx, "go", "build", "-o", binaryPath, "./handler")
	cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64", "CGO_ENABLED=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("building handler: %v\n%s", err, out)
	}
	t.Log("handler built")

	zipData, err := buildHandlerZip(binaryPath)
	if err != nil {
		t.Fatalf("zipping handler: %v", err)
	}

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(fc.GetRegion()),
		config.WithBaseEndpoint(fc.GetEndpoint()),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			fc.GetAccessKey(), fc.GetSecretKey(), "",
		)),
	)
	if err != nil {
		t.Fatalf("loading AWS config: %v", err)
	}

	client := lambda.NewFromConfig(cfg)

	roleARN := "arn:aws:iam::" + fc.GetAccountID() + ":role/lambda-role"

	_, err = client.CreateFunction(ctx, &lambda.CreateFunctionInput{
		FunctionName:  aws.String("hello-fn"),
		Runtime:       ltypes.RuntimeProvidedal2023,
		Role:          aws.String(roleARN),
		Handler:       aws.String("bootstrap"),
		Architectures: []ltypes.Architecture{ltypes.ArchitectureX8664},
		Code: &ltypes.FunctionCode{
			ZipFile: zipData,
		},
	})
	if err != nil {
		t.Fatalf("creating function: %v", err)
	}
	t.Log("created function: hello-fn")

	waiter := lambda.NewFunctionActiveV2Waiter(client)
	if err := waiter.Wait(ctx, &lambda.GetFunctionInput{
		FunctionName: aws.String("hello-fn"),
	}, 60*time.Second); err != nil {
		t.Fatalf("waiting for function to be active: %v", err)
	}
	t.Log("function is active")

	result, err := client.Invoke(ctx, &lambda.InvokeInput{
		FunctionName: aws.String("hello-fn"),
	})
	if err != nil {
		t.Fatalf("invoking function: %v", err)
	}

	var response struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(result.Payload, &response); err != nil {
		t.Fatalf("parsing response: %v", err)
	}
	if response.Message != "hello from lambda" {
		t.Errorf("expected message %q, got %q", "hello from lambda", response.Message)
	}
	t.Logf("lambda response: %s", response.Message)
}

func buildHandlerZip(binaryPath string) ([]byte, error) {
	data, err := os.ReadFile(binaryPath)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	hdr := &zip.FileHeader{
		Name:   "bootstrap",
		Method: zip.Deflate,
	}
	hdr.SetMode(0755)
	f, err := w.CreateHeader(hdr)
	if err != nil {
		return nil, err
	}
	if _, err := f.Write(data); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
