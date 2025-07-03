package queueservice

import (
	"context"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/uuidservice"
	"testing"
	"time"

	apiconfig "github.com/andrewhollamon/millioncheckboxes-api/internal/config"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/queueservice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TestUserUuid        = "550e8400-e29b-41d4-a716-446655440000"
	TestRequestUuid     = "550e8400-e29b-41d4-a716-446655440001"
	TestUserIp          = "127.0.0.1"
	TestApiServer       = "api1"
	TestCheckboxNbr     = 1
	TestActionChecked   = "checked"
	TestActionUnchecked = "unchecked"
)

func TestPublishReceiveDeleteCheckedAction(t *testing.T) {
	// Create the test payload with specified values
	requestTime := time.Now()
	requestUuid, err := uuidservice.NewRequestUuid()
	assert.NoError(t, err, "Failed to generate request UUID")

	payload := queueservice.CheckboxActionPayload{
		Action:      TestActionChecked,
		CheckboxNbr: TestCheckboxNbr,
		UserUuid:    TestUserUuid,
		RequestUuid: requestUuid.String(),
		RequestTime: requestTime,
		UserIp:      TestUserIp,
		ApiServer:   TestApiServer,
	}

	testPublishCheckboxCheckedAction(t, payload)
	testPullCheckboxActionMessages(t, payload)
}

// TestPublishCheckboxCheckedAction tests the PublishCheckboxAction function
// This is an integration test that will hit the real AWS SNS topic
func testPublishCheckboxCheckedAction(t *testing.T, payload queueservice.CheckboxActionPayload) {
	// Initialize configuration to load local.env file
	err := apiconfig.InitConfigWithFolder("../../../config/", "")
	require.NoError(t, err, "Failed to initialize config")

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

func testPullCheckboxActionMessages(t *testing.T, payload queueservice.CheckboxActionPayload) {
	// Initialize configuration to load local.env file
	err := apiconfig.InitConfigWithFolder("../../../config/", "")
	require.NoError(t, err, "Failed to initialize config")

	ctx := context.Background()
	messages, apiErr := queueservice.PullCheckboxActionMessages(ctx)

	// Assert that no error occurred
	assert.Nil(t, apiErr, "PullCheckboxActionMessages should not return an error")
	assert.True(t, len(messages) > 0, "PullCheckboxActionMessages should return at least one message")

	for i, m := range messages {
		t.Logf("Message #%d: %+v", i, m)

		body := queueservice.CheckboxActionMessage{}
		err := m.UnmarshalBody(&body)
		assert.NoErrorf(t, err, "UnmarshalBody should not return an error for message %d", i)
		assert.NotEmpty(t, body.Payload.Action, "Action should not be empty for message %d", i)

		matches := true
		matches = matches && assert.Equal(t, payload.Action, body.Payload.Action)
		matches = matches && assert.Equal(t, payload.CheckboxNbr, body.Payload.CheckboxNbr)
		matches = matches && assert.Equal(t, payload.UserUuid, body.Payload.UserUuid)
		matches = matches && assert.Equal(t, payload.RequestUuid, body.Payload.RequestUuid)
		matches = matches && assert.Equal(t, payload.UserIp, body.Payload.UserIp)
		matches = matches && assert.Equal(t, payload.ApiServer, body.Payload.ApiServer)
		assert.True(t, matches, "Message %d should match expected values", i)

		err = queueservice.DeleteMessage(ctx, &m)
		assert.NoErrorf(t, err, "DeleteMessage should not return an error for message %d", i)
	}
}
