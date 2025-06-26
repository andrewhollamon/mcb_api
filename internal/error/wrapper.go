package error

import (
	"context"
)

// NewAPIError creates a new APIError with the given code, message, and status
func NewAPIError(code, message string, status int) APIError {
	return &BaseError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

// NewAPIErrorFromCode creates a new APIError using a predefined error code
func NewAPIErrorFromCode(code string, message string) APIError {
	status := GetStatusCode(code)
	return &BaseError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

// Wrap wraps an existing error with additional message and creates an APIError
func Wrap(err error, message string) APIError {
	if err == nil {
		return nil
	}

	// If it's already an APIError, preserve the original error code and status
	if apiErr, ok := err.(APIError); ok {
		return &BaseError{
			Code:    apiErr.ErrorCode(),
			Message: message,
			Status:  apiErr.StatusCode(),
			Cause:   err,
		}
	}

	return &BaseError{
		Code:    ErrInternalServer,
		Message: message,
		Status:  GetStatusCode(ErrInternalServer),
		Cause:   err,
	}
}

// WrapWithCode wraps an existing error with a specific error code, message, and status
func WrapWithCode(err error, code, message string, status int) APIError {
	if err == nil {
		return nil
	}

	return &BaseError{
		Code:    code,
		Message: message,
		Status:  status,
		Cause:   err,
	}
}

// WrapWithCodeFromConstants wraps an existing error using predefined error constants
func WrapWithCodeFromConstants(err error, code string, message string) APIError {
	if err == nil {
		return nil
	}

	status := GetStatusCode(code)
	return &BaseError{
		Code:    code,
		Message: message,
		Status:  status,
		Cause:   err,
	}
}

// ValidationError creates a validation error with a specific message
func ValidationError(message string) APIError {
	return NewAPIErrorFromCode(ErrValidationFailed, message)
}

// InternalError creates an internal server error with a specific message
func InternalError(message string) APIError {
	return NewAPIErrorFromCode(ErrInternalServer, message)
}

// QueueError creates a queue unavailable error with a specific message
func QueueError(message string) APIError {
	return NewAPIErrorFromCode(ErrQueueUnavailable, message)
}

// DatabaseError creates a database error with a specific message
func DatabaseError(message string) APIError {
	return NewAPIErrorFromCode(ErrDatabaseError, message)
}

// WithContext adds context to an APIError
func WithContext(err APIError, ctx context.Context) APIError {
	if err == nil {
		return nil
	}
	return err.WithContext(ctx)
}

// WithStackTrace adds stack trace to an APIError
func WithStackTrace(err APIError) APIError {
	if err == nil {
		return nil
	}
	return err.WithStackTrace()
}

// IsErrorType checks if an error is of a specific error code type
func IsErrorType(err error, errorCode string) bool {
	if apiErr, ok := err.(APIError); ok {
		return apiErr.ErrorCode() == errorCode
	}
	return false
}

// GetErrorCode extracts the error code from an error, returns empty string if not an APIError
func GetErrorCode(err error) string {
	if apiErr, ok := err.(APIError); ok {
		return apiErr.ErrorCode()
	}
	return ""
}

// GetStatusCodeFromError extracts the HTTP status code from an error
func GetStatusCodeFromError(err error) int {
	if apiErr, ok := err.(APIError); ok {
		return apiErr.StatusCode()
	}
	return GetStatusCode(ErrInternalServer)
}
