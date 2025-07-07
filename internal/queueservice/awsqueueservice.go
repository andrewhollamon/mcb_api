package queueservice

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	apiconfig "github.com/andrewhollamon/millioncheckboxes-api/internal/config"
	apierror "github.com/andrewhollamon/millioncheckboxes-api/internal/error"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/rs/zerolog/log"
)

type awsQueueProvider struct{}

type SqsMessage struct {
	MessageId      string
	ReceiptHandle  string
	Body           string
	GroupId        string
	SequenceNumber string
	Attributes     map[string]string
}

func (m *SqsMessage) UnmarshalBody(v interface{}) apierror.APIError {
	err := json.Unmarshal([]byte(m.Body), v)
	if err != nil {
		return apierror.WrapWithCodeFromConstants(err, apierror.ErrInternalServer, fmt.Sprintf("Could not unmarshal message json into type %T", v))
	}
	return nil
}

func (a *awsQueueProvider) PullCheckboxActionMessages(ctx context.Context) ([]Message, apierror.APIError) {
	appconfig := apiconfig.GetConfig()

	sqsClient, err := a.newSqsClient(ctx, appconfig.GetString("AWS_AUTH_PROFILE_NAME"))
	if err != nil {
		log.Error().Err(err).Msg("failed to create SQS client")
		return nil, apierror.WrapWithCodeFromConstants(err, apierror.ErrQueueUnavailable, "failed to create SQS client")
	}
	log.Debug().Msg("SQS client created")

	queueUrl := appconfig.GetString("AWS_SQS_CHECKBOXACTION_BASE_URL") + appconfig.GetString("AWS_SQS_CHECKBOXACTION_CONSUMER1")
	log.Debug().Msgf("Pulling messages from SQS queue %s", queueUrl)
	result, sqserr := sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueUrl),
		MaxNumberOfMessages: appconfig.GetInt32("AWS_SQS_BATCHSIZE"),
		WaitTimeSeconds:     appconfig.GetInt32("AWS_SQS_WAITTIMESECONDS"),
		VisibilityTimeout:   appconfig.GetInt32("AWS_SQS_VISIBILITYTIMEOUT"),
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
	log.Debug().Msgf("Received %d messages from SQS", len(result.Messages))

	messages := make([]Message, 0, len(result.Messages))
	for _, resultMessage := range result.Messages {
		msg := Message{
			MessageId:     aws.ToString(resultMessage.MessageId),
			ReceiptHandle: aws.ToString(resultMessage.ReceiptHandle),
			Body:          aws.ToString(resultMessage.Body),
			Attributes:    make(map[string]string),
		}

		// Extract FIFO-specific attributes
		if groupID, ok := resultMessage.Attributes["MessageGroupId"]; ok {
			msg.GroupId = groupID
		}
		if seqNum, ok := resultMessage.Attributes["SequenceNumber"]; ok {
			msg.SequenceNumber = seqNum
		}

		for k, v := range resultMessage.Attributes {
			msg.Attributes[k] = v
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

func (a *awsQueueProvider) DeleteMessage(ctx context.Context, message *Message) apierror.APIError {
	appconfig := apiconfig.GetConfig()

	queueUrl := appconfig.GetString("AWS_SQS_CHECKBOXACTION_BASE_URL") + appconfig.GetString("AWS_SQS_CHECKBOXACTION_CONSUMER1")
	log.Debug().Msgf("Preparing to delete messages from SQS queue %s", queueUrl)

	sqsClient, apierr := a.newSqsClient(ctx, appconfig.GetString("AWS_AUTH_PROFILE_NAME"))
	if apierr != nil {
		log.Error().Err(apierr).Msg("failed to create SQS client")
		return apierror.WrapWithCodeFromConstants(apierr, apierror.ErrQueueUnavailable, "failed to create SQS client")
	}
	log.Debug().Msg("SQS client created")

	_, err := sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueUrl),
		ReceiptHandle: aws.String(message.ReceiptHandle),
	})
	if err != nil {
		log.Error().Err(err).Msgf("failed to delete message ID %s and receipt handle %s from SQS queue '%s'", message.MessageId, message.ReceiptHandle, queueUrl)
		return apierror.WrapWithCodeFromConstants(err, apierror.ErrQueueUnavailable, fmt.Sprintf("failed to delete message ID %s and receipt handle %s from SQS queue '%s'", message.MessageId, message.ReceiptHandle, queueUrl))
	}

	return nil
}

func (a *awsQueueProvider) PublishCheckboxAction(ctx context.Context, message *CheckboxActionMessage) (PublishMessageResult, apierror.APIError) {
	appconfig := apiconfig.GetConfig()
	topicArn := appconfig.GetString("AWS_SNS_CHECKBOXACTION_TOPIC_ARN")

	snsClient, err := a.newSnsClient(ctx, appconfig.GetString("AWS_AUTH_PROFILE_NAME"))
	if err != nil {
		return PublishMessageResult{}, apierror.WrapWithCodeFromConstants(err, apierror.ErrQueueUnavailable, "failed to create SNS client")
	}

	jsonBytes, baseerr := json.Marshal(message)
	if baseerr != nil {
		log.Error().Err(baseerr).Msg("failed to marshal message to JSON")
		return PublishMessageResult{}, apierror.WrapWithCodeFromConstants(baseerr, apierror.ErrInternalServer, "failed to marshal message to JSON")
	}

	publishInput := sns.PublishInput{
		TopicArn:               aws.String(topicArn),
		Message:                aws.String(string(jsonBytes)),
		MessageGroupId:         aws.String(message.Header.GroupId),
		MessageDeduplicationId: aws.String(message.Header.DeduplicationId),
	}

	fmt.Println("Publishing message to SNS")
	pubOut, baseerr := snsClient.Publish(ctx, &publishInput)
	if baseerr != nil {
		log.Error().Err(baseerr).Msg("failed to publish message to SNS")
		return PublishMessageResult{}, apierror.WrapWithCodeFromConstants(baseerr, apierror.ErrQueueUnavailable, "failed to publish message to SNS")
	}

	log.Debug().Msg("Message sent to SNS")
	log.Debug().Str("MessageID", aws.ToString(pubOut.MessageId)).Msg("SNS publish result")
	log.Debug().Str("SequenceNumber", aws.ToString(pubOut.SequenceNumber)).Msg("SNS publish result")

	return PublishMessageResult{
		MessageId:      aws.ToString(pubOut.MessageId),
		SequenceNumber: aws.ToString(pubOut.SequenceNumber),
		PublishTime:    time.Now(),
	}, nil
}

func (a *awsQueueProvider) configAndAuthN(ctx context.Context, awsprofilename string) (aws.Config, apierror.APIError) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithSharedConfigProfile(awsprofilename),
		config.WithRegion("us-east-1"))
	if err != nil {
		log.Error().Err(err).Msg("failed to load AWS Config")
		return cfg, apierror.WrapWithCodeFromConstants(err, apierror.ErrQueueUnavailable, "failed to load AWS Config")
	}
	fmt.Println("AWS config loaded")
	return cfg, nil
}

func (a *awsQueueProvider) newSnsClient(ctx context.Context, awsprofilename string) (*sns.Client, apierror.APIError) {
	cfg, err := a.configAndAuthN(ctx, awsprofilename)
	if err != nil {
		log.Error().Err(err).Msg("failed to create SNS client")
		return nil, apierror.WrapWithCodeFromConstants(err, apierror.ErrQueueUnavailable, "failed to create SNS client")
	}
	client := sns.NewFromConfig(cfg)
	fmt.Println("SNS client created", client)
	return client, nil
}

func (a *awsQueueProvider) newSqsClient(ctx context.Context, awsprofilename string) (*sqs.Client, apierror.APIError) {
	cfg, err := a.configAndAuthN(ctx, awsprofilename)
	if err != nil {
		log.Error().Err(err).Msg("failed to create SQS client")
		return nil, apierror.WrapWithCodeFromConstants(err, apierror.ErrQueueUnavailable, "failed to create SQS client")
	}

	return sqs.NewFromConfig(cfg), nil
}
