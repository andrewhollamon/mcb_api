#!/bin/bash

# Database Setup Script for Million Checkboxes API
# This script:
# 1. Launches PostgreSQL in Docker
# 2. Creates database and roles
# 3. Runs migrations

set -e  # Exit on any error

# Configuration
CONTAINER_NAME="millcheck"
POSTGRES_PASSWORD="postgres"
POSTGRES_PORT="5432"
DB_NAME="millcheckdb"
DB_USER="mcbadminuser"
DB_PASSWORD="mcbadminuser"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if container is running
is_container_running() {
    docker ps --filter "name=$CONTAINER_NAME" --format "table {{.Names}}" | grep -q "$CONTAINER_NAME"
}

# Function to check if container exists (running or stopped)
container_exists() {
    docker ps -a --filter "name=$CONTAINER_NAME" --format "table {{.Names}}" | grep -q "$CONTAINER_NAME"
}

# Function to wait for PostgreSQL to be ready
wait_for_postgres() {
    log_info "Waiting for PostgreSQL to be ready..."
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if docker exec $CONTAINER_NAME pg_isready -U postgres -h localhost > /dev/null 2>&1; then
            log_info "PostgreSQL is ready!"
            return 0
        fi
        
        log_warn "Attempt $attempt/$max_attempts: PostgreSQL not ready yet, waiting..."
        sleep 2
        ((attempt++))
    done
    
    log_error "PostgreSQL failed to become ready after $max_attempts attempts"
    return 1
}

# Function to execute SQL file in container
execute_sql_file() {
    local sql_file="$1"
    local description="$2"
    
    log_info "Executing $description..."
    
    if docker exec -i $CONTAINER_NAME psql -U postgres -d postgres < "$sql_file"; then
        log_info "$description completed successfully"
    else
        log_error "$description failed"
        return 1
    fi
}

# Function to run migrations
run_migrations() {
    log_info "Running database migrations..."
    
    # Check if migrate tool is available
    if ! command -v migrate &> /dev/null; then
        log_error "migrate tool not found. Please install golang-migrate/migrate"
        log_error "Install with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
        return 1
    fi
    
    # Run migrations using the correct database name
    local db_url="postgres://$DB_USER:$DB_PASSWORD@localhost:$POSTGRES_PORT/$DB_NAME?sslmode=disable"
    
    if migrate -path database/migrations -database "$db_url" up; then
        log_info "Database migrations completed successfully"
    else
        log_error "Database migrations failed"
        return 1
    fi
}

# Main script execution
main() {
    log_info "Starting database setup..."
    
    # Step 1: Handle existing container
    if is_container_running; then
        log_info "Container $CONTAINER_NAME is already running"
    elif container_exists; then
        log_info "Container $CONTAINER_NAME exists but is stopped. Starting..."
        docker start $CONTAINER_NAME
    else
        log_info "Creating and starting new PostgreSQL container..."
        docker run --name $CONTAINER_NAME \
            -e POSTGRES_PASSWORD=$POSTGRES_PASSWORD \
            -d \
            -p $POSTGRES_PORT:5432 \
            postgres:17
    fi
    
    # Step 2: Wait for PostgreSQL to be ready
    if ! wait_for_postgres; then
        log_error "Failed to connect to PostgreSQL"
        exit 1
    fi
    
    # Step 3: Execute database and roles creation
    local sql_file="database/system/create_db_and_roles.sql"
    if [ ! -f "$sql_file" ]; then
        log_error "SQL file not found: $sql_file"
        exit 1
    fi
    
    if ! execute_sql_file "$sql_file" "database and roles setup"; then
        log_error "Database setup failed"
        exit 1
    fi
    
    # Step 4: Run migrations
    if ! run_migrations; then
        log_error "Migration setup failed"
        exit 1
    fi
    
    log_info "Database setup completed successfully!"
    log_info "Connection details:"
    log_info "  Host: localhost"
    log_info "  Port: $POSTGRES_PORT"
    log_info "  Database: $DB_NAME"
    log_info "  Username: $DB_USER"
}

# Check if script is being run from correct directory
if [ ! -f "database/system/create_db_and_roles.sql" ]; then
    log_error "This script must be run from the project root directory"
    log_error "Current directory: $(pwd)"
    exit 1
fi

# Run main function
main "$@"