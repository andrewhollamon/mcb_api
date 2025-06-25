# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Development Commands

### Building and Running
- Build the API server: `go build -o bin/api-server ./cmd/api-server`
- Run the API server: `go run ./cmd/api-server/main.go` (runs on port 8080)
- Test the server: `curl http://localhost:8080/ping` (should return "pong")

### Database Setup
- Launch local PostgreSQL docker container: `./database/system/docker_launch_db`
- Setup database and roles: Run `database/system/create_db_and_roles.sql`
- Run migrations: `migrate -source database/migrations -database postgres://localhost:5432/database up`
  - Migration uses the `golang-migrate/migrate` tool, which you can assume is installed

### Testing
- Run tests: `go test ./...`
- Run specific package tests: `go test ./internal/memorystore`

## Architecture Overview

This is a "Million Checkboxes" API backend implementing a multi-layered architecture:

All backend services are written in golang 1.24.

### Core Components
- **API Servers**: Golang and Gin-based HTTP server (`cmd/api-server/main.go`) with endpoints for checkbox operations
  - Includes the web page that serves new connections and the client UI
  - Includes the web page that shows status and metrics about the system
  - Includes an in-memory store of the checkboxes, that is eventually consistent with the database
  - Includes a Queue publisher, that publishes checkbox action messages, once validated
  - Includes a Queue consumer, which reads fully persisted checkbox actions, and updates the in-memory store
- **Persistence Server**: Golang server (no api or webpage) that does the following:
  - Consumes queue messages about checkbox actions and persist them to the postgresql db
  - Publishes queue messages about persisted checkbox actions so that consumers can be eventually consistent
  - Log all action attempts, successes and failures in a durable database log 
  - Monitor postgresql database, queue system, and running api servers for performance and throughput metrics

### Database Design
PostgreSQL with partitioned tables:
- `CLIENT_T`: Frontend client tracking with UUIDs
- `CHECKBOX_T`: Million checkboxes with state (partitioned by checkbox number)
- `CHECKBOX_DETAILS_T`: Checkbox metadata and last update info
- `UPDATE_T`: Audit trail of all checkbox changes
- `METRICS_T`: System metrics tracking

### API Endpoints
- `GET /ping`: Health check
- `GET /api/v1/currentstate`: Get the current state of the entire million checkboxes
- `GET /api/v1/checkbox/{checkboxNbr}/status`: Get the persisted state of a specific checkbox
- `POST /api/v1/checkbox/{checkboxNbr}/check/{userUuid}`: Check a checkbox
- `POST /api/v1/checkbox/{checkboxNbr}/uncheck/{userUuid}`: Uncheck a checkbox
- `GET /api/v1/changes/{versionNumber}`: retrieve all persisted checkbox actions since versionNumber

### Key Design Patterns
- API Servers keep an eventually consistent in-memory store to act as a cache against the postgresql db
- Database partitioning planned for checkbox table scalability
- One worker thread per db partition reading from the queue and persisting
  - This ensures that there is only a single-threaded writer per db partition
- Web clients send a checkbox action when the user checks or unchecks a checkbox, and assumes locally its true
  - Log console shows when the API server reports back that the action request was persisted 
- Web clients pull down the initial state of the checkboxes as a binary blob from the API server, then periodically poll for changes since versionNumber 

### Development Notes
- Memory store supports 1M checkboxes (0-999999)
- Database migrations are numbered sequentially (005, 010, 015, etc.)
- API Server runs on port 8080 by default
- Trust proxy is disabled for security

## Folder Structure

- `cmd/api-server`: The api and webserver
- `cmd/backend`: The backend server which publishes and consumes to the queue, updates the postgresql database, and maintains system metrics
- `config`: Configuration files for non-prod, production will use env variables, using 12 Factor App approach
- `database`: The postgresql database setup, including docker commands, db/roles setup, and `golang-migrate/migrate` based migrations
- `docs`: Documentation for the system, no code
- `internal`: The bulk of the code for the system, all written in golang
  - `internal/api`: API server (uses gin framework)
  - `internal/error`: Error types and utility code
  - `internal/memorystore`: In-memory store of the state of checkboxes, eventually consistent from the postgresql, which is the canonical source of truth
  - `internal/queueservice`: Queue services. Exported service functions should be stable across different queue implementations (azure, aws, kafka, etc), and the implementations hidden
  - `internal/uuidservice`: This system uses alot of UUIDs, mostly v7 so they sort well in the DB
- `pkg`: Exported code, if we need any
- `test`: Unit and integration tests
