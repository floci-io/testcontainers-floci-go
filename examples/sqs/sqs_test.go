package sqs_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	floci "github.com/floci-io/testcontainers-floci-go"
)

func TestSQSExample(t *testing.T) {
	ctx := context.Background()

	fc, err := floci.NewFlociContainer().Start(ctx)
	if err != nil {
		t.Fatalf("starting floci: %v", err)
	}
	t.Cleanup(func() { _ = fc.Stop(ctx) })

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

	client := sqs.NewFromConfig(cfg)

	createOut, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String("test-queue"),
	})
	if err != nil {
		t.Fatalf("creating queue: %v", err)
	}
	queueURL := createOut.QueueUrl
	t.Logf("created queue: %s", aws.ToString(queueURL))

	_, err = client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    queueURL,
		MessageBody: aws.String("hello from floci"),
	})
	if err != nil {
		t.Fatalf("sending message: %v", err)
	}
	t.Log("sent message")

	recv, err := client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            queueURL,
		MaxNumberOfMessages: 1,
		WaitTimeSeconds:     3,
	})
	if err != nil {
		t.Fatalf("receiving message: %v", err)
	}

	if len(recv.Messages) == 0 {
		t.Fatal("expected at least one message")
	}

	body := aws.ToString(recv.Messages[0].Body)
	if body != "hello from floci" {
		t.Errorf("expected body %q, got %q", "hello from floci", body)
	}
	t.Logf("received message: %q", body)

	_, err = client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      queueURL,
		ReceiptHandle: recv.Messages[0].ReceiptHandle,
	})
	if err != nil {
		t.Fatalf("deleting message: %v", err)
	}
	t.Log("deleted message")
}
