package queueservice

import (
	"context"
	"time"

	apierror "github.com/andrewhollamon/millioncheckboxes-api/internal/error"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/logging"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/tracing"
)

type MessageHeader struct {
	Topic                string `json:"topic"`
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

func PublishCheckboxActionMessageWithContext(ctx context.Context, payload CheckboxActionPayload) (PublishMessageResult, apierror.APIError) {
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

	// TODO: Implement actual queue message production
	/*
		1. Lookup from viper config which queue provider we're using
		2. Load the specific queue provider into a local variable
		3. Call the provider to publish the message, handle any errors
		4. Get the result, and log the result values and success/failure
	*/

	return PublishMessageResult{}, nil
}
