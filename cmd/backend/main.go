package main

import (
	"context"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/workers"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/workers/backend"
	"github.com/rs/zerolog/log"
	"math"
	"time"
)

const (
	consumeCheckboxActionMinSleepTimeDuration = time.Duration(5) * time.Second
)

// launches the backend server, which publishes and consumes to the queue, and updates the postgresql db
func main() {

	ctx := context.Background()

	for {
		starttime := time.Now()
		result := backend.ConsumeCheckboxActionQueue(ctx)
		endtime := time.Now()
		runtimeSeconds := int(math.Round(endtime.Sub(starttime).Seconds()))

		log.Info().Msgf("Processed %d messages in %d seconds, errors: %t",
			result.NumProcessed,
			runtimeSeconds,
			result.Result == workers.ResultEnum.Failure)

		// wait time is 5x the runtime ... this provides a poor-man's automatic backoff if the processing slows down
		sleeptime := time.Duration(runtimeSeconds) * time.Second * 5
		if sleeptime < consumeCheckboxActionMinSleepTimeDuration {
			sleeptime = consumeCheckboxActionMinSleepTimeDuration
		}
		time.Sleep(sleeptime)
	}

}
