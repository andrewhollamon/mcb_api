# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Development Commands

### Building and Running
- Build the API server: `go build -o bin/api-server ./cmd/api-server`
- Run the API server: `go run ./cmd/api-server/main.go` (runs on port 8080)
- Test the server: `curl http://localhost:8080/ping` (should return "pong")

### Database Setup
- Launch local PostgreSQL container: `./database/system/docker_launch_db`
- Run migrations: Apply SQL files in `database/migrations/` in numerical order (005, 010, 015, etc.)
- Setup database and roles: Run `database/system/create_db_and_roles.sql`

### Testing
- Run tests: `go test ./...`
- Run specific package tests: `go test ./internal/memorystore`

## Architecture Overview

This is a "Million Checkboxes" API backend implementing a multi-layered architecture:

### Core Components
- **API Layer**: Gin-based HTTP server (`cmd/api-server/main.go`) with endpoints for checkbox operations
- **Memory Store**: In-memory array of 1M booleans (`internal/memorystore`) for fast checkbox state access
- **UUID Service**: UUIDv7 generation for clients and requests (`internal/uuidservice`)
- **Error Handling**: Unified error system (`pkg/errors`) with 4xx/5xx distinction, trace IDs, and context attachment
- **Queue Service**: Planned async processing layer (currently placeholder in `internal/queueservice`)

### Database Design
PostgreSQL with partitioned tables:
- `CLIENT_T`: Frontend client tracking with UUIDs
- `CHECKBOX_T`: Million checkboxes with state (partitioned by checkbox number)
- `CHECKBOX_DETAILS_T`: Checkbox metadata and last update info
- `UPDATE_T`: Audit trail of all checkbox changes
- `METRICS_T`: System metrics tracking

### API Endpoints
- `GET /ping`: Health check
- `GET /api/v1/checkbox/{checkboxNbr}/status`: Get checkbox state
- `POST /api/v1/checkbox/{checkboxNbr}/check/{userUuid}`: Check a checkbox
- `POST /api/v1/checkbox/{checkboxNbr}/uncheck/{userUuid}`: Uncheck a checkbox

### Key Design Patterns
- Memory store acts as fast cache layer before database persistence
- UUIDv7 used for time-ordered identifiers
- Database partitioning planned for checkbox table scalability
- Async queue processing planned between API and database layers

### Error Handling System
- Use `pkg/errors` for all API errors instead of standard Go errors
- All errors include trace IDs, timestamps, and contextual information
- 4xx errors: ValidationError for client mistakes (bad params, auth, etc.)
- 5xx errors: InternalError for server issues (database, queue failures, etc.)
- Middleware handles panics and provides trace ID correlation
- Helper functions: `ValidateParams()`, `InvalidCheckboxNumber()`, etc.

### Development Notes
- Memory store supports 1M checkboxes (0-999999)
- Database migrations are numbered sequentially (005, 010, 015, etc.)
- Server runs on port 8080 by default
- Trust proxy is disabled for security
- Always use the unified error system for consistent API responses