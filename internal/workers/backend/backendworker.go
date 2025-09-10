package backend

import (
	"context"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/dbservice"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/queueservice"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/workers"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"runtime"
	"time"
)

func ConsumeCheckboxActionQueue(ctx context.Context) workers.QueueConsumerResult {
	startTime := time.Now()
	initialGoroutines := runtime.NumGoroutine()

	messages, err := queueservice.PullCheckboxActionMessages(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to pull messages from checkbox action queue")
		return workers.QueueConsumerResult{
			Result:       workers.ResultEnum.Failure,
			NumProcessed: 0,
		}
	}

	// if there are no messages, then we're done
	if len(messages) == 0 {
		return workers.QueueConsumerResult{
			Result:       workers.ResultEnum.Success,
			NumProcessed: 0,
		}
	}

	// This is a sanity check in case we change queue providers, to something that can return a very large number
	// of messages in one queue consume batch. The code below will spawn as many goroutines as there are messages
	// in this batch, so lets just put a guard here, just in case.
	if len(messages) > 100 {
		log.Error().Msgf("queue consumer received %d messages, this is too many", len(messages))
		return workers.QueueConsumerResult{
			Result:       workers.ResultEnum.Failure,
			NumProcessed: 0,
		}
	}

	result := workers.ResultEnum.Success
	processed := 0
	failed := 0
	messageCount := len(messages)
	c := make(chan workers.Result, messageCount)
	defer close(c)

	// kick off each received queue message on separate goroutine, since they're largely io bound
	// NOTE: This looks like it can spawn infinite goroutines, but it actually cannot, since the call to
	// queueservice.PullCheckboxActionMessages above can return a max of 10 messages at a time.
	for _, message := range messages {
		go func(msg queueservice.Message) {
			defer func() {
				if r := recover(); r != nil {
					log.Error().Msgf("panic in processCheckboxActionMessage: %v", r)
					c <- workers.ResultEnum.Failure
				}
			}()
			processCheckboxActionMessage(ctx, msg, c)
		}(message)
	}

	// process all the message results
	for i := 0; i < messageCount; i++ {
		innerresult := <-c
		if innerresult == workers.ResultEnum.Success {
			processed++
		} else {
			failed++
			result = workers.ResultEnum.Failure
		}
	}

	// Log metrics
	processingTime := time.Since(startTime)
	finalGoroutines := runtime.NumGoroutine()
	goroutinesDelta := finalGoroutines - initialGoroutines

	log.Info().Msgf("Queue processing metrics: processed=%d, failed=%d, duration=%v, goroutines_start=%d, goroutines_end=%d, goroutines_delta=%d",
		processed, failed, processingTime, initialGoroutines, finalGoroutines, goroutinesDelta)

	return workers.QueueConsumerResult{
		Result:       result,
		NumProcessed: processed,
	}
}

// TODO update to match WorkerProcessFunc signature in workerpool.go
func processCheckboxActionMessage(ctx context.Context, message queueservice.Message, c chan workers.Result) {
	// get the Body
	body := queueservice.CheckboxActionMessage{}
	err := message.UnmarshalBody(&body)
	if err != nil {
		log.Error().Err(err).Msg("failed to unmarshal message body")
		c <- workers.ResultEnum.Failure
		return
	}

	// unpack everything
	payload := body.Payload
	userUuid, baseerr := uuid.Parse(payload.UserUuid)
	if baseerr != nil {
		log.Error().Err(baseerr).Msgf("failed to parse user uuid '%s'", payload.UserUuid)
		c <- workers.ResultEnum.Failure
		return
	}
	requestUuid, baseerr := uuid.Parse(payload.RequestUuid)
	if baseerr != nil {
		log.Error().Err(baseerr).Msgf("failed to parse request uuid '%s'", payload.RequestUuid)
		c <- workers.ResultEnum.Failure
		return
	}

	// attempt to update the DB
	err = dbservice.UpdateCheckbox(
		ctx,
		payload.CheckboxNbr,
		payload.Action == queueservice.CheckboxActionChecked,
		userUuid,
		requestUuid)
	if err != nil {
		log.Error().Err(err).Msgf("failed to update checkbox %d for requestUuid %v", payload.CheckboxNbr, requestUuid)
		c <- workers.ResultEnum.Failure
		return
	}

	// remove it from the queue
	err = queueservice.DeleteMessage(ctx, &message)
	if err != nil {
		log.Error().Err(err).Msgf("failed to delete messageId %s sequenceNumber %s", message.MessageId, message.SequenceNumber)
		c <- workers.ResultEnum.Failure
		return
	}

	c <- workers.ResultEnum.Success
	return
}
