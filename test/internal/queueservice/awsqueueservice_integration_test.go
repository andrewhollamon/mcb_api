package queueservice

import (
	"context"
	"testing"
	"time"

	apiconfig "github.com/andrewhollamon/millioncheckboxes-api/internal/config"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/queueservice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPublishCheckboxActionIntegration tests the PublishCheckboxAction function
// This is an integration test that will hit the real AWS SNS topic
func TestPublishCheckboxActionIntegration(t *testing.T) {
	// Initialize configuration to load local.env file
	err := apiconfig.InitConfigWithFolder("../../../config/", "")
	require.NoError(t, err, "Failed to initialize config")

	// Create the test payload with specified values
	requestTime := time.Now()
	payload := queueservice.CheckboxActionPayload{
		Action:      "checked",
		CheckboxNbr: 1,
		UserUuid:    "550e8400-e29b-41d4-a716-446655440000",
		RequestUuid: "550e8400-e29b-41d4-a716-446655440001",
		RequestTime: requestTime,
		UserIp:      "127.0.0.1",
		ApiServer:   "api1",
	}

	// Create context for the test
	ctx := context.Background()

	// Call the PublishCheckboxAction function
	result, apiErr := queueservice.PublishCheckboxAction(ctx, payload)

	// Assert that no error occurred
	assert.Nil(t, apiErr, "PublishCheckboxAction should not return an error")

	// Assert that we got a valid result
	assert.NotEmpty(t, result.MessageId, "MessageId should not be empty")
	assert.NotEmpty(t, result.SequenceNumber, "SequenceNumber should not be empty")
	assert.False(t, result.PublishTime.IsZero(), "PublishTime should be set")

	// Log the results for verification
	t.Logf("Message published successfully:")
	t.Logf("  MessageId: %s", result.MessageId)
	t.Logf("  SequenceNumber: %s", result.SequenceNumber)
	t.Logf("  PublishTime: %s", result.PublishTime.Format(time.RFC3339))
	t.Logf("  Payload Action: %s", payload.Action)
	t.Logf("  Payload CheckboxNbr: %d", payload.CheckboxNbr)
	t.Logf("  Payload UserUuid: %s", payload.UserUuid)
	t.Logf("  Payload RequestUuid: %s", payload.RequestUuid)
	t.Logf("  Payload UserIp: %s", payload.UserIp)
	t.Logf("  Payload ApiServer: %s", payload.ApiServer)
}
