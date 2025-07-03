package queueservice

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	apiconfig "github.com/andrewhollamon/millioncheckboxes-api/internal/config"
	apierror "github.com/andrewhollamon/millioncheckboxes-api/internal/error"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/logging"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/tracing"
)

type MessageHeader struct {
	PayloadSchemaVersion string `json:"payload_schema_version"`
	GroupId              string `json:"group_id"`
	DeduplicationId      string `json:"deduplication_id"`
}

type CheckboxActionPayload struct {
	Action      string    `json:"action"`
	CheckboxNbr int       `json:"checkbox_nbr"`
	UserUuid    string    `json:"user_uuid"`
	RequestUuid string    `json:"request_uuid"`
	RequestTime time.Time `json:"request_time"`
	UserIp      string    `json:"user_ip"`
	ApiServer   string    `json:"api_server"`
}

type CheckboxActionMessage struct {
	Header  MessageHeader         `json:"header"`
	Payload CheckboxActionPayload `json:"payload"`
}

type PublishMessageResult struct {
	MessageId      string    `json:"message_id"`
	SequenceNumber string    `json:"sequence_number"`
	PublishTime    time.Time `json:"publish_time"`
}

type QueueProvider interface {
	PublishCheckboxAction(ctx context.Context, message *CheckboxActionMessage) (PublishMessageResult, apierror.APIError)
	PullCheckboxActionMessages(ctx context.Context) ([]Message, apierror.APIError)
	DeleteMessage(ctx context.Context, message *Message) apierror.APIError
}

type Message struct {
	MessageId      string
	ReceiptHandle  string
	Body           string
	GroupId        string
	SequenceNumber string
	Attributes     map[string]string
}

var (
	providerInstance QueueProvider
	providerOnce     sync.Once
)

func (m *Message) UnmarshalBody(v interface{}) apierror.APIError {
	err := json.Unmarshal([]byte(m.Body), v)
	if err != nil {
		return apierror.WrapWithCodeFromConstants(err, apierror.ErrInternalServer, fmt.Sprintf("Could not unmarshal message json into type %T", v))
	}
	return nil
}

func getQueueProvider() QueueProvider {
	providerOnce.Do(func() {
		config := apiconfig.GetConfig()
		queueProvider := config.GetString("QUEUE_PROVIDER")

		switch queueProvider {
		case "aws":
			providerInstance = &awsQueueProvider{}
		default:
			// Default to AWS if not specified or invalid
			providerInstance = &awsQueueProvider{}
		}
	})
	return providerInstance
}

func PublishCheckboxAction(ctx context.Context, payload CheckboxActionPayload) (PublishMessageResult, apierror.APIError) {
	traceID := tracing.GetTraceIDFromContext(ctx)

	// Log the queue operation
	logging.LogQueueOperation(traceID, "publish_checkbox_action", map[string]interface{}{
		"action":       payload.Action,
		"checkbox_nbr": payload.CheckboxNbr,
		"user_uuid":    payload.UserUuid,
		"request_uuid": payload.RequestUuid,
		"user_ip":      payload.UserIp,
		"api_server":   payload.ApiServer,
		"trace_id":     traceID,
	})

	// Create the message with header
	message := &CheckboxActionMessage{
		Header: MessageHeader{
			PayloadSchemaVersion: "1.0",
			GroupId:              fmt.Sprintf("checkbox-%d", payload.CheckboxNbr),
			DeduplicationId:      payload.RequestUuid,
		},
		Payload: payload,
	}

	// Get the provider and publish the message
	provider := getQueueProvider()
	result, err := provider.PublishCheckboxAction(ctx, message)
	if err != nil {
		logging.LogQueueOperation(traceID, "publish_checkbox_action_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return PublishMessageResult{}, err
	}

	// Log successful publication
	logging.LogQueueOperation(traceID, "publish_checkbox_action_success", map[string]interface{}{
		"message_id":      result.MessageId,
		"sequence_number": result.SequenceNumber,
	})

	return result, nil
}

func PullCheckboxActionMessages(ctx context.Context) ([]Message, apierror.APIError) {
	logging.LogQueueOperation(tracing.GetTraceIDFromContext(ctx), "pull_checkbox_action_messages", nil)
	provider := getQueueProvider()
	messages, err := provider.PullCheckboxActionMessages(ctx)

	return messages, err
}

func DeleteMessage(ctx context.Context, message *Message) apierror.APIError {
	logging.LogQueueOperation(tracing.GetTraceIDFromContext(ctx), "delete_message", map[string]interface{}{
		"message_id":      message.MessageId,
		"sequence_number": message.SequenceNumber,
	})

	provider := getQueueProvider()
	return provider.DeleteMessage(ctx, message)
}
