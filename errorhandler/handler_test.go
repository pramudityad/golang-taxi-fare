package errorhandler

import (
	"errors"
	"testing"
	"time"

	"golang-taxi-fare/datavalidator"
	"golang-taxi-fare/inputparser"
)

func TestExitCode_String(t *testing.T) {
	tests := []struct {
		name     string
		exitCode ExitCode
		expected string
	}{
		{"success", ExitSuccess, "success"},
		{"format error", ExitFormatError, "format error"},
		{"timing error", ExitTimingError, "timing error"},
		{"insufficient data", ExitInsufficientData, "insufficient data"},
		{"calculation error", ExitCalculationError, "calculation error"},
		{"general error", ExitGeneralError, "general error"},
		{"unknown", ExitCode(99), "unknown error"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.exitCode.String(); got != tt.expected {
				t.Errorf("ExitCode.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestErrorContext_String(t *testing.T) {
	timestamp := time.Date(2023, 1, 1, 12, 30, 45, 123000000, time.UTC)
	ec := ErrorContext{
		Timestamp: timestamp,
		ErrorType: "validation",
		Message:   "test error message",
	}
	
	str := ec.String()
	expectedSubstrings := []string{"validation", "test error message", "12:30:45.123"}
	
	for _, substr := range expectedSubstrings {
		if !contains(str, substr) {
			t.Errorf("ErrorContext.String() = %q, should contain %q", str, substr)
		}
	}
}

func TestNewErrorHandler(t *testing.T) {
	handler := NewErrorHandler()
	
	if handler == nil {
		t.Error("Expected non-nil handler")
	}
	
	// Test that it implements the ErrorHandler interface
	_, ok := handler.(ErrorHandler)
	if !ok {
		t.Error("Handler should implement ErrorHandler interface")
	}
	
	// Test default settings
	appHandler, ok := handler.(*ApplicationErrorHandler)
	if !ok {
		t.Fatal("Expected *ApplicationErrorHandler")
	}
	
	if !appHandler.CaptureStackTrace {
		t.Error("Expected CaptureStackTrace to be true by default")
	}
	
	if !appHandler.ExitOnError {
		t.Error("Expected ExitOnError to be true by default")
	}
}

func TestNewErrorHandlerWithOptions(t *testing.T) {
	handler := NewErrorHandlerWithOptions(false, false)
	appHandler, ok := handler.(*ApplicationErrorHandler)
	if !ok {
		t.Fatal("Expected *ApplicationErrorHandler")
	}
	
	if appHandler.CaptureStackTrace {
		t.Error("Expected CaptureStackTrace to be false")
	}
	
	if appHandler.ExitOnError {
		t.Error("Expected ExitOnError to be false")
	}
}

func TestApplicationErrorHandler_HandleError(t *testing.T) {
	// Use a handler that doesn't exit so we can test
	handler := NewErrorHandlerWithOptions(false, false).(*ApplicationErrorHandler)
	
	tests := []struct {
		name         string
		err          error
		expectedCode ExitCode
	}{
		{"no error", nil, ExitSuccess},
		{"validation format error", datavalidator.FormatError(0, "field", "invalid format", "input"), ExitFormatError},
		{"validation timing error", datavalidator.TimingError(1, "timing issue", "timestamp"), ExitTimingError},
		{"validation mileage error", datavalidator.MileageError(2, "mileage issue", "distance"), ExitTimingError},
		{"validation sequence error", datavalidator.SequenceError("empty sequence", 0), ExitInsufficientData},
		{"validation constraint error", datavalidator.ConstraintError(3, "field", "constraint violation", "value"), ExitFormatError},
		{"parsing format error", &inputparser.ParsingError{Type: inputparser.ErrorTypeFormat, Message: "format error"}, ExitFormatError},
		{"parsing timestamp error", &inputparser.ParsingError{Type: inputparser.ErrorTypeTimestamp, Message: "timestamp error"}, ExitFormatError},
		{"parsing distance error", &inputparser.ParsingError{Type: inputparser.ErrorTypeDistance, Message: "distance error"}, ExitFormatError},
		{"parsing IO error", &inputparser.ParsingError{Type: inputparser.ErrorTypeIO, Message: "IO error"}, ExitGeneralError},
		{"general error with format keyword", errors.New("invalid format detected"), ExitFormatError},
		{"general error with timing keyword", errors.New("timing sequence violated"), ExitTimingError},
		{"general error with insufficient keyword", errors.New("insufficient data provided"), ExitInsufficientData},
		{"general error with calculation keyword", errors.New("calculation failed"), ExitCalculationError},
		{"unknown general error", errors.New("unknown problem"), ExitGeneralError},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handler.HandleError(tt.err)
			if got != tt.expectedCode {
				t.Errorf("HandleError() = %v, want %v", got, tt.expectedCode)
			}
		})
	}
}

func TestApplicationErrorHandler_HandleErrorWithContext(t *testing.T) {
	handler := NewErrorHandlerWithOptions(false, false).(*ApplicationErrorHandler)
	
	err := errors.New("test error")
	context := map[string]interface{}{
		"user_id": 12345,
		"action":  "processing",
	}
	
	got := handler.HandleErrorWithContext(err, context)
	if got != ExitGeneralError {
		t.Errorf("HandleErrorWithContext() = %v, want %v", got, ExitGeneralError)
	}
}

func TestApplicationErrorHandler_CreateErrorContext(t *testing.T) {
	handler := NewErrorHandlerWithOptions(false, false).(*ApplicationErrorHandler)
	
	t.Run("nil error", func(t *testing.T) {
		ctx := handler.CreateErrorContext(nil, nil)
		if ctx.ErrorType != "none" {
			t.Errorf("Expected error type 'none', got %s", ctx.ErrorType)
		}
		if ctx.Message != "no error" {
			t.Errorf("Expected message 'no error', got %s", ctx.Message)
		}
	})
	
	t.Run("validation error", func(t *testing.T) {
		validationErr := datavalidator.TimingError(5, "timing constraint violated", "12:30:45.123")
		context := map[string]interface{}{"test": "value"}
		
		ctx := handler.CreateErrorContext(validationErr, context)
		
		if ctx.ErrorType != "validation" {
			t.Errorf("Expected error type 'validation', got %s", ctx.ErrorType)
		}
		
		if ctx.Message == "" {
			t.Error("Expected non-empty message")
		}
		
		if ctx.Context == nil {
			t.Error("Expected context to be preserved")
		}
		
		if ctx.Context["record_index"] != 5 {
			t.Errorf("Expected record_index 5, got %v", ctx.Context["record_index"])
		}
		
		if ctx.Context["field"] != "timestamp" {
			t.Errorf("Expected field 'timestamp', got %v", ctx.Context["field"])
		}
		
		if ctx.Context["test"] != "value" {
			t.Errorf("Expected original context to be preserved")
		}
	})
	
	t.Run("parsing error", func(t *testing.T) {
		parsingErr := &inputparser.ParsingError{
			Type:    inputparser.ErrorTypeFormat,
			Message: "invalid format",
			Line:    10,
			Input:   "malformed input",
		}
		
		ctx := handler.CreateErrorContext(parsingErr, nil)
		
		if ctx.ErrorType != "parsing" {
			t.Errorf("Expected error type 'parsing', got %s", ctx.ErrorType)
		}
		
		if ctx.Context["line_number"] != 10 {
			t.Errorf("Expected line_number 10, got %v", ctx.Context["line_number"])
		}
		
		if ctx.Context["input"] != "malformed input" {
			t.Errorf("Expected input 'malformed input', got %v", ctx.Context["input"])
		}
	})
	
	t.Run("general error", func(t *testing.T) {
		generalErr := errors.New("unknown error")
		
		ctx := handler.CreateErrorContext(generalErr, nil)
		
		if ctx.ErrorType != "general" {
			t.Errorf("Expected error type 'general', got %s", ctx.ErrorType)
		}
		
		if ctx.Message != "unknown error" {
			t.Errorf("Expected message 'unknown error', got %s", ctx.Message)
		}
	})
}

func TestApplicationErrorHandler_CreateErrorContextWithStackTrace(t *testing.T) {
	handler := NewErrorHandlerWithOptions(true, false).(*ApplicationErrorHandler)
	
	err := errors.New("test error")
	ctx := handler.CreateErrorContext(err, nil)
	
	if len(ctx.StackTrace) == 0 {
		t.Error("Expected stack trace to be captured")
	}
	
	// Verify stack trace contains meaningful information
	found := false
	for _, frame := range ctx.StackTrace {
		if contains(frame, "TestApplicationErrorHandler_CreateErrorContextWithStackTrace") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected stack trace to contain test function name")
	}
}

func TestCategorizeError(t *testing.T) {
	handler := NewErrorHandlerWithOptions(false, false).(*ApplicationErrorHandler)
	
	tests := []struct {
		name     string
		err      error
		expected ExitCode
	}{
		{
			"validation timing error",
			&datavalidator.ValidationError{Type: datavalidator.ValidationErrorTypeTiming},
			ExitTimingError,
		},
		{
			"validation format error",
			&datavalidator.ValidationError{Type: datavalidator.ValidationErrorTypeFormat},
			ExitFormatError,
		},
		{
			"validation mileage error",
			&datavalidator.ValidationError{Type: datavalidator.ValidationErrorTypeMileage},
			ExitTimingError,
		},
		{
			"validation sequence error",
			&datavalidator.ValidationError{Type: datavalidator.ValidationErrorTypeSequence},
			ExitInsufficientData,
		},
		{
			"validation constraint error",
			&datavalidator.ValidationError{Type: datavalidator.ValidationErrorTypeConstraint},
			ExitFormatError,
		},
		{
			"parsing format error",
			&inputparser.ParsingError{Type: inputparser.ErrorTypeFormat},
			ExitFormatError,
		},
		{
			"parsing IO error",
			&inputparser.ParsingError{Type: inputparser.ErrorTypeIO},
			ExitGeneralError,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handler.categorizeError(tt.err)
			if got != tt.expected {
				t.Errorf("categorizeError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestContainsKeyword(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		keywords []string
		expected bool
	}{
		{"contains format", "Invalid format detected", []string{"format"}, true},
		{"contains timing", "Timing sequence violation", []string{"timing"}, true},
		{"case insensitive", "INVALID FORMAT", []string{"format"}, true},
		{"multiple keywords match", "format and timing error", []string{"format", "timing"}, true},
		{"no match", "unknown error", []string{"format", "timing"}, false},
		{"empty keywords", "any text", []string{}, false},
		{"empty text", "", []string{"format"}, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsKeyword(tt.text, tt.keywords...)
			if got != tt.expected {
				t.Errorf("containsKeyword(%q, %v) = %v, want %v", tt.text, tt.keywords, got, tt.expected)
			}
		})
	}
}

func TestToLower(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ABC", "abc"},
		{"Hello World", "hello world"},
		{"MiXeD cAsE", "mixed case"},
		{"123", "123"},
		{"", ""},
		{"already lowercase", "already lowercase"},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toLower(tt.input)
			if got != tt.expected {
				t.Errorf("toLower(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{"contains substring", "hello world", "world", true},
		{"contains at start", "hello world", "hello", true},
		{"contains at end", "hello world", "world", true},
		{"does not contain", "hello world", "xyz", false},
		{"empty substring", "hello world", "", true},
		{"empty string", "", "world", false},
		{"exact match", "world", "world", true},
		{"case sensitive", "Hello", "hello", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contains(tt.s, tt.substr)
			if got != tt.expected {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.expected)
			}
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkHandleError(b *testing.B) {
	handler := NewErrorHandlerWithOptions(false, false).(*ApplicationErrorHandler)
	err := errors.New("test error")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.HandleError(err)
	}
}

func BenchmarkCreateErrorContext(b *testing.B) {
	handler := NewErrorHandlerWithOptions(false, false).(*ApplicationErrorHandler)
	err := datavalidator.TimingError(5, "timing error", "input")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.CreateErrorContext(err, nil)
	}
}

func BenchmarkCreateErrorContextWithStackTrace(b *testing.B) {
	handler := NewErrorHandlerWithOptions(true, false).(*ApplicationErrorHandler)
	err := errors.New("test error")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.CreateErrorContext(err, nil)
	}
}