package loggingsystem

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		expected string
	}{
		{"debug", LevelDebug, "DEBUG"},
		{"info", LevelInfo, "INFO"},
		{"warn", LevelWarn, "WARN"},
		{"error", LevelError, "ERROR"},
		{"unknown", LogLevel(999), "UNKNOWN"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("LogLevel.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestLogLevel_ToSlogLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		expected string // We'll check the string representation since slog.Level doesn't have direct equality
	}{
		{"debug", LevelDebug, "DEBUG"},
		{"info", LevelInfo, "INFO"},
		{"warn", LevelWarn, "WARN"},
		{"error", LevelError, "ERROR"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slogLevel := tt.level.ToSlogLevel()
			if slogLevel.String() != tt.expected {
				t.Errorf("LogLevel.ToSlogLevel() = %v, want %v", slogLevel.String(), tt.expected)
			}
		})
	}
}

func TestNewLogger(t *testing.T) {
	logger := NewLogger()
	if logger == nil {
		t.Error("Expected non-nil logger")
	}
	
	// Test that it implements the Logger interface
	_, ok := logger.(Logger)
	if !ok {
		t.Error("Logger should implement Logger interface")
	}
}

func TestNewLoggerWithOptions(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithOptions(&buf, LevelDebug)
	
	if logger == nil {
		t.Error("Expected non-nil logger")
	}
	
	// Test that debug messages are logged
	logger.Debug("test debug message")
	
	output := buf.String()
	if output == "" {
		t.Error("Expected some output for debug message")
	}
	
	// Verify it's valid JSON
	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logData); err != nil {
		t.Errorf("Expected valid JSON output, got error: %v", err)
	}
}

func TestStructuredLogger_BasicLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithOptions(&buf, LevelDebug).(*StructuredLogger)
	
	tests := []struct {
		name     string
		logFunc  func(string, ...interface{})
		message  string
		level    string
	}{
		{"debug", logger.Debug, "debug message", "DEBUG"},
		{"info", logger.Info, "info message", "INFO"},
		{"warn", logger.Warn, "warn message", "WARN"},
		{"error", logger.Error, "error message", "ERROR"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc(tt.message)
			
			output := buf.String()
			if output == "" {
				t.Error("Expected output for log message")
				return
			}
			
			// Parse JSON output
			var logData map[string]interface{}
			if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logData); err != nil {
				t.Errorf("Expected valid JSON output, got error: %v", err)
				return
			}
			
			// Check required fields
			if logData["level"] != tt.level {
				t.Errorf("Expected level %s, got %v", tt.level, logData["level"])
			}
			
			if logData["msg"] != tt.message {
				t.Errorf("Expected message %s, got %v", tt.message, logData["msg"])
			}
			
			// Check timestamp exists
			if _, exists := logData["time"]; !exists {
				t.Error("Expected timestamp in log output")
			}
		})
	}
}

func TestStructuredLogger_LogWithLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithOptions(&buf, LevelDebug).(*StructuredLogger)
	
	logger.LogWithLevel(LevelInfo, "test message", "key1", "value1", "key2", 42)
	
	output := buf.String()
	if output == "" {
		t.Error("Expected output for log message")
		return
	}
	
	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logData); err != nil {
		t.Errorf("Expected valid JSON output, got error: %v", err)
		return
	}
	
	// Check context fields
	if logData["key1"] != "value1" {
		t.Errorf("Expected key1=value1, got %v", logData["key1"])
	}
	
	if logData["key2"] != float64(42) { // JSON numbers are float64
		t.Errorf("Expected key2=42, got %v", logData["key2"])
	}
}

func TestStructuredLogger_WithContext(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithOptions(&buf, LevelDebug)
	
	contextLogger := logger.WithContext(map[string]interface{}{
		"user_id": "12345",
		"session": "abc-def",
	})
	
	contextLogger.Info("test message")
	
	output := buf.String()
	if output == "" {
		t.Error("Expected output for log message")
		return
	}
	
	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logData); err != nil {
		t.Errorf("Expected valid JSON output, got error: %v", err)
		return
	}
	
	// Check context fields
	if logData["user_id"] != "12345" {
		t.Errorf("Expected user_id=12345, got %v", logData["user_id"])
	}
	
	if logData["session"] != "abc-def" {
		t.Errorf("Expected session=abc-def, got %v", logData["session"])
	}
}

func TestStructuredLogger_WithComponent(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithOptions(&buf, LevelDebug)
	
	componentLogger := logger.WithComponent("parser")
	componentLogger.Info("parsing started")
	
	output := buf.String()
	if output == "" {
		t.Error("Expected output for log message")
		return
	}
	
	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logData); err != nil {
		t.Errorf("Expected valid JSON output, got error: %v", err)
		return
	}
	
	if logData["component"] != "parser" {
		t.Errorf("Expected component=parser, got %v", logData["component"])
	}
}

func TestStructuredLogger_WithRecordID(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithOptions(&buf, LevelDebug)
	
	recordLogger := logger.WithRecordID("record-001")
	recordLogger.Info("processing record")
	
	output := buf.String()
	if output == "" {
		t.Error("Expected output for log message")
		return
	}
	
	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logData); err != nil {
		t.Errorf("Expected valid JSON output, got error: %v", err)
		return
	}
	
	if logData["record_id"] != "record-001" {
		t.Errorf("Expected record_id=record-001, got %v", logData["record_id"])
	}
}

func TestStructuredLogger_WithProcessingState(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithOptions(&buf, LevelDebug)
	
	stateLogger := logger.WithProcessingState("validating")
	stateLogger.Info("validation in progress")
	
	output := buf.String()
	if output == "" {
		t.Error("Expected output for log message")
		return
	}
	
	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logData); err != nil {
		t.Errorf("Expected valid JSON output, got error: %v", err)
		return
	}
	
	if logData["processing_state"] != "validating" {
		t.Errorf("Expected processing_state=validating, got %v", logData["processing_state"])
	}
}

func TestStructuredLogger_CombinedContext(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithOptions(&buf, LevelDebug)
	
	combinedLogger := logger.
		WithComponent("validator").
		WithRecordID("rec-123").
		WithProcessingState("checking").
		WithContext(map[string]interface{}{"rule": "timing"})
	
	combinedLogger.Warn("validation warning", "details", "out of range")
	
	output := buf.String()
	if output == "" {
		t.Error("Expected output for log message")
		return
	}
	
	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logData); err != nil {
		t.Errorf("Expected valid JSON output, got error: %v", err)
		return
	}
	
	// Check all context fields
	expectedFields := map[string]interface{}{
		"component":        "validator",
		"record_id":        "rec-123",
		"processing_state": "checking",
		"rule":             "timing",
		"details":          "out of range",
		"level":            "WARN",
		"msg":              "validation warning",
	}
	
	for key, expectedValue := range expectedFields {
		if logData[key] != expectedValue {
			t.Errorf("Expected %s=%v, got %v", key, expectedValue, logData[key])
		}
	}
}

func TestStructuredLogger_SetLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithOptions(&buf, LevelInfo).(*StructuredLogger)
	
	// Debug should not be logged initially
	buf.Reset()
	logger.Debug("debug message")
	if buf.String() != "" {
		t.Error("Debug message should not be logged when level is INFO")
	}
	
	// Set level to Debug
	logger.SetLevel(LevelDebug)
	
	// Debug should now be logged
	buf.Reset()
	logger.Debug("debug message")
	if buf.String() == "" {
		t.Error("Debug message should be logged when level is DEBUG")
	}
}

func TestStructuredLogger_IsEnabled(t *testing.T) {
	logger := NewLoggerWithOptions(&bytes.Buffer{}, LevelWarn).(*StructuredLogger)
	
	tests := []struct {
		level    LogLevel
		expected bool
	}{
		{LevelDebug, false},
		{LevelInfo, false},
		{LevelWarn, true},
		{LevelError, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			if got := logger.IsEnabled(tt.level); got != tt.expected {
				t.Errorf("IsEnabled(%v) = %v, want %v", tt.level, got, tt.expected)
			}
		})
	}
}

func TestLogProcessingStart(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithOptions(&buf, LevelDebug)
	
	LogProcessingStart(logger, 100)
	
	output := buf.String()
	if output == "" {
		t.Error("Expected output for processing start log")
		return
	}
	
	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logData); err != nil {
		t.Errorf("Expected valid JSON output, got error: %v", err)
		return
	}
	
	expectedFields := map[string]interface{}{
		"processing_state": "start",
		"record_count":     float64(100),
		"operation":        "process_records",
		"level":            "INFO",
	}
	
	for key, expectedValue := range expectedFields {
		if logData[key] != expectedValue {
			t.Errorf("Expected %s=%v, got %v", key, expectedValue, logData[key])
		}
	}
}

func TestLogProcessingComplete(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithOptions(&buf, LevelDebug)
	
	LogProcessingComplete(logger, 100, 250*time.Millisecond)
	
	output := buf.String()
	if output == "" {
		t.Error("Expected output for processing complete log")
		return
	}
	
	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logData); err != nil {
		t.Errorf("Expected valid JSON output, got error: %v", err)
		return
	}
	
	expectedFields := map[string]interface{}{
		"processing_state": "complete",
		"record_count":     float64(100),
		"duration_ms":      float64(250),
		"operation":        "process_records",
		"level":            "INFO",
	}
	
	for key, expectedValue := range expectedFields {
		if logData[key] != expectedValue {
			t.Errorf("Expected %s=%v, got %v", key, expectedValue, logData[key])
		}
	}
}

func TestLogValidationError(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithOptions(&buf, LevelDebug)
	
	LogValidationError(logger, 5, "timing", "timestamp out of sequence")
	
	output := buf.String()
	if output == "" {
		t.Error("Expected output for validation error log")
		return
	}
	
	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logData); err != nil {
		t.Errorf("Expected valid JSON output, got error: %v", err)
		return
	}
	
	expectedFields := map[string]interface{}{
		"processing_state":   "validation_error",
		"record_index":       float64(5),
		"error_type":         "timing",
		"validation_message": "timestamp out of sequence",
		"operation":          "validate_record",
		"level":              "ERROR",
	}
	
	for key, expectedValue := range expectedFields {
		if logData[key] != expectedValue {
			t.Errorf("Expected %s=%v, got %v", key, expectedValue, logData[key])
		}
	}
}

func TestLogParsingError(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithOptions(&buf, LevelDebug)
	
	LogParsingError(logger, 10, "format", "12:30:45 invalid_distance")
	
	output := buf.String()
	if output == "" {
		t.Error("Expected output for parsing error log")
		return
	}
	
	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logData); err != nil {
		t.Errorf("Expected valid JSON output, got error: %v", err)
		return
	}
	
	expectedFields := map[string]interface{}{
		"processing_state": "parsing_error",
		"line_number":      float64(10),
		"error_type":       "format",
		"input_data":       "12:30:45 invalid_distance",
		"operation":        "parse_line",
		"level":            "ERROR",
	}
	
	for key, expectedValue := range expectedFields {
		if logData[key] != expectedValue {
			t.Errorf("Expected %s=%v, got %v", key, expectedValue, logData[key])
		}
	}
}

func TestLogCalculationResult(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithOptions(&buf, LevelDebug)
	
	LogCalculationResult(logger, "1250", 25)
	
	output := buf.String()
	if output == "" {
		t.Error("Expected output for calculation result log")
		return
	}
	
	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logData); err != nil {
		t.Errorf("Expected valid JSON output, got error: %v", err)
		return
	}
	
	expectedFields := map[string]interface{}{
		"processing_state": "calculation_complete",
		"total_fare":       "1250",
		"record_count":     float64(25),
		"operation":        "calculate_fare",
		"level":            "INFO",
	}
	
	for key, expectedValue := range expectedFields {
		if logData[key] != expectedValue {
			t.Errorf("Expected %s=%v, got %v", key, expectedValue, logData[key])
		}
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		levelStr string
		expected LogLevel
	}{
		{"DEBUG", LevelDebug},
		{"INFO", LevelInfo},
		{"WARN", LevelWarn},
		{"ERROR", LevelError},
		{"UNKNOWN", LevelInfo}, // Default fallback
		{"", LevelInfo},        // Default fallback
	}
	
	for _, tt := range tests {
		t.Run(tt.levelStr, func(t *testing.T) {
			got := parseLogLevel(tt.levelStr)
			if got != tt.expected {
				t.Errorf("parseLogLevel(%q) = %v, want %v", tt.levelStr, got, tt.expected)
			}
		})
	}
}

func TestContextToInterfaceSlice(t *testing.T) {
	context := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}
	
	result := contextToInterfaceSlice(context)
	
	// Should have 6 elements (3 key-value pairs)
	if len(result) != 6 {
		t.Errorf("Expected 6 elements, got %d", len(result))
	}
	
	// Convert back to map to verify content
	resultMap := make(map[string]interface{})
	for i := 0; i < len(result); i += 2 {
		if i+1 < len(result) {
			if key, ok := result[i].(string); ok {
				resultMap[key] = result[i+1]
			}
		}
	}
	
	for key, expectedValue := range context {
		if resultMap[key] != expectedValue {
			t.Errorf("Expected %s=%v, got %v", key, expectedValue, resultMap[key])
		}
	}
}

// Benchmark tests for performance validation
func BenchmarkStructuredLogger_Info(b *testing.B) {
	var buf bytes.Buffer
	logger := NewLoggerWithOptions(&buf, LevelInfo)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		logger.Info("benchmark message", "iteration", i)
	}
}

func BenchmarkStructuredLogger_InfoWithContext(b *testing.B) {
	var buf bytes.Buffer
	logger := NewLoggerWithOptions(&buf, LevelInfo).
		WithComponent("benchmark").
		WithRecordID("bench-001")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		logger.Info("benchmark message", "iteration", i, "data", "test")
	}
}

func BenchmarkLogProcessingStart(b *testing.B) {
	var buf bytes.Buffer
	logger := NewLoggerWithOptions(&buf, LevelInfo)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		LogProcessingStart(logger, 100)
	}
}