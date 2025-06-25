package queueservice

import (
	"context"
	"encoding/json"
	"fmt"
	error0 "github.com/andrewhollamon/millioncheckboxes-api/internal/error"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func consumeSqsMessage(ctx context.Context, queueUrl string) error {
	sqsClient, err := newSqsClient()
	if err != nil {
		return err
	}
}

func publishSnsMessage(ctx context.Context, topicArn string, message *CheckboxActionMessage) error {
	snsClient, err := newSnsClient()
	if err != nil {
		return err
	}

	jsonBytes, err := json.Marshal(message)
	if err != nil {
		return error0.NewQueueError("Failed to marshal message to JSON", err)
	}

	publishInput := sns.PublishInput{
		TopicArn:               aws.String(topicArn),
		Message:                aws.String(string(jsonBytes)),
		MessageGroupId:         aws.String(message.Header.GroupId),
		MessageDeduplicationId: aws.String(message.Header.DeduplicationId),
	}

	pubOut, err := snsClient.Publish(ctx, &publishInput)
	if err != nil {
		return error0.NewQueueError("Failed to publish message to SNS", err)
	}

	fmt.Println("Message sent to SNS")
	fmt.Printf("MessageID: %v\n", pubOut.MessageId)
	fmt.Printf("SequenceNumber: %v\n", pubOut.SequenceNumber)
	fmt.Printf("Metadata: %v\n", pubOut.ResultMetadata)

	return nil
}

func configAndAuthN() (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithSharedConfigProfile("dev"))
	if err != nil {
		return cfg, error0.NewQueueError("Failed to load AWS Config", err)
	}
	return cfg, nil
}

func newSnsClient() (*sns.Client, error) {
	cfg, err := configAndAuthN()
	if err != nil {
		return nil, err
	}
	return sns.NewFromConfig(cfg), nil
}

func newSqsClient() (*sqs.Client, error) {
	cfg, err := configAndAuthN()
	if err != nil {
		return nil, err
	}
	return sqs.NewFromConfig(cfg), nil
}
