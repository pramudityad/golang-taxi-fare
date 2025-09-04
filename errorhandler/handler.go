// Package errorhandler provides centralized error processing and exit code management
// for the taxi fare calculation system. It categorizes errors and provides appropriate
// exit codes for different error conditions.
package errorhandler

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"golang-taxi-fare/datavalidator"
	"golang-taxi-fare/inputparser"
)

// ExitCode represents different application exit codes
type ExitCode int

const (
	// ExitSuccess indicates successful execution
	ExitSuccess ExitCode = 0
	// ExitFormatError indicates format validation errors (exit code 1)
	ExitFormatError ExitCode = 1
	// ExitTimingError indicates timing constraint violations (exit code 2)
	ExitTimingError ExitCode = 2
	// ExitInsufficientData indicates insufficient data for processing (exit code 3)
	ExitInsufficientData ExitCode = 3
	// ExitCalculationError indicates calculation or processing errors (exit code 4)
	ExitCalculationError ExitCode = 4
	// ExitGeneralError indicates general application errors (exit code 5)
	ExitGeneralError ExitCode = 5
)

// String returns a human-readable description of the exit code
func (ec ExitCode) String() string {
	switch ec {
	case ExitSuccess:
		return "success"
	case ExitFormatError:
		return "format error"
	case ExitTimingError:
		return "timing error"
	case ExitInsufficientData:
		return "insufficient data"
	case ExitCalculationError:
		return "calculation error"
	case ExitGeneralError:
		return "general error"
	default:
		return "unknown error"
	}
}

// ErrorContext provides detailed context information for error handling
type ErrorContext struct {
	// Timestamp when the error occurred
	Timestamp time.Time `json:"timestamp"`
	// ErrorType categorizes the type of error
	ErrorType string `json:"error_type"`
	// Message contains the main error message
	Message string `json:"message"`
	// Details provides additional error details
	Details string `json:"details,omitempty"`
	// StackTrace contains the call stack at the time of error
	StackTrace []string `json:"stack_trace,omitempty"`
	// Context provides additional contextual information
	Context map[string]interface{} `json:"context,omitempty"`
}

// String implements the Stringer interface for ErrorContext
func (ec ErrorContext) String() string {
	return fmt.Sprintf("ErrorContext{Type: %s, Message: %s, Timestamp: %s}",
		ec.ErrorType, ec.Message, ec.Timestamp.Format("15:04:05.000"))
}

// ErrorHandler defines the interface for error handling operations
type ErrorHandler interface {
	// HandleError processes an error and returns the appropriate exit code
	HandleError(err error) ExitCode
	
	// HandleErrorWithContext processes an error with additional context
	HandleErrorWithContext(err error, context map[string]interface{}) ExitCode
	
	// CreateErrorContext creates an ErrorContext from an error
	CreateErrorContext(err error, context map[string]interface{}) ErrorContext
}

// ApplicationErrorHandler implements the ErrorHandler interface
type ApplicationErrorHandler struct {
	// CaptureStackTrace determines whether to capture stack traces
	CaptureStackTrace bool
	// ExitOnError determines whether to call os.Exit when handling errors
	ExitOnError bool
}

// NewErrorHandler creates a new ApplicationErrorHandler with default settings
func NewErrorHandler() ErrorHandler {
	return &ApplicationErrorHandler{
		CaptureStackTrace: true,
		ExitOnError:       true,
	}
}

// NewErrorHandlerWithOptions creates a new ApplicationErrorHandler with custom options
func NewErrorHandlerWithOptions(captureStackTrace, exitOnError bool) ErrorHandler {
	return &ApplicationErrorHandler{
		CaptureStackTrace: captureStackTrace,
		ExitOnError:       exitOnError,
	}
}

// HandleError processes an error and returns the appropriate exit code
func (aeh *ApplicationErrorHandler) HandleError(err error) ExitCode {
	return aeh.HandleErrorWithContext(err, nil)
}

// HandleErrorWithContext processes an error with additional context
func (aeh *ApplicationErrorHandler) HandleErrorWithContext(err error, context map[string]interface{}) ExitCode {
	if err == nil {
		return ExitSuccess
	}
	
	exitCode := aeh.categorizeError(err)
	errorContext := aeh.CreateErrorContext(err, context)
	
	// Print error information to stderr
	fmt.Fprintf(os.Stderr, "Error: %s\n", errorContext.Message)
	if errorContext.Details != "" {
		fmt.Fprintf(os.Stderr, "Details: %s\n", errorContext.Details)
	}
	
	// Exit if configured to do so
	if aeh.ExitOnError {
		os.Exit(int(exitCode))
	}
	
	return exitCode
}

// CreateErrorContext creates an ErrorContext from an error
func (aeh *ApplicationErrorHandler) CreateErrorContext(err error, context map[string]interface{}) ErrorContext {
	if err == nil {
		return ErrorContext{
			Timestamp: time.Now(),
			ErrorType: "none",
			Message:   "no error",
		}
	}
	
	errorContext := ErrorContext{
		Timestamp: time.Now(),
		Message:   err.Error(),
		Context:   context,
	}
	
	// Categorize error type and add details
	switch e := err.(type) {
	case *datavalidator.ValidationError:
		errorContext.ErrorType = "validation"
		errorContext.Details = fmt.Sprintf("Validation failed: %s (type: %s)", e.Message, e.Type.String())
		if e.RecordIndex >= 0 {
			if errorContext.Context == nil {
				errorContext.Context = make(map[string]interface{})
			}
			errorContext.Context["record_index"] = e.RecordIndex
			errorContext.Context["field"] = e.Field
		}
		
	case *inputparser.ParsingError:
		errorContext.ErrorType = "parsing"
		errorContext.Details = fmt.Sprintf("Parsing failed: %s (type: %s)", e.Message, e.Type.String())
		if e.Line > 0 {
			if errorContext.Context == nil {
				errorContext.Context = make(map[string]interface{})
			}
			errorContext.Context["line_number"] = e.Line
			errorContext.Context["input"] = e.Input
		}
		
	default:
		errorContext.ErrorType = "general"
		errorContext.Details = fmt.Sprintf("Unexpected error: %s", err.Error())
	}
	
	// Capture stack trace if enabled
	if aeh.CaptureStackTrace {
		errorContext.StackTrace = captureStackTrace()
	}
	
	return errorContext
}

// categorizeError determines the appropriate exit code for an error
func (aeh *ApplicationErrorHandler) categorizeError(err error) ExitCode {
	switch e := err.(type) {
	case *datavalidator.ValidationError:
		switch e.Type {
		case datavalidator.ValidationErrorTypeTiming:
			return ExitTimingError
		case datavalidator.ValidationErrorTypeFormat:
			return ExitFormatError
		case datavalidator.ValidationErrorTypeMileage:
			return ExitTimingError // Mileage progression is a timing-related constraint
		case datavalidator.ValidationErrorTypeSequence:
			return ExitInsufficientData
		case datavalidator.ValidationErrorTypeConstraint:
			return ExitFormatError
		default:
			return ExitGeneralError
		}
		
	case *inputparser.ParsingError:
		switch e.Type {
		case inputparser.ErrorTypeFormat:
			return ExitFormatError
		case inputparser.ErrorTypeTimestamp:
			return ExitFormatError
		case inputparser.ErrorTypeDistance:
			return ExitFormatError
		case inputparser.ErrorTypeIO:
			return ExitGeneralError
		default:
			return ExitGeneralError
		}
		
	default:
		// Check for common error patterns
		errStr := err.Error()
		switch {
		case containsKeyword(errStr, "format", "invalid", "malformed"):
			return ExitFormatError
		case containsKeyword(errStr, "timing", "time", "sequence"):
			return ExitTimingError
		case containsKeyword(errStr, "insufficient", "empty", "missing"):
			return ExitInsufficientData
		case containsKeyword(errStr, "calculation", "compute", "arithmetic"):
			return ExitCalculationError
		default:
			return ExitGeneralError
		}
	}
}

// captureStackTrace captures the current call stack
func captureStackTrace() []string {
	const maxDepth = 10
	stackTrace := make([]string, 0, maxDepth)
	
	for i := 2; i < maxDepth+2; i++ { // Skip captureStackTrace and CreateErrorContext
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			stackTrace = append(stackTrace, fmt.Sprintf("%s:%d", file, line))
		} else {
			stackTrace = append(stackTrace, fmt.Sprintf("%s:%d %s", file, line, fn.Name()))
		}
	}
	
	return stackTrace
}

// containsKeyword checks if a string contains any of the specified keywords
func containsKeyword(s string, keywords ...string) bool {
	lowerS := toLower(s)
	for _, keyword := range keywords {
		if contains(lowerS, toLower(keyword)) {
			return true
		}
	}
	return false
}

// toLower converts a string to lowercase (simple implementation)
func toLower(s string) string {
	result := make([]rune, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + 32
		} else {
			result[i] = r
		}
	}
	return string(result)
}

// contains checks if a string contains a substring (simple implementation)
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}