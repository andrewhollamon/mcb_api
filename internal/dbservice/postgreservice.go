package dbservice

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	apiconfig "github.com/andrewhollamon/millioncheckboxes-api/internal/config"
)

var pool *pgxpool.Pool

// InitializePool creates and configures the PostgreSQL connection pool
func InitializePool(ctx context.Context) error {
	if pool != nil {
		return nil // Already initialized
	}

	appconfig := apiconfig.GetConfig()
	dburl := appconfig.GetString("DATABASE_URL")
	dbuser := appconfig.GetString("DATABASE_USER")
	dbpassword := appconfig.GetString("DATABASE_PASSWORD")

	if dburl == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if dbuser == "" {
		return fmt.Errorf("DATABASE_USER is required")
	}
	if dbpassword == "" {
		return fmt.Errorf("DATABASE_PASSWORD is required")
	}

	// Build connection string with credentials
	connStr := fmt.Sprintf("%s?user=%s&password=%s", dburl, dbuser, dbpassword)

	// Configure connection pool
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return fmt.Errorf("failed to parse database config: %w", err)
	}

	// Set pool configuration with static defaults
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = 30 * time.Second

	// Create the pool
	pool, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info().
		Int32("max_conns", config.MaxConns).
		Int32("min_conns", config.MinConns).
		Dur("max_conn_lifetime", config.MaxConnLifetime).
		Dur("max_conn_idle_time", config.MaxConnIdleTime).
		Msg("PostgreSQL connection pool initialized successfully")

	return nil
}

// Query executes a parameterized query that returns zero to many rows
func Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	if pool == nil {
		return nil, fmt.Errorf("database pool not initialized")
	}

	log.Debug().
		Str("query", query).
		Interface("args", args).
		Msg("Executing query")

	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		log.Error().
			Err(err).
			Str("query", query).
			Interface("args", args).
			Msg("Query execution failed")
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	return rows, nil
}

// Exec executes a parameterized query that returns zero or one return value
// and returns the number of affected rows
func Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	if pool == nil {
		return pgconn.CommandTag{}, fmt.Errorf("database pool not initialized")
	}

	log.Debug().
		Str("query", query).
		Interface("args", args).
		Msg("Executing command")

	tag, err := pool.Exec(ctx, query, args...)
	if err != nil {
		log.Error().
			Err(err).
			Str("query", query).
			Interface("args", args).
			Msg("Command execution failed")
		return pgconn.CommandTag{}, fmt.Errorf("command execution failed: %w", err)
	}

	log.Debug().
		Str("query", query).
		Int64("rows_affected", tag.RowsAffected()).
		Msg("Command executed successfully")

	return tag, nil
}

// ClosePool closes the connection pool
func ClosePool() {
	if pool != nil {
		pool.Close()
		pool = nil
		log.Info().Msg("PostgreSQL connection pool closed")
	}
}

// GetPoolStats returns connection pool statistics
func GetPoolStats() *pgxpool.Stat {
	if pool == nil {
		return nil
	}
	return pool.Stat()
}

// BeginTx starts a new database transaction
func BeginTx(ctx context.Context) (pgx.Tx, error) {
	if pool == nil {
		return nil, fmt.Errorf("database pool not initialized")
	}

	log.Debug().Msg("Beginning database transaction")

	tx, err := pool.Begin(ctx)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to begin transaction")
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	log.Debug().Msg("Transaction started successfully")
	return tx, nil
}

// QueryTx executes a parameterized query within a transaction that returns zero to many rows
func QueryTx(ctx context.Context, tx pgx.Tx, query string, args ...interface{}) (pgx.Rows, error) {
	if tx == nil {
		return nil, fmt.Errorf("transaction is nil")
	}

	log.Debug().
		Str("query", query).
		Interface("args", args).
		Msg("Executing query in transaction")

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		log.Error().
			Err(err).
			Str("query", query).
			Interface("args", args).
			Msg("Query execution failed in transaction")
		return nil, fmt.Errorf("query execution failed in transaction: %w", err)
	}

	return rows, nil
}

// ExecTx executes a parameterized command within a transaction that returns zero or one return value
// and returns the number of affected rows
func ExecTx(ctx context.Context, tx pgx.Tx, query string, args ...interface{}) (pgconn.CommandTag, error) {
	if tx == nil {
		return pgconn.CommandTag{}, fmt.Errorf("transaction is nil")
	}

	log.Debug().
		Str("query", query).
		Interface("args", args).
		Msg("Executing command in transaction")

	tag, err := tx.Exec(ctx, query, args...)
	if err != nil {
		log.Error().
			Err(err).
			Str("query", query).
			Interface("args", args).
			Msg("Command execution failed in transaction")
		return pgconn.CommandTag{}, fmt.Errorf("command execution failed in transaction: %w", err)
	}

	log.Debug().
		Str("query", query).
		Int64("rows_affected", tag.RowsAffected()).
		Msg("Command executed successfully in transaction")

	return tag, nil
}

// CommitTx commits a database transaction
func CommitTx(ctx context.Context, tx pgx.Tx) error {
	if tx == nil {
		return fmt.Errorf("transaction is nil")
	}

	log.Debug().Msg("Committing transaction")

	err := tx.Commit(ctx)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to commit transaction")
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Debug().Msg("Transaction committed successfully")
	return nil
}

// RollbackTx rolls back a database transaction
func RollbackTx(ctx context.Context, tx pgx.Tx) error {
	if tx == nil {
		return fmt.Errorf("transaction is nil")
	}

	log.Debug().Msg("Rolling back transaction")

	err := tx.Rollback(ctx)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to rollback transaction")
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	log.Debug().Msg("Transaction rolled back successfully")
	return nil
}
