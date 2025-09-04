// Package loggingsystem provides structured JSON logging functionality for the taxi fare calculation system.
// It uses Go 1.21+ log/slog package for high-performance structured logging with contextual information.
package loggingsystem

import (
	"context"
	"io"
	"log/slog"
	"os"
	"time"
)

// LogLevel represents different logging levels
type LogLevel int

const (
	// LevelDebug provides detailed debugging information
	LevelDebug LogLevel = iota
	// LevelInfo provides general information messages
	LevelInfo
	// LevelWarn provides warning messages for potentially problematic situations
	LevelWarn
	// LevelError provides error messages for error conditions
	LevelError
)

// String returns a human-readable description of the log level
func (ll LogLevel) String() string {
	switch ll {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ToSlogLevel converts LogLevel to slog.Level
func (ll LogLevel) ToSlogLevel() slog.Level {
	switch ll {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// LogEntry represents a structured log entry with contextual information
type LogEntry struct {
	// Timestamp when the log entry was created
	Timestamp time.Time `json:"timestamp"`
	// Level indicates the severity/importance of the log entry
	Level string `json:"level"`
	// Message contains the main log message
	Message string `json:"message"`
	// Context provides additional contextual information
	Context map[string]interface{} `json:"context,omitempty"`
	// Component identifies which part of the system generated the log
	Component string `json:"component,omitempty"`
	// RecordID links the log entry to a specific record being processed
	RecordID string `json:"record_id,omitempty"`
	// ProcessingState indicates the current state of processing
	ProcessingState string `json:"processing_state,omitempty"`
}

// Logger defines the interface for logging operations
type Logger interface {
	// Debug logs a debug-level message with optional context
	Debug(message string, keyValues ...interface{})
	
	// Info logs an info-level message with optional context
	Info(message string, keyValues ...interface{})
	
	// Warn logs a warning-level message with optional context
	Warn(message string, keyValues ...interface{})
	
	// Error logs an error-level message with optional context
	Error(message string, keyValues ...interface{})
	
	// LogWithLevel logs a message at the specified level with context
	LogWithLevel(level LogLevel, message string, keyValues ...interface{})
	
	// WithContext creates a new logger with additional context
	WithContext(context map[string]interface{}) Logger
	
	// WithComponent creates a new logger with component identification
	WithComponent(component string) Logger
	
	// WithRecordID creates a new logger with record ID context
	WithRecordID(recordID string) Logger
	
	// WithProcessingState creates a new logger with processing state context
	WithProcessingState(state string) Logger
	
	// SetLevel sets the minimum logging level
	SetLevel(level LogLevel)
	
	// IsEnabled checks if a log level is enabled
	IsEnabled(level LogLevel) bool
}

// StructuredLogger implements the Logger interface using Go's slog package
type StructuredLogger struct {
	slogger         *slog.Logger
	minLevel        LogLevel
	baseContext     map[string]interface{}
	component       string
	recordID        string
	processingState string
	output          io.Writer // Keep track of output for level changes
}

// NewLogger creates a new StructuredLogger with JSON output to stderr
func NewLogger() Logger {
	return NewLoggerWithOptions(os.Stderr, LevelInfo)
}

// NewLoggerWithOptions creates a new StructuredLogger with custom options
func NewLoggerWithOptions(output io.Writer, minLevel LogLevel) Logger {
	// Create JSON handler for structured logging
	handler := slog.NewJSONHandler(output, &slog.HandlerOptions{
		Level:     minLevel.ToSlogLevel(),
		AddSource: false, // We'll add our own contextual information
	})
	
	slogger := slog.New(handler)
	
	return &StructuredLogger{
		slogger:     slogger,
		minLevel:    minLevel,
		baseContext: make(map[string]interface{}),
		output:      output,
	}
}

// Debug logs a debug-level message with optional context
func (sl *StructuredLogger) Debug(message string, keyValues ...interface{}) {
	sl.LogWithLevel(LevelDebug, message, keyValues...)
}

// Info logs an info-level message with optional context
func (sl *StructuredLogger) Info(message string, keyValues ...interface{}) {
	sl.LogWithLevel(LevelInfo, message, keyValues...)
}

// Warn logs a warning-level message with optional context
func (sl *StructuredLogger) Warn(message string, keyValues ...interface{}) {
	sl.LogWithLevel(LevelWarn, message, keyValues...)
}

// Error logs an error-level message with optional context
func (sl *StructuredLogger) Error(message string, keyValues ...interface{}) {
	sl.LogWithLevel(LevelError, message, keyValues...)
}

// LogWithLevel logs a message at the specified level with context
func (sl *StructuredLogger) LogWithLevel(level LogLevel, message string, keyValues ...interface{}) {
	if !sl.IsEnabled(level) {
		return
	}
	
	// Build attributes from context and logger state
	attrs := make([]slog.Attr, 0, len(keyValues)/2+10) // Pre-allocate for performance
	
	// Add component if set
	if sl.component != "" {
		attrs = append(attrs, slog.String("component", sl.component))
	}
	
	// Add record ID if set
	if sl.recordID != "" {
		attrs = append(attrs, slog.String("record_id", sl.recordID))
	}
	
	// Add processing state if set
	if sl.processingState != "" {
		attrs = append(attrs, slog.String("processing_state", sl.processingState))
	}
	
	// Add base context
	for key, value := range sl.baseContext {
		attrs = append(attrs, slog.Any(key, value))
	}
	
	// Add provided context (expects key-value pairs)
	for i := 0; i < len(keyValues); i += 2 {
		if i+1 < len(keyValues) {
			key := keyValues[i]
			value := keyValues[i+1]
			if keyStr, ok := key.(string); ok {
				attrs = append(attrs, slog.Any(keyStr, value))
			}
		}
	}
	
	// Log with the appropriate slog level
	ctx := context.Background()
	sl.slogger.LogAttrs(ctx, level.ToSlogLevel(), message, attrs...)
}

// WithContext creates a new logger with additional context
func (sl *StructuredLogger) WithContext(context map[string]interface{}) Logger {
	newContext := make(map[string]interface{})
	
	// Copy existing context
	for k, v := range sl.baseContext {
		newContext[k] = v
	}
	
	// Add new context
	for k, v := range context {
		newContext[k] = v
	}
	
	return &StructuredLogger{
		slogger:         sl.slogger,
		minLevel:        sl.minLevel,
		baseContext:     newContext,
		component:       sl.component,
		recordID:        sl.recordID,
		processingState: sl.processingState,
		output:          sl.output,
	}
}

// WithComponent creates a new logger with component identification
func (sl *StructuredLogger) WithComponent(component string) Logger {
	return &StructuredLogger{
		slogger:         sl.slogger,
		minLevel:        sl.minLevel,
		baseContext:     sl.baseContext,
		component:       component,
		recordID:        sl.recordID,
		processingState: sl.processingState,
		output:          sl.output,
	}
}

// WithRecordID creates a new logger with record ID context
func (sl *StructuredLogger) WithRecordID(recordID string) Logger {
	return &StructuredLogger{
		slogger:         sl.slogger,
		minLevel:        sl.minLevel,
		baseContext:     sl.baseContext,
		component:       sl.component,
		recordID:        recordID,
		processingState: sl.processingState,
		output:          sl.output,
	}
}

// WithProcessingState creates a new logger with processing state context
func (sl *StructuredLogger) WithProcessingState(state string) Logger {
	return &StructuredLogger{
		slogger:         sl.slogger,
		minLevel:        sl.minLevel,
		baseContext:     sl.baseContext,
		component:       sl.component,
		recordID:        sl.recordID,
		processingState: state,
		output:          sl.output,
	}
}

// SetLevel sets the minimum logging level
func (sl *StructuredLogger) SetLevel(level LogLevel) {
	sl.minLevel = level
	
	// Update the slog handler's level with the original output
	handler := slog.NewJSONHandler(sl.output, &slog.HandlerOptions{
		Level:     level.ToSlogLevel(),
		AddSource: false,
	})
	sl.slogger = slog.New(handler)
}

// IsEnabled checks if a log level is enabled
func (sl *StructuredLogger) IsEnabled(level LogLevel) bool {
	return level >= sl.minLevel
}

// LogProcessingStart logs the start of record processing
func LogProcessingStart(logger Logger, recordCount int) {
	logger.WithProcessingState("start").Info("Starting record processing",
		"record_count", recordCount,
		"operation", "process_records",
	)
}

// LogProcessingComplete logs the completion of record processing
func LogProcessingComplete(logger Logger, recordCount int, duration time.Duration) {
	logger.WithProcessingState("complete").Info("Record processing completed",
		"record_count", recordCount,
		"duration_ms", duration.Milliseconds(),
		"operation", "process_records",
	)
}

// LogValidationError logs validation errors with detailed context
func LogValidationError(logger Logger, recordIndex int, errorType string, message string) {
	logger.WithProcessingState("validation_error").Error("Record validation failed",
		"record_index", recordIndex,
		"error_type", errorType,
		"validation_message", message,
		"operation", "validate_record",
	)
}

// LogParsingError logs parsing errors with detailed context
func LogParsingError(logger Logger, lineNumber int, errorType string, input string) {
	logger.WithProcessingState("parsing_error").Error("Line parsing failed",
		"line_number", lineNumber,
		"error_type", errorType,
		"input_data", input,
		"operation", "parse_line",
	)
}

// LogCalculationResult logs fare calculation results
func LogCalculationResult(logger Logger, totalFare interface{}, recordCount int) {
	logger.WithProcessingState("calculation_complete").Info("Fare calculation completed",
		"total_fare", totalFare,
		"record_count", recordCount,
		"operation", "calculate_fare",
	)
}

// Performance optimizations and utilities

// BufferedLogger wraps a logger with buffering for high-performance scenarios
type BufferedLogger struct {
	underlying Logger
	buffer     []LogEntry
	maxBuffer  int
}

// NewBufferedLogger creates a buffered logger for high-volume logging scenarios
func NewBufferedLogger(underlying Logger, maxBuffer int) *BufferedLogger {
	return &BufferedLogger{
		underlying: underlying,
		buffer:     make([]LogEntry, 0, maxBuffer),
		maxBuffer:  maxBuffer,
	}
}

// Flush flushes any buffered log entries to the underlying logger
func (bl *BufferedLogger) Flush() {
	for _, entry := range bl.buffer {
		bl.underlying.LogWithLevel(
			parseLogLevel(entry.Level),
			entry.Message,
			contextToInterfaceSlice(entry.Context)...,
		)
	}
	bl.buffer = bl.buffer[:0] // Reset buffer
}

// parseLogLevel parses a string log level back to LogLevel
func parseLogLevel(levelStr string) LogLevel {
	switch levelStr {
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARN":
		return LevelWarn
	case "ERROR":
		return LevelError
	default:
		return LevelInfo
	}
}

// contextToInterfaceSlice converts a context map to interface slice for logging
func contextToInterfaceSlice(context map[string]interface{}) []interface{} {
	result := make([]interface{}, 0, len(context)*2)
	for k, v := range context {
		result = append(result, k, v)
	}
	return result
}