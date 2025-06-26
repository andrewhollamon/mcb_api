package queueservice

import (
	"context"
	"encoding/json"
	apierror "github.com/andrewhollamon/millioncheckboxes-api/internal/error"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/rs/zerolog/log"
)

func consumeSqsMessage(ctx context.Context, queueUrl string) error {
	/*	sqsClient, err := newSqsClient()
		if err != nil {
			return err
		}
	*/return nil
}

func publishSnsMessage(ctx context.Context, topicArn string, message *CheckboxActionMessage) apierror.APIError {
	snsClient, err := newSnsClient()
	if err != nil {
		return apierror.WrapWithCodeFromConstants(err, apierror.ErrQueueUnavailable, "failed to create SNS client")
	}

	jsonBytes, err := json.Marshal(message)
	if err != nil {
		return apierror.WrapWithCodeFromConstants(err, apierror.ErrInternalServer, "failed to marshal message to JSON")
	}

	publishInput := sns.PublishInput{
		TopicArn:               aws.String(topicArn),
		Message:                aws.String(string(jsonBytes)),
		MessageGroupId:         aws.String(message.Header.GroupId),
		MessageDeduplicationId: aws.String(message.Header.DeduplicationId),
	}

	pubOut, err := snsClient.Publish(ctx, &publishInput)
	if err != nil {
		return apierror.WrapWithCodeFromConstants(err, apierror.ErrQueueUnavailable, "failed to publish message to SNS")
	}

	log.Debug().Msg("Message sent to SNS")
	log.Debug().Str("MessageID: %v\n", *pubOut.MessageId)
	log.Debug().Str("SequenceNumber: %v\n", *pubOut.SequenceNumber)

	return nil
}

func configAndAuthN() (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithSharedConfigProfile("dev"))
	if err != nil {
		return cfg, apierror.WrapWithCodeFromConstants(err, apierror.ErrInternalServer, "failed to load AWS Config")
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
