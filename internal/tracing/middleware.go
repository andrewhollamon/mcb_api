package tracing

import (
	"context"

	"github.com/andrewhollamon/millioncheckboxes-api/internal/uuidservice"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

const (
	TraceIDKey    = "trace_id"
	TraceIDHeader = "X-Trace-ID"
	RequestIDKey  = "request_id" // Alternative key name for compatibility
)

// RequestIDMiddleware generates and adds trace ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if trace ID already exists in headers (for distributed tracing)
		traceID := c.GetHeader(TraceIDHeader)

		// If no trace ID in headers, generate a new one
		if traceID == "" {
			uuid, err := uuidservice.NewRequestUuid()
			if err != nil {
				log.Error().
					Err(err).
					Str("path", c.Request.URL.Path).
					Str("method", c.Request.Method).
					Msg("Failed to generate trace ID, using fallback")

				panic("Failed to generate trace ID")
			} else {
				traceID = uuid.String()
			}
		}

		// Set trace ID in gin context
		c.Set(TraceIDKey, traceID)
		c.Set(RequestIDKey, traceID) // Set both keys for compatibility

		// Add trace ID to the Go context for downstream services
		ctx := context.WithValue(c.Request.Context(), TraceIDKey, traceID)
		c.Request = c.Request.WithContext(ctx)

		// Add trace ID to response headers for clients
		c.Header(TraceIDHeader, traceID)

		// Continue processing
		c.Next()
	}
}

// GetTraceID extracts trace ID from gin context
func GetTraceID(c *gin.Context) string {
	if traceID, exists := c.Get(TraceIDKey); exists {
		if traceStr, ok := traceID.(string); ok {
			return traceStr
		}
	}
	return ""
}

// GetTraceIDFromContext extracts trace ID from Go context
func GetTraceIDFromContext(ctx context.Context) string {
	if traceID := ctx.Value(TraceIDKey); traceID != nil {
		if traceStr, ok := traceID.(string); ok {
			return traceStr
		}
	}
	return ""
}

// WithTraceID adds trace ID to a Go context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// PropagateTraceID creates a new context with trace ID from gin context
func PropagateTraceID(c *gin.Context) context.Context {
	traceID := GetTraceID(c)
	if traceID == "" {
		return c.Request.Context()
	}
	return WithTraceID(context.Background(), traceID)
}

// LogWithTraceID creates a log event with trace ID from gin context
func LogWithTraceID(c *gin.Context) *gin.Context {
	traceID := GetTraceID(c)
	if traceID != "" {
		log.Info().Str(TraceIDKey, traceID)
	}
	return c
}

// Config holds configuration for tracing
type Config struct {
	Enabled             bool   `json:"enabled"`
	ServiceName         string `json:"service_name"`
	HeaderName          string `json:"header_name"`
	PropagateDownstream bool   `json:"propagate_downstream"`
}

// DefaultTracingConfig returns default tracing configuration
func DefaultTracingConfig() Config {
	return Config{
		Enabled:             true,
		ServiceName:         "mcb-api",
		HeaderName:          TraceIDHeader,
		PropagateDownstream: true,
	}
}

// ConfigurableRequestIDMiddleware creates middleware with custom configuration
func ConfigurableRequestIDMiddleware(config Config) gin.HandlerFunc {
	if !config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	headerName := config.HeaderName
	if headerName == "" {
		headerName = TraceIDHeader
	}

	return func(c *gin.Context) {
		// Check if trace ID already exists in headers
		traceID := c.GetHeader(headerName)

		// If no trace ID in headers, generate a new one
		if traceID == "" {
			uuid, err := uuidservice.NewRequestUuid()
			if err != nil {
				log.Error().
					Err(err).
					Str("service", config.ServiceName).
					Str("path", c.Request.URL.Path).
					Str("method", c.Request.Method).
					Msg("Failed to generate trace ID")
				panic("Failed to generate trace ID")
			} else {
				traceID = uuid.String()
			}
		}

		// Set trace ID in gin context
		c.Set(TraceIDKey, traceID)
		c.Set(RequestIDKey, traceID)

		// Add trace ID to the Go context if downstream propagation is enabled
		if config.PropagateDownstream {
			ctx := context.WithValue(c.Request.Context(), TraceIDKey, traceID)
			c.Request = c.Request.WithContext(ctx)
		}

		// Add trace ID to response headers for clients
		c.Header(headerName, traceID)

		// Continue processing
		c.Next()
	}
}

// TraceOperation logs the start and end of an operation with trace ID
func TraceOperation(c *gin.Context, operation string, fn func() error) error {
	traceID := GetTraceID(c)

	log.Debug().
		Str(TraceIDKey, traceID).
		Str("operation", operation).
		Msg("Operation started")

	err := fn()

	if err != nil {
		log.Error().
			Str(TraceIDKey, traceID).
			Str("operation", operation).
			Err(err).
			Msg("Operation failed")
	} else {
		log.Debug().
			Str(TraceIDKey, traceID).
			Str("operation", operation).
			Msg("Operation completed")
	}

	return err
}

// TraceOperationWithContext logs operation with explicit context
func TraceOperationWithContext(ctx context.Context, operation string, fn func() error) error {
	traceID := GetTraceIDFromContext(ctx)

	log.Debug().
		Str(TraceIDKey, traceID).
		Str("operation", operation).
		Msg("Operation started")

	err := fn()

	if err != nil {
		log.Error().
			Str(TraceIDKey, traceID).
			Str("operation", operation).
			Err(err).
			Msg("Operation failed")
	} else {
		log.Debug().
			Str(TraceIDKey, traceID).
			Str("operation", operation).
			Msg("Operation completed")
	}

	return err
}
