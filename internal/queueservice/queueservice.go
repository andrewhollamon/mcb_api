package queueservice

import (
	error0 "github.com/andrewhollamon/millioncheckboxes-api/internal/error"
	"time"
)

// not sure the best way to setup this struct to handle structural changes in future version
// maybe postfix the name with 10 for 1.0?

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
	Hostname    string    `json:"hostname"`
	Hostip      string    `json:"hostip"`
}

type CheckboxActionMessage struct {
	Header  MessageHeader         `json:"header"`
	Payload CheckboxActionPayload `json:"payload"`
}

func ProduceCheckboxActionMessage(payload CheckboxActionPayload) error0.APIError {
	return nil
}
