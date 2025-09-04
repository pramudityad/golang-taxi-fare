package inputparser

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"golang-taxi-fare/models"
)

func TestNewParser(t *testing.T) {
	parser := NewParser()
	if parser == nil {
		t.Fatal("NewParser() returned nil")
	}
	
	// Verify it implements the Parser interface
	var _ Parser = parser
}

func TestParsingError(t *testing.T) {
	t.Run("Error method", func(t *testing.T) {
		err := &ParsingError{
			Type:    ErrorTypeFormat,
			Message: "invalid format",
			Line:    5,
			Input:   "bad input",
		}
		
		expected := `parsing error at line 5: invalid format (input: "bad input")`
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})
}

func TestErrorType_String(t *testing.T) {
	tests := []struct {
		errorType ErrorType
		expected  string
	}{
		{ErrorTypeFormat, "format"},
		{ErrorTypeTimestamp, "timestamp"},
		{ErrorTypeDistance, "distance"},
		{ErrorTypeIO, "io"},
		{ErrorType(999), "unknown"},
	}
	
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.errorType.String(); got != tt.expected {
				t.Errorf("ErrorType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseResult(t *testing.T) {
	t.Run("Valid ParseResult structure", func(t *testing.T) {
		record := models.DistanceRecord{
			Timestamp: time.Now(),
			Distance:  mustDecimal("12.5"),
		}
		
		result := ParseResult{
			Record: record,
			Error:  nil,
			Line:   1,
		}
		
		if result.Error != nil {
			t.Errorf("Expected no error, got %v", result.Error)
		}
		if result.Line != 1 {
			t.Errorf("Expected line 1, got %d", result.Line)
		}
		if !result.Record.Distance.Equal(mustDecimal("12.5")) {
			t.Errorf("Expected distance 12.5, got %s", result.Record.Distance)
		}
	})
	
	t.Run("ParseResult with error", func(t *testing.T) {
		err := &ParsingError{
			Type:    ErrorTypeFormat,
			Message: "test error",
			Line:    3,
			Input:   "invalid",
		}
		
		result := ParseResult{
			Record: models.DistanceRecord{},
			Error:  err,
			Line:   3,
		}
		
		if result.Error == nil {
			t.Error("Expected error, got nil")
		}
		if result.Line != 3 {
			t.Errorf("Expected line 3, got %d", result.Line)
		}
	})
}

func TestStreamParser_Interface(t *testing.T) {
	t.Run("StreamParser implements Parser interface", func(t *testing.T) {
		parser := &StreamParser{}
		var _ Parser = parser
	})
}

func TestStreamParser_ParseStream_Implemented(t *testing.T) {
	parser := &StreamParser{}
	ctx := context.Background()
	reader := strings.NewReader("12:34:56.789 12345678.5")
	
	channel, err := parser.ParseStream(ctx, reader)
	
	if err != nil {
		t.Errorf("ParseStream() unexpected error = %v", err)
	}
	if channel == nil {
		t.Error("ParseStream() returned nil channel")
	}
	
	// Read one result to verify it works
	if channel != nil {
		results := make([]ParseResult, 0)
		for result := range channel {
			results = append(results, result)
		}
		
		if len(results) != 1 {
			t.Errorf("ParseStream() got %d results, want 1", len(results))
		}
		
		if len(results) > 0 && results[0].Error != nil {
			t.Errorf("ParseStream() result error = %v, want nil", results[0].Error)
		}
	}
}

func TestStreamParser_ParseLine_Implemented(t *testing.T) {
	parser := &StreamParser{}
	
	record, err := parser.ParseLine("12:34:56.789 12345678.5")
	
	if err != nil {
		t.Errorf("ParseLine() unexpected error = %v", err)
	}
	
	// Check that we get proper values
	if record.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp, got zero")
	}
	if record.Distance.IsZero() {
		t.Error("Expected non-zero distance, got zero")
	}
	
	// Test error case
	_, err = parser.ParseLine("invalid line")
	if err == nil {
		t.Error("ParseLine() expected error for invalid line, got nil")
	}
}

// Helper function for creating decimal values in tests
func mustDecimal(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		panic("invalid decimal in test: " + s)
	}
	return d
}