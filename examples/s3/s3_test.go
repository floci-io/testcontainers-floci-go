package s3_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	floci "github.com/floci-io/testcontainers-floci-go"
)

func TestS3Example(t *testing.T) {
	ctx := context.Background()

	// Start Floci
	fc, err := floci.NewFlociContainer().Start(ctx)
	if err != nil {
		t.Fatalf("starting floci: %v", err)
	}
	t.Cleanup(func() { _ = fc.Stop(ctx) })

	// Build an S3 client pointed at Floci
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

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true // required for local endpoints
	})

	bucket := fmt.Sprintf("test-bucket-%d", time.Now().UnixMilli())

	// Create bucket
	_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		t.Fatalf("creating bucket: %v", err)
	}
	t.Logf("created bucket: %s", bucket)

	// Upload a document
	body := "Hello from Floci!"
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String("hello.txt"),
		Body:   bytes.NewBufferString(body),
	})
	if err != nil {
		t.Fatalf("putting object: %v", err)
	}
	t.Log("uploaded hello.txt")

	// Upload a second document
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String("world.txt"),
		Body:   bytes.NewBufferString("Hello world!"),
	})
	if err != nil {
		t.Fatalf("putting object: %v", err)
	}
	t.Log("uploaded world.txt")

	// List objects
	list, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		t.Fatalf("listing objects: %v", err)
	}

	t.Logf("objects in %s:", bucket)
	for _, obj := range list.Contents {
		t.Logf("  - %s (%d bytes)", aws.ToString(obj.Key), obj.Size)
	}

	if len(list.Contents) != 2 {
		t.Errorf("expected 2 objects, got %d", len(list.Contents))
	}

	// Get object and verify content
	resp, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String("hello.txt"),
	})
	if err != nil {
		t.Fatalf("getting object: %v", err)
	}
	defer resp.Body.Close()

	got, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading object body: %v", err)
	}

	if string(got) != body {
		t.Errorf("expected body %q, got %q", body, string(got))
	}
	t.Logf("verified content of hello.txt: %q", string(got))
}
