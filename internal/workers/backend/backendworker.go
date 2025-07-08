package backend

import (
	"context"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/dbservice"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/queueservice"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/workers"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func ConsumeCheckboxActionQueue(ctx context.Context) workers.QueueConsumerResult {
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

	result := workers.ResultEnum.Success
	processed := 0
	messageCount := len(messages)
	c := make(chan workers.Result, messageCount)

	// kick off each received queue message on separate goroutine, since they're largely io bound
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
			result = workers.ResultEnum.Failure
		}
	}

	return workers.QueueConsumerResult{
		Result:       result,
		NumProcessed: processed,
	}
}

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
