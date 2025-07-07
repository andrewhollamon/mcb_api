package main

import (
	"context"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/api"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/config"
	"github.com/andrewhollamon/millioncheckboxes-api/internal/dbservice"
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

	// Dump configuration for debugging
	//config.DumpConfig()

	// Initialize logging system from environment variables
	err = logging.InitLoggerFromEnv()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize logging system")
	}

	log.Info().Msg("Starting MCB API Server")

	log.Info().Msg("Initializing database connection pool")
	apierr := dbservice.InitDbPool(context.Background())
	if apierr != nil {
		log.Fatal().Err(apierr).Msg("Failed to initialize database connection pool")
		panic("Failed to initialize database connection pool")
	}
	defer dbservice.ClosePool()
	log.Info().Msg("Database connection pool initialized")

	// Setup the memory store
	log.Info().Msg("Initializing memory store")
	memorystore.Init()

	// Setup router with middleware
	log.Info().Msg("Setting up HTTP router")
	r := api.SetupRouter()

	// Start server
	port := config.GetString("apiserver_port")
	log.Info().Str("port", port).Msg("Starting HTTP server")
	err = r.Run(":" + port)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start HTTP server")
	}

	// test whether the queue is reachable, and stop accepting changes if not
	// TODO: Add queue health check

}
