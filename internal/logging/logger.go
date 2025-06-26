package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/andrewhollamon/millioncheckboxes-api/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// LogConfig holds logging configuration
type LogConfig struct {
	Level      string   `json:"level"`       // trace, debug, info, warn, error, fatal, panic
	Format     string   `json:"format"`      // json, console
	Outputs    []string `json:"outputs"`     // stdout, stderr, file, cloudwatch, azure
	FilePath   string   `json:"file_path"`   // Path for file output
	MaxSize    int      `json:"max_size"`    // Max size in MB for log rotation
	MaxBackups int      `json:"max_backups"` // Max number of backup files
	MaxAge     int      `json:"max_age"`     // Max age in days
}

// DefaultConfig returns default logging configuration
func DefaultConfig() LogConfig {
	return LogConfig{
		Level:      "info",
		Format:     "json",
		Outputs:    []string{"stdout"},
		FilePath:   "/var/log/mcb-api.log",
		MaxSize:    100, // 100 MB
		MaxBackups: 3,
		MaxAge:     28, // 28 days
	}
}

// InitLogger initializes the global logger with the given configuration
func InitLogger(config LogConfig) error {
	// Set log level
	level, err := zerolog.ParseLevel(config.Level)
	if err != nil {
		return fmt.Errorf("invalid log level '%s': %w", config.Level, err)
	}
	zerolog.SetGlobalLevel(level)

	// Create writers based on outputs
	var writers []io.Writer
	
	for _, output := range config.Outputs {
		switch strings.ToLower(output) {
		case "stdout":
			if config.Format == "console" {
				writers = append(writers, zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
			} else {
				writers = append(writers, os.Stdout)
			}
		case "stderr":
			if config.Format == "console" {
				writers = append(writers, zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
			} else {
				writers = append(writers, os.Stderr)
			}
		case "file":
			fileWriter, err := createFileWriter(config.FilePath)
			if err != nil {
				return fmt.Errorf("failed to create file writer: %w", err)
			}
			writers = append(writers, fileWriter)
		case "cloudwatch":
			// TODO: Implement CloudWatch writer
			// For now, fall back to file
			fileWriter, err := createFileWriter(config.FilePath)
			if err != nil {
				return fmt.Errorf("failed to create cloudwatch fallback file writer: %w", err)
			}
			writers = append(writers, fileWriter)
		case "azure":
			// TODO: Implement Azure Monitor writer
			// For now, fall back to file
			fileWriter, err := createFileWriter(config.FilePath)
			if err != nil {
				return fmt.Errorf("failed to create azure fallback file writer: %w", err)
			}
			writers = append(writers, fileWriter)
		default:
			return fmt.Errorf("unsupported log output: %s", output)
		}
	}

	// Create multi-writer if multiple outputs
	var writer io.Writer
	if len(writers) == 1 {
		writer = writers[0]
	} else {
		writer = zerolog.MultiLevelWriter(writers...)
	}

	// Create and set global logger
	logger := zerolog.New(writer).With().
		Timestamp().
		Caller().
		Str("service", "mcb-api").
		Logger()

	log.Logger = logger

	return nil
}

// InitLoggerFromEnv initializes logger from environment variables
func InitLoggerFromEnv() error {
	logConfig := LogConfig{
		Level:      config.GetStringWithDefault("LOG_LEVEL", "info"),
		Format:     config.GetStringWithDefault("LOG_FORMAT", "json"),
		Outputs:    strings.Split(config.GetStringWithDefault("LOG_OUTPUT", "stdout"), ","),
		FilePath:   config.GetStringWithDefault("LOG_FILE_PATH", "/var/log/mcb-api.log"),
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     28,
	}

	// Clean up outputs (remove whitespace)
	for i, output := range logConfig.Outputs {
		logConfig.Outputs[i] = strings.TrimSpace(output)
	}

	return InitLogger(logConfig)
}

// createFileWriter creates a file writer with log rotation
func createFileWriter(filePath string) (io.Writer, error) {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// For now, just use a simple file writer
	// In a production system, you'd want to use a rotating file writer
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return file, nil
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	return config.GetStringWithDefault(key, defaultValue)
}

// WithTraceID adds trace ID to a log event
func WithTraceID(traceID string) *zerolog.Event {
	return log.Info().Str("trace_id", traceID)
}

// WithError adds error information to a log event
func WithError(err error) *zerolog.Event {
	return log.Error().Err(err)
}

// WithFields adds multiple fields to a log event
func WithFields(fields map[string]interface{}) *zerolog.Event {
	event := log.Info()
	for key, value := range fields {
		event = event.Interface(key, value)
	}
	return event
}

// LogRequest logs HTTP request information
func LogRequest(method, path, userAgent, clientIP, traceID string, duration time.Duration, statusCode int) {
	log.Info().
		Str("trace_id", traceID).
		Str("method", method).
		Str("path", path).
		Str("user_agent", userAgent).
		Str("client_ip", clientIP).
		Dur("duration", duration).
		Int("status_code", statusCode).
		Msg("HTTP request completed")
}

// LogError logs error with context
func LogError(err error, traceID, message string, fields map[string]interface{}) {
	event := log.Error().
		Err(err).
		Str("trace_id", traceID).
		Str("message", message)
	
	for key, value := range fields {
		event = event.Interface(key, value)
	}
	
	event.Send()
}

// LogInfo logs info message with context
func LogInfo(traceID, message string, fields map[string]interface{}) {
	event := log.Info().
		Str("trace_id", traceID).
		Str("message", message)
	
	for key, value := range fields {
		event = event.Interface(key, value)
	}
	
	event.Send()
}

// LogDebug logs debug message with context
func LogDebug(traceID, message string, fields map[string]interface{}) {
	event := log.Debug().
		Str("trace_id", traceID).
		Str("message", message)
	
	for key, value := range fields {
		event = event.Interface(key, value)
	}
	
	event.Send()
}