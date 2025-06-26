package error

import (
	"context"
	"fmt"
	"runtime"
	"strings"
)

// APIError interface with HTTP status code support
type APIError interface {
	error
	StatusCode() int
	ErrorCode() string
	WithContext(ctx context.Context) APIError
	WithStackTrace() APIError
	StackTrace() string
	TraceID() string
}

// BaseError implements APIError
type BaseError struct {
	Code    string
	Message string
	Status  int
	Cause   error
	Stack   string
	Ctx     context.Context
	Trace   string
}

func (e *BaseError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *BaseError) StatusCode() int {
	return e.Status
}

func (e *BaseError) ErrorCode() string {
	return e.Code
}

func (e *BaseError) WithContext(ctx context.Context) APIError {
	newErr := *e
	newErr.Ctx = ctx

	// Extract trace ID from context if available
	if traceID := ctx.Value("trace_id"); traceID != nil {
		if traceStr, ok := traceID.(string); ok {
			newErr.Trace = traceStr
		}
	}

	return &newErr
}

func (e *BaseError) WithStackTrace() APIError {
	newErr := *e
	newErr.Stack = captureStack()
	return &newErr
}

func (e *BaseError) StackTrace() string {
	return e.Stack
}

func (e *BaseError) TraceID() string {
	return e.Trace
}

func (e *BaseError) Unwrap() error {
	return e.Cause
}

// captureStack captures the current stack trace
func captureStack() string {
	var stack []string
	for i := 2; i < 10; i++ { // Skip current function and WithStackTrace
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		// Only include our application code
		if !strings.Contains(file, "millioncheckboxes-api") {
			continue
		}

		stack = append(stack, fmt.Sprintf("%s:%d %s", file, line, fn.Name()))
	}

	return strings.Join(stack, "\n")
}
