package error

import (
	"net/http"
	"os"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// ErrorHandlingMiddleware handles panics and APIErrors in Gin handlers
func ErrorHandlingMiddleware() gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(gin.DefaultWriter, func(c *gin.Context, recovered interface{}) {
		// Handle panics
		if recovered != nil {
			handlePanic(c, recovered)
			return
		}

		// Handle APIErrors that were set in the context
		if len(c.Errors) > 0 {
			handleAPIErrors(c)
			return
		}

		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

// handlePanic processes panic recovery
func handlePanic(c *gin.Context, recovered interface{}) {
	traceID := getTraceID(c)

	// Log the panic with stack trace
	log.Error().
		Str("trace_id", traceID).
		Str("method", c.Request.Method).
		Str("path", c.Request.URL.Path).
		Str("user_agent", c.Request.UserAgent()).
		Str("client_ip", c.ClientIP()).
		Interface("panic", recovered).
		Bytes("stack", debug.Stack()).
		Msg("Panic recovered in HTTP handler")

	// Create APIError for panic
	apiErr := NewAPIErrorFromCode(ErrInternalServer, "Internal server error occurred")
	if traceID != "" {
		apiErr = apiErr.WithContext(c.Request.Context())
	}

	// Send error response
	sendErrorResponse(c, apiErr)
}

// handleAPIErrors processes APIErrors stored in gin.Context
func handleAPIErrors(c *gin.Context) {
	// Get the last error (most recent)
	lastError := c.Errors.Last()
	if lastError == nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Check if it's an APIError
	if apiErr, ok := lastError.Err.(APIError); ok {
		sendErrorResponse(c, apiErr)
		return
	}

	// Convert regular error to APIError
	apiErr := InternalError("Internal server error")
	if traceID := getTraceID(c); traceID != "" {
		apiErr = apiErr.WithContext(c.Request.Context())
	}

	sendErrorResponse(c, apiErr)
}

// sendErrorResponse sends a standardized error response
func sendErrorResponse(c *gin.Context, apiErr APIError) {
	traceID := getTraceID(c)

	// Log the error
	log.Error().
		Str("trace_id", traceID).
		Str("error_code", apiErr.ErrorCode()).
		Int("status_code", apiErr.StatusCode()).
		Str("method", c.Request.Method).
		Str("path", c.Request.URL.Path).
		Str("user_agent", c.Request.UserAgent()).
		Str("client_ip", c.ClientIP()).
		Err(apiErr).
		Str("stack_trace", apiErr.StackTrace()).
		Msg("API error occurred")

	// Prepare response body
	errorResponse := gin.H{
		"error": gin.H{
			"code":    apiErr.ErrorCode(),
			"message": apiErr.Error(),
		},
	}

	// Add trace ID to response if available
	if traceID != "" {
		errorResponse["trace_id"] = traceID
	}

	// Add stack trace in development environment
	if isDevelopment() && apiErr.StackTrace() != "" {
		errorResponse["stack_trace"] = apiErr.StackTrace()
	}

	c.AbortWithStatusJSON(apiErr.StatusCode(), errorResponse)
}

// getTraceID extracts trace ID from gin context
func getTraceID(c *gin.Context) string {
	if traceID, exists := c.Get("trace_id"); exists {
		if traceStr, ok := traceID.(string); ok {
			return traceStr
		}
	}
	return ""
}

// isDevelopment checks if we're running in development environment
func isDevelopment() bool {
	env := os.Getenv("GIN_MODE")
	return env == "" || env == "debug"
}

// AbortWithAPIError is a helper function to abort with an APIError
func AbortWithAPIError(c *gin.Context, err APIError) {
	// Add context to error if trace ID is available
	if traceID := getTraceID(c); traceID != "" {
		err = err.WithContext(c.Request.Context())
	}

	// Add error to gin context and abort
	c.Error(err)
	c.Abort()
}

// AbortWithError is a helper function to convert regular error to APIError and abort
func AbortWithError(c *gin.Context, err error, message string) {
	if err == nil {
		return
	}

	var apiErr APIError
	if existingAPIErr, ok := err.(APIError); ok {
		apiErr = existingAPIErr
	} else {
		apiErr = Wrap(err, message)
	}

	AbortWithAPIError(c, apiErr)
}

// AbortWithValidationError is a helper function to abort with a validation error
func AbortWithValidationError(c *gin.Context, message string) {
	err := ValidationError(message)
	AbortWithAPIError(c, err)
}

// AbortWithInternalError is a helper function to abort with an internal server error
func AbortWithInternalError(c *gin.Context, message string) {
	err := InternalError(message)
	AbortWithAPIError(c, err)
}
