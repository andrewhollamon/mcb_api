package logging

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// RequestLoggingMiddleware logs HTTP requests with detailed information
func RequestLoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: customLogFormatter,
		Output:    &logWriter{},
		SkipPaths: []string{"/ping"}, // Skip health check endpoints
	})
}

// customLogFormatter formats log entries for HTTP requests
func customLogFormatter(param gin.LogFormatterParams) string {
	// Get trace ID from context
	traceID := ""
	if param.Keys != nil {
		if id, exists := param.Keys["trace_id"]; exists {
			if idStr, ok := id.(string); ok {
				traceID = idStr
			}
		}
	}

	// Log using zerolog
	log.Info().
		Str("trace_id", traceID).
		Str("method", param.Method).
		Str("path", param.Path).
		Int("status_code", param.StatusCode).
		Dur("latency", param.Latency).
		Str("client_ip", param.ClientIP).
		Str("user_agent", param.Request.UserAgent()).
		Int("body_size", param.BodySize).
		Msg("HTTP request completed")

	// Return empty string since we handle logging through zerolog
	return ""
}

// logWriter implements io.Writer but doesn't actually write
// We use this to satisfy gin.LoggerConfig but handle logging through zerolog
type logWriter struct{}

func (w *logWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// RequestTimingMiddleware adds request timing to context
func RequestTimingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Set("request_start_time", start)
		c.Next()
	}
}

// DetailedRequestLoggingMiddleware provides more detailed request logging
func DetailedRequestLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Get trace ID
		traceID := ""
		if id, exists := c.Get("trace_id"); exists {
			if idStr, ok := id.(string); ok {
				traceID = idStr
			}
		}

		// Log request start
		log.Debug().
			Str("trace_id", traceID).
			Str("method", c.Request.Method).
			Str("path", path).
			Str("query", raw).
			Str("client_ip", c.ClientIP()).
			Str("user_agent", c.Request.UserAgent()).
			Msg("HTTP request started")

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Log request completion
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		logLevel := log.Info()
		if statusCode >= 400 {
			logLevel = log.Error()
		} else if statusCode >= 300 {
			logLevel = log.Warn()
		}

		logLevel.
			Str("trace_id", traceID).
			Str("method", c.Request.Method).
			Str("path", path).
			Str("query", raw).
			Int("status_code", statusCode).
			Dur("latency", latency).
			Str("client_ip", c.ClientIP()).
			Str("user_agent", c.Request.UserAgent()).
			Int("body_size", bodySize).
			Msg("HTTP request completed")

		// Log errors if any
		if len(c.Errors) > 0 {
			for _, ginErr := range c.Errors {
				log.Error().
					Str("trace_id", traceID).
					Str("method", c.Request.Method).
					Str("path", path).
					Err(ginErr.Err).
					//Str("error_type", ginErr.Type).
					Msg("Request error occurred")
			}
		}
	}
}

// LogAPICall logs API calls with parameters
func LogAPICall(c *gin.Context, operation string, params map[string]interface{}) {
	traceID := ""
	if id, exists := c.Get("trace_id"); exists {
		if idStr, ok := id.(string); ok {
			traceID = idStr
		}
	}

	event := log.Info().
		Str("trace_id", traceID).
		Str("operation", operation).
		Str("method", c.Request.Method).
		Str("path", c.Request.URL.Path)

	for key, value := range params {
		event = event.Interface(key, value)
	}

	event.Msg("API call initiated")
}

// LogAPIResponse logs API response
func LogAPIResponse(c *gin.Context, operation string, statusCode int, responseData interface{}) {
	traceID := ""
	if id, exists := c.Get("trace_id"); exists {
		if idStr, ok := id.(string); ok {
			traceID = idStr
		}
	}

	log.Info().
		Str("trace_id", traceID).
		Str("operation", operation).
		Int("status_code", statusCode).
		Interface("response_data", responseData).
		Msg("API call completed")
}

// LogQueueOperation logs queue operations
func LogQueueOperation(traceID, operation string, params map[string]interface{}) {
	event := log.Info().
		Str("trace_id", traceID).
		Str("operation", operation).
		Str("component", "queue")

	for key, value := range params {
		event = event.Interface(key, value)
	}

	event.Msg("Queue operation")
}

// LogDatabaseOperation logs database operations
func LogDatabaseOperation(traceID, operation string, params map[string]interface{}) {
	event := log.Info().
		Str("trace_id", traceID).
		Str("operation", operation).
		Str("component", "database")

	for key, value := range params {
		event = event.Interface(key, value)
	}

	event.Msg("Database operation")
}
