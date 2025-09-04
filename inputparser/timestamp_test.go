package inputparser

import (
	"testing"
	"time"
)

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		expectedErr string
	}{
		{
			name:    "valid timestamp with milliseconds",
			input:   "12:34:56.789",
			wantErr: false,
		},
		{
			name:    "valid timestamp midnight",
			input:   "00:00:00.000",
			wantErr: false,
		},
		{
			name:    "valid timestamp end of day",
			input:   "23:59:59.999",
			wantErr: false,
		},
		{
			name:        "empty timestamp",
			input:       "",
			wantErr:     true,
			expectedErr: "empty timestamp",
		},
		{
			name:        "invalid format - missing milliseconds",
			input:       "12:34:56",
			wantErr:     true,
			expectedErr: "invalid timestamp format",
		},
		{
			name:        "invalid format - wrong separator",
			input:       "12-34-56.789",
			wantErr:     true,
			expectedErr: "invalid timestamp format",
		},
		{
			name:        "invalid hour",
			input:       "25:34:56.789",
			wantErr:     true,
			expectedErr: "invalid timestamp format",
		},
		{
			name:        "invalid minute",
			input:       "12:67:56.789",
			wantErr:     true,
			expectedErr: "invalid timestamp format",
		},
		{
			name:        "invalid second",
			input:       "12:34:67.789",
			wantErr:     true,
			expectedErr: "invalid timestamp format",
		},
		{
			name:        "non-numeric characters",
			input:       "ab:cd:ef.ghi",
			wantErr:     true,
			expectedErr: "invalid timestamp format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseTimestamp(tt.input)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseTimestamp() expected error, got nil")
					return
				}
				
				if tt.expectedErr != "" && !contains(err.Error(), tt.expectedErr) {
					t.Errorf("parseTimestamp() error = %v, expected to contain %v", err.Error(), tt.expectedErr)
				}
				
				// Verify it's a ParsingError with correct type
				if pe, ok := err.(*ParsingError); ok {
					if pe.Type != ErrorTypeTimestamp {
						t.Errorf("parseTimestamp() error type = %v, want %v", pe.Type, ErrorTypeTimestamp)
					}
					if pe.Input != tt.input {
						t.Errorf("parseTimestamp() error input = %v, want %v", pe.Input, tt.input)
					}
				} else {
					t.Errorf("parseTimestamp() error is not ParsingError type")
				}
			} else {
				if err != nil {
					t.Errorf("parseTimestamp() unexpected error = %v", err)
					return
				}
				
				// Verify the parsed time has correct components
				expected, _ := time.Parse(timestampLayout, tt.input)
				if !result.Equal(expected) {
					t.Errorf("parseTimestamp() = %v, want %v", result, expected)
				}
			}
		})
	}
}

func TestValidateTimestampFormat(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		expectedErr string
	}{
		{
			name:    "valid format",
			input:   "12:34:56.789",
			wantErr: false,
		},
		{
			name:        "too short",
			input:       "12:34:56",
			wantErr:     true,
			expectedErr: "invalid timestamp length",
		},
		{
			name:        "too long",
			input:       "12:34:56.7890",
			wantErr:     true,
			expectedErr: "invalid timestamp length",
		},
		{
			name:        "missing first colon",
			input:       "12334:56.789",
			wantErr:     true,
			expectedErr: "missing colon separator at position 2",
		},
		{
			name:        "missing second colon",
			input:       "12:34567.789",
			wantErr:     true,
			expectedErr: "missing colon separator at position 5",
		},
		{
			name:        "missing dot",
			input:       "12:34:567789",
			wantErr:     true,
			expectedErr: "missing dot separator at position 8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTimestampFormat(tt.input)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateTimestampFormat() expected error, got nil")
					return
				}
				
				if tt.expectedErr != "" && !contains(err.Error(), tt.expectedErr) {
					t.Errorf("validateTimestampFormat() error = %v, expected to contain %v", err.Error(), tt.expectedErr)
				}
			} else {
				if err != nil {
					t.Errorf("validateTimestampFormat() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestParseTimestampWithValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid timestamp passes both validation and parsing",
			input:   "12:34:56.789",
			wantErr: false,
		},
		{
			name:    "boundary case - midnight",
			input:   "00:00:00.000",
			wantErr: false,
		},
		{
			name:    "boundary case - end of day",
			input:   "23:59:59.999",
			wantErr: false,
		},
		{
			name:    "fails validation - wrong length",
			input:   "12:34:56",
			wantErr: true,
		},
		{
			name:    "fails parsing - invalid hour",
			input:   "25:34:56.789",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseTimestampWithValidation(tt.input)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseTimestampWithValidation() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("parseTimestampWithValidation() unexpected error = %v", err)
					return
				}
				
				// Verify the result is not zero time
				if result.IsZero() {
					t.Errorf("parseTimestampWithValidation() returned zero time")
				}
				
				// Verify millisecond precision is maintained
				expected, _ := time.Parse(timestampLayout, tt.input)
				if !result.Equal(expected) {
					t.Errorf("parseTimestampWithValidation() = %v, want %v", result, expected)
				}
			}
		})
	}
}

func TestTimestampLayout(t *testing.T) {
	// Test that our layout constant is correct
	testTime := "14:25:36.123"
	parsed, err := time.Parse(timestampLayout, testTime)
	if err != nil {
		t.Fatalf("timestampLayout is invalid: %v", err)
	}
	
	// Verify the parsed components
	if parsed.Hour() != 14 {
		t.Errorf("Expected hour 14, got %d", parsed.Hour())
	}
	if parsed.Minute() != 25 {
		t.Errorf("Expected minute 25, got %d", parsed.Minute())
	}
	if parsed.Second() != 36 {
		t.Errorf("Expected second 36, got %d", parsed.Second())
	}
	if parsed.Nanosecond() != 123000000 {
		t.Errorf("Expected nanosecond 123000000 (123ms), got %d", parsed.Nanosecond())
	}
}

// Helper function for string contains check (reused from parser_test.go)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		 s[len(s)-len(substr):] == substr || 
		 containsInner(s, substr))))
}

func containsInner(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}