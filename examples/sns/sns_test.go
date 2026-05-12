package sns_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"

	floci "github.com/floci-io/testcontainers-floci-go"
)

func TestSNSExample(t *testing.T) {
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

	snsClient := sns.NewFromConfig(cfg)
	sqsClient := sqs.NewFromConfig(cfg)

	// Create SNS topic
	topicOut, err := snsClient.CreateTopic(ctx, &sns.CreateTopicInput{
		Name: aws.String("test-topic"),
	})
	if err != nil {
		t.Fatalf("creating topic: %v", err)
	}
	topicARN := topicOut.TopicArn
	t.Logf("created topic: %s", aws.ToString(topicARN))

	// Create SQS queue to receive SNS messages
	queueOut, err := sqsClient.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String("test-topic-queue"),
	})
	if err != nil {
		t.Fatalf("creating queue: %v", err)
	}
	queueURL := queueOut.QueueUrl
	t.Logf("created queue: %s", aws.ToString(queueURL))

	// Get queue ARN for the subscription
	attrOut, err := sqsClient.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       queueURL,
		AttributeNames: []sqstypes.QueueAttributeName{"QueueArn"},
	})
	if err != nil {
		t.Fatalf("getting queue attributes: %v", err)
	}
	queueARN := attrOut.Attributes["QueueArn"]
	t.Logf("queue ARN: %s", queueARN)

	// Subscribe SQS queue to SNS topic
	_, err = snsClient.Subscribe(ctx, &sns.SubscribeInput{
		TopicArn: topicARN,
		Protocol: aws.String("sqs"),
		Endpoint: aws.String(queueARN),
	})
	if err != nil {
		t.Fatalf("subscribing: %v", err)
	}
	t.Log("subscribed queue to topic")

	// Publish a message
	_, err = snsClient.Publish(ctx, &sns.PublishInput{
		TopicArn: topicARN,
		Message:  aws.String("hello from sns"),
	})
	if err != nil {
		t.Fatalf("publishing message: %v", err)
	}
	t.Log("published message")

	// Receive the message from SQS
	recv, err := sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            queueURL,
		MaxNumberOfMessages: 1,
		WaitTimeSeconds:     3,
	})
	if err != nil {
		t.Fatalf("receiving message: %v", err)
	}

	if len(recv.Messages) == 0 {
		t.Fatal("expected at least one message in queue")
	}

	var notification struct {
		Message string `json:"Message"`
	}
	if err := json.Unmarshal([]byte(aws.ToString(recv.Messages[0].Body)), &notification); err != nil {
		t.Fatalf("parsing SNS notification: %v", err)
	}
	if notification.Message != "hello from sns" {
		t.Errorf("expected message %q, got %q", "hello from sns", notification.Message)
	}
	t.Logf("received SNS notification via SQS: %s", notification.Message)
}
