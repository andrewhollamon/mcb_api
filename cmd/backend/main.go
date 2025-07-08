package main

import (
	"context"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/workers"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/workers/backend"
	"github.com/rs/zerolog/log"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	consumeCheckboxActionMinSleepTimeDuration = time.Duration(5) * time.Second
	sleepTimeMultiplier                       = 5 // wait time is 5x the runtime for automatic backoff
)

// launches the backend server, which publishes and consumes to the queue, and updates the postgresql db
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Info().Msgf("Received signal %v, shutting down gracefully...", sig)
		cancel()
	}()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Context cancelled, shutting down")
			return
		default:
			starttime := time.Now()
			result := backend.ConsumeCheckboxActionQueue(ctx)
			endtime := time.Now()
			runtimeSeconds := int(math.Round(endtime.Sub(starttime).Seconds()))

			log.Info().Msgf("Processed %d messages in %d seconds, errors: %t",
				result.NumProcessed,
				runtimeSeconds,
				result.Result == workers.ResultEnum.Failure)

			// wait time is sleepTimeMultiplier * runtime ... this provides a poor-man's automatic backoff if the processing slows down
			sleeptime := time.Duration(runtimeSeconds) * time.Second * sleepTimeMultiplier
			if sleeptime < consumeCheckboxActionMinSleepTimeDuration {
				sleeptime = consumeCheckboxActionMinSleepTimeDuration
			}

			// Context-aware sleep
			timer := time.NewTimer(sleeptime)
			select {
			case <-ctx.Done():
				timer.Stop()
				log.Info().Msg("Context cancelled during sleep, shutting down")
				return
			case <-timer.C:
			}
		}
	}
}
