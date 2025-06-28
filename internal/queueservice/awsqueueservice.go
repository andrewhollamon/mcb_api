package queueservice

import (
	"context"
	"encoding/json"
	apiconfig "github.com/andrewhollamon/millioncheckboxes-api/internal/config"
	apierror "github.com/andrewhollamon/millioncheckboxes-api/internal/error"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/rs/zerolog/log"
)

type SqsMessage struct {
	MessageId      string
	ReceiptHandle  string
	Body           string
	GroupId        string
	SequenceNumber string
	Attributes     map[string]string
}

func pullMessages(ctx context.Context) ([]SqsMessage, apierror.APIError) {
	myconfig := apiconfig.GetConfig()

	sqsClient, err := newSqsClient(ctx)
	if err != nil {
		return nil, apierror.WrapWithCodeFromConstants(err, apierror.ErrQueueUnavailable, "failed to create SQS client")
	}

	result, sqserr := sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(apiconfig.GetString("AWS_SQS_CHECKBOXACTION_BASE_URL" + "AWS_SQS_CHECKBOXACTION_CONSUMER1")),
		MaxNumberOfMessages: myconfig.GetInt32("AWS_SQS_BATCHSIZE"),
		WaitTimeSeconds:     myconfig.GetInt32("AWS_SQS_WAITTIMESECONDS"),
		VisibilityTimeout:   myconfig.GetInt32("AWS_SQS_VISIBILITYTIMEOUT"),
		MessageAttributeNames: []string{
			"All",
		},
		MessageSystemAttributeNames: []types.MessageSystemAttributeName{
			types.MessageSystemAttributeNameAll,
		},
	})
	if sqserr != nil {
		log.Error().Err(sqserr).Msg("failed to receive message from SQS")
		return nil, apierror.WrapWithCodeFromConstants(sqserr, apierror.ErrQueueUnavailable, "failed to receive message from SQS")
	}

	sqsMessages := make([]SqsMessage, 0, len(result.Messages))
	for _, resultMessage := range result.Messages {
		sqsMessage := SqsMessage{
			MessageId:     aws.ToString(resultMessage.MessageId),
			ReceiptHandle: aws.ToString(resultMessage.ReceiptHandle),
			Body:          aws.ToString(resultMessage.Body),
			Attributes:    make(map[string]string),
		}

		// Extract FIFO-specific attributes
		if groupID, ok := resultMessage.Attributes["MessageGroupId"]; ok {
			sqsMessage.GroupId = groupID
		}
		if seqNum, ok := resultMessage.Attributes["SequenceNumber"]; ok {
			sqsMessage.SequenceNumber = seqNum
		}

		for k, v := range resultMessage.Attributes {
			sqsMessage.Attributes[k] = v
		}

		sqsMessages = append(sqsMessages, sqsMessage)
	}

	return sqsMessages, nil
}

func publishSnsMessage(ctx context.Context, topicArn string, message *CheckboxActionMessage) apierror.APIError {
	snsClient, err := newSnsClient(ctx)
	if err != nil {
		return apierror.WrapWithCodeFromConstants(err, apierror.ErrQueueUnavailable, "failed to create SNS client")
	}

	jsonBytes, baseerr := json.Marshal(message)
	if baseerr != nil {
		return apierror.WrapWithCodeFromConstants(baseerr, apierror.ErrInternalServer, "failed to marshal message to JSON")
	}

	publishInput := sns.PublishInput{
		TopicArn:               aws.String(topicArn),
		Message:                aws.String(string(jsonBytes)),
		MessageGroupId:         aws.String(message.Header.GroupId),
		MessageDeduplicationId: aws.String(message.Header.DeduplicationId),
	}

	pubOut, baseerr := snsClient.Publish(ctx, &publishInput)
	if baseerr != nil {
		return apierror.WrapWithCodeFromConstants(baseerr, apierror.ErrQueueUnavailable, "failed to publish message to SNS")
	}

	log.Debug().Msg("Message sent to SNS")
	log.Debug().Str("MessageID: %v\n", *pubOut.MessageId)
	log.Debug().Str("SequenceNumber: %v\n", *pubOut.SequenceNumber)

	return nil
}

func configAndAuthN(ctx context.Context) (aws.Config, apierror.APIError) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithSharedConfigProfile("dev"))
	if err != nil {
		log.Error().Err(err).Msg("failed to load AWS Config")
		return cfg, apierror.WrapWithCodeFromConstants(err, apierror.ErrQueueUnavailable, "failed to load AWS Config")
	}
	return cfg, nil
}

func newSnsClient(ctx context.Context) (*sns.Client, apierror.APIError) {
	cfg, err := configAndAuthN(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to create SNS client")
		return nil, apierror.WrapWithCodeFromConstants(err, apierror.ErrQueueUnavailable, "failed to create SNS client")
	}
	return sns.NewFromConfig(cfg), nil
}

func newSqsClient(ctx context.Context) (*sqs.Client, apierror.APIError) {
	cfg, err := configAndAuthN(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to create SQS client")
		return nil, apierror.WrapWithCodeFromConstants(err, apierror.ErrQueueUnavailable, "failed to create SQS client")
	}

	return sqs.NewFromConfig(cfg), nil
}
