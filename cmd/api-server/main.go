package main

import (
	"github.com/andrewhollamon/millioncheckboxes-api/internal/api"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/config"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/logging"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/memorystore"
	"github.com/rs/zerolog/log"
)

func main() {
	// Initialize configuration
	err := config.InitConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize configuration")
	}

	// Initialize logging system from environment variables
	err = logging.InitLoggerFromEnv()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize logging system")
	}

	log.Info().Msg("Starting MCB API Server")

	// Setup the memory store
	log.Info().Msg("Initializing memory store")
	memorystore.Init()

	// Setup router with middleware
	log.Info().Msg("Setting up HTTP router")
	r := api.SetupRouter()

	// Start server
	log.Info().Str("port", "8080").Msg("Starting HTTP server")
	err = r.Run(":8080")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start HTTP server")
	}

	// test whether the queue is reachable, and stop accepting changes if not
	// TODO: Add queue health check

}
