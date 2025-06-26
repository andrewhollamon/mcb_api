package error

import "net/http"

// Error constants with fixed string values
const (
	// Validation errors
	ErrValidationFailed     = "VALIDATION_FAILED"
	ErrMissingParameter     = "MISSING_PARAMETER"
	ErrInvalidParameter     = "INVALID_PARAMETER"
	ErrParameterOutOfRange  = "PARAMETER_OUT_OF_RANGE"
	ErrInvalidUUID          = "INVALID_UUID"
	ErrInvalidCheckboxNumber = "INVALID_CHECKBOX_NUMBER"
	
	// Server errors
	ErrInternalServer    = "INTERNAL_SERVER_ERROR"
	ErrServiceUnavailable = "SERVICE_UNAVAILABLE"
	ErrTimeout           = "TIMEOUT"
	
	// Queue errors
	ErrQueueUnavailable  = "QUEUE_UNAVAILABLE"
	ErrQueueTimeout      = "QUEUE_TIMEOUT"
	ErrQueueFull         = "QUEUE_FULL"
	ErrMessageTooLarge   = "MESSAGE_TOO_LARGE"
	
	// Database errors
	ErrDatabaseError     = "DATABASE_ERROR"
	ErrDatabaseTimeout   = "DATABASE_TIMEOUT"
	ErrDatabaseConnection = "DATABASE_CONNECTION_ERROR"
	ErrRecordNotFound    = "RECORD_NOT_FOUND"
	ErrDuplicateRecord   = "DUPLICATE_RECORD"
	
	// Memory store errors
	ErrMemoryStoreError  = "MEMORY_STORE_ERROR"
	ErrMemoryStoreFull   = "MEMORY_STORE_FULL"
	
	// Authentication/Authorization errors
	ErrUnauthorized      = "UNAUTHORIZED"
	ErrForbidden         = "FORBIDDEN"
	ErrTokenExpired      = "TOKEN_EXPIRED"
	ErrInvalidToken      = "INVALID_TOKEN"
)

// ErrorCodeToStatus maps error codes to HTTP status codes
var ErrorCodeToStatus = map[string]int{
	// Validation errors - 400 Bad Request
	ErrValidationFailed:     http.StatusBadRequest,
	ErrMissingParameter:     http.StatusBadRequest,
	ErrInvalidParameter:     http.StatusBadRequest,
	ErrParameterOutOfRange:  http.StatusBadRequest,
	ErrInvalidUUID:          http.StatusBadRequest,
	ErrInvalidCheckboxNumber: http.StatusBadRequest,
	
	// Server errors - 500 Internal Server Error
	ErrInternalServer:       http.StatusInternalServerError,
	ErrDatabaseError:        http.StatusInternalServerError,
	ErrDatabaseTimeout:      http.StatusInternalServerError,
	ErrDatabaseConnection:   http.StatusInternalServerError,
	ErrMemoryStoreError:     http.StatusInternalServerError,
	
	// Service unavailable - 503
	ErrServiceUnavailable:   http.StatusServiceUnavailable,
	ErrQueueUnavailable:     http.StatusServiceUnavailable,
	ErrQueueTimeout:         http.StatusServiceUnavailable,
	ErrQueueFull:           http.StatusServiceUnavailable,
	ErrMemoryStoreFull:     http.StatusServiceUnavailable,
	
	// Request timeout - 408
	ErrTimeout:             http.StatusRequestTimeout,
	
	// Bad request - 400
	ErrMessageTooLarge:     http.StatusBadRequest,
	
	// Not found - 404
	ErrRecordNotFound:      http.StatusNotFound,
	
	// Conflict - 409
	ErrDuplicateRecord:     http.StatusConflict,
	
	// Authentication/Authorization
	ErrUnauthorized:        http.StatusUnauthorized,
	ErrForbidden:          http.StatusForbidden,
	ErrTokenExpired:       http.StatusUnauthorized,
	ErrInvalidToken:       http.StatusUnauthorized,
}

// GetStatusCode returns the HTTP status code for an error code
func GetStatusCode(errorCode string) int {
	if status, exists := ErrorCodeToStatus[errorCode]; exists {
		return status
	}
	return http.StatusInternalServerError
}