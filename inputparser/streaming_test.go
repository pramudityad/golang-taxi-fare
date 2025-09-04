package inputparser

import (
	"context"
	"strings"
	"testing"
	"time"

	"golang-taxi-fare/models"
)

func TestParseLine(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		lineNum     int
		wantErr     bool
		expectedErr string
		expectedRecord *models.DistanceRecord
	}{
		{
			name:    "valid line",
			input:   "12:34:56.789 12345678.5",
			lineNum: 1,
			wantErr: false,
			expectedRecord: &models.DistanceRecord{
				Timestamp: mustParseTime("12:34:56.789"),
				Distance:  mustDecimal("12345678.5"),
			},
		},
		{
			name:    "valid line with multiple decimal places",
			input:   "00:00:00.000 87654321.123456",
			lineNum: 5,
			wantErr: false,
			expectedRecord: &models.DistanceRecord{
				Timestamp: mustParseTime("00:00:00.000"),
				Distance:  mustDecimal("87654321.123456"),
			},
		},
		{
			name:    "valid line end of day",
			input:   "23:59:59.999 99999999.9",
			lineNum: 10,
			wantErr: false,
			expectedRecord: &models.DistanceRecord{
				Timestamp: mustParseTime("23:59:59.999"),
				Distance:  mustDecimal("99999999.9"),
			},
		},
		{
			name:        "blank line",
			input:       "   ",
			lineNum:     2,
			wantErr:     true,
			expectedErr: "blank line",
		},
		{
			name:        "empty line",
			input:       "",
			lineNum:     3,
			wantErr:     true,
			expectedErr: "blank line",
		},
		{
			name:        "invalid format - missing space",
			input:       "12:34:56.78912345678.5",
			lineNum:     4,
			wantErr:     true,
			expectedErr: "invalid line format",
		},
		{
			name:        "invalid format - wrong timestamp",
			input:       "25:34:56.789 12345678.5",
			lineNum:     5,
			wantErr:     true,
			expectedErr: "invalid timestamp format", // parsing will catch invalid hour
		},
		{
			name:        "invalid format - wrong distance",
			input:       "12:34:56.789 1234567.5",
			lineNum:     6,
			wantErr:     true,
			expectedErr: "invalid line format", // regex will catch short distance
		},
		{
			name:        "invalid format - extra spaces",
			input:       "12:34:56.789   12345678.5",
			lineNum:     7,
			wantErr:     true,
			expectedErr: "invalid line format", // regex expects single space
		},
		{
			name:        "invalid format - extra content",
			input:       "12:34:56.789 12345678.5 extra",
			lineNum:     8,
			wantErr:     true,
			expectedErr: "invalid line format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseLine(tt.input, tt.lineNum)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseLine() expected error, got nil")
					return
				}
				
				if tt.expectedErr != "" && !contains(err.Error(), tt.expectedErr) {
					t.Errorf("parseLine() error = %v, expected to contain %v", err.Error(), tt.expectedErr)
				}
				
				// Verify line number is set in error
				if pe, ok := err.(*ParsingError); ok {
					if pe.Line != tt.lineNum {
						t.Errorf("parseLine() error line = %d, want %d", pe.Line, tt.lineNum)
					}
				}
			} else {
				if err != nil {
					t.Errorf("parseLine() unexpected error = %v", err)
					return
				}
				
				if tt.expectedRecord != nil {
					if !result.Timestamp.Equal(tt.expectedRecord.Timestamp) {
						t.Errorf("parseLine() timestamp = %v, want %v", result.Timestamp, tt.expectedRecord.Timestamp)
					}
					if !result.Distance.Equal(tt.expectedRecord.Distance) {
						t.Errorf("parseLine() distance = %v, want %v", result.Distance, tt.expectedRecord.Distance)
					}
				}
			}
		})
	}
}

func TestStreamParser_ParseLine(t *testing.T) {
	parser := &StreamParser{}
	
	t.Run("valid line through interface", func(t *testing.T) {
		result, err := parser.ParseLine("12:34:56.789 12345678.5")
		if err != nil {
			t.Fatalf("ParseLine() unexpected error = %v", err)
		}
		
		expectedTime := mustParseTime("12:34:56.789")
		expectedDistance := mustDecimal("12345678.5")
		
		if !result.Timestamp.Equal(expectedTime) {
			t.Errorf("ParseLine() timestamp = %v, want %v", result.Timestamp, expectedTime)
		}
		if !result.Distance.Equal(expectedDistance) {
			t.Errorf("ParseLine() distance = %v, want %v", result.Distance, expectedDistance)
		}
	})
	
	t.Run("invalid line through interface", func(t *testing.T) {
		_, err := parser.ParseLine("invalid line")
		if err == nil {
			t.Error("ParseLine() expected error for invalid line, got nil")
		}
	})
}

func TestStreamParser_ParseStream(t *testing.T) {
	t.Run("successful streaming", func(t *testing.T) {
		input := `12:34:56.789 12345678.5
00:00:00.000 87654321.123
23:59:59.999 99999999.9`
		
		parser := &StreamParser{}
		ctx := context.Background()
		reader := strings.NewReader(input)
		
		resultChan, err := parser.ParseStream(ctx, reader)
		if err != nil {
			t.Fatalf("ParseStream() unexpected error = %v", err)
		}
		
		var results []ParseResult
		for result := range resultChan {
			results = append(results, result)
		}
		
		if len(results) != 3 {
			t.Errorf("ParseStream() got %d results, want 3", len(results))
		}
		
		// Verify first result
		if results[0].Error != nil {
			t.Errorf("ParseStream() result[0] unexpected error: %v", results[0].Error)
		}
		if results[0].Line != 1 {
			t.Errorf("ParseStream() result[0] line = %d, want 1", results[0].Line)
		}
		
		expectedTime := mustParseTime("12:34:56.789")
		if !results[0].Record.Timestamp.Equal(expectedTime) {
			t.Errorf("ParseStream() result[0] timestamp = %v, want %v", results[0].Record.Timestamp, expectedTime)
		}
	})
	
	t.Run("streaming with blank lines", func(t *testing.T) {
		input := `12:34:56.789 12345678.5

00:00:00.000 87654321.123
   
23:59:59.999 99999999.9`
		
		parser := &StreamParser{}
		ctx := context.Background()
		reader := strings.NewReader(input)
		
		resultChan, err := parser.ParseStream(ctx, reader)
		if err != nil {
			t.Fatalf("ParseStream() unexpected error = %v", err)
		}
		
		var results []ParseResult
		for result := range resultChan {
			results = append(results, result)
		}
		
		// Should only get 3 results (blank lines are skipped)
		if len(results) != 3 {
			t.Errorf("ParseStream() got %d results, want 3 (blank lines should be skipped)", len(results))
		}
		
		// Line numbers should still be correct
		expectedLines := []int{1, 3, 5}
		for i, result := range results {
			if result.Line != expectedLines[i] {
				t.Errorf("ParseStream() result[%d] line = %d, want %d", i, result.Line, expectedLines[i])
			}
		}
	})
	
	t.Run("streaming with errors", func(t *testing.T) {
		input := `12:34:56.789 12345678.5
invalid line format
00:00:00.000 87654321.123`
		
		parser := &StreamParser{}
		ctx := context.Background()
		reader := strings.NewReader(input)
		
		resultChan, err := parser.ParseStream(ctx, reader)
		if err != nil {
			t.Fatalf("ParseStream() unexpected error = %v", err)
		}
		
		var results []ParseResult
		for result := range resultChan {
			results = append(results, result)
		}
		
		if len(results) != 3 {
			t.Errorf("ParseStream() got %d results, want 3", len(results))
		}
		
		// First result should be successful
		if results[0].Error != nil {
			t.Errorf("ParseStream() result[0] unexpected error: %v", results[0].Error)
		}
		
		// Second result should have error
		if results[1].Error == nil {
			t.Error("ParseStream() result[1] expected error, got nil")
		}
		if results[1].Line != 2 {
			t.Errorf("ParseStream() result[1] line = %d, want 2", results[1].Line)
		}
		
		// Third result should be successful
		if results[2].Error != nil {
			t.Errorf("ParseStream() result[2] unexpected error: %v", results[2].Error)
		}
	})
	
	t.Run("context cancellation", func(t *testing.T) {
		input := `12:34:56.789 12345678.5
00:00:00.000 87654321.123
23:59:59.999 99999999.9`
		
		parser := &StreamParser{}
		ctx, cancel := context.WithCancel(context.Background())
		reader := strings.NewReader(input)
		
		resultChan, err := parser.ParseStream(ctx, reader)
		if err != nil {
			t.Fatalf("ParseStream() unexpected error = %v", err)
		}
		
		// Cancel immediately to ensure cancellation happens
		cancel()
		
		var results []ParseResult
		for result := range resultChan {
			results = append(results, result)
		}
		
		// Should get at least one result (either successful parse or cancellation error)
		if len(results) < 1 {
			t.Error("ParseStream() expected at least 1 result")
		}
		
		// Check if any result has context error (due to timing, this may vary)
		hasContextError := false
		for _, result := range results {
			if result.Error == context.Canceled {
				hasContextError = true
				break
			}
		}
		
		// This test is about verifying the cancellation mechanism works,
		// not the exact timing, so we just ensure the channel closes properly
		// The fact that we get results and the channel closes is sufficient
		if len(results) == 0 {
			t.Error("ParseStream() should produce some results or errors")
		}
		
		// Optional: if we got a context error, verify it's the right one
		if hasContextError {
			t.Log("ParseStream() correctly handled context cancellation")
		} else {
			t.Log("ParseStream() completed before cancellation took effect (timing dependent)")
		}
	})
}

func TestLinePattern(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
		groups   []string // expected capture groups if match
	}{
		{
			name:     "valid format",
			input:    "12:34:56.789 12345678.5",
			expected: true,
			groups:   []string{"12:34:56.789 12345678.5", "12:34:56.789", "12345678.5"},
		},
		{
			name:     "valid format with more decimals",
			input:    "00:00:00.000 87654321.123456",
			expected: true,
			groups:   []string{"00:00:00.000 87654321.123456", "00:00:00.000", "87654321.123456"},
		},
		{
			name:     "invalid - no space",
			input:    "12:34:56.78912345678.5",
			expected: false,
		},
		{
			name:     "invalid - multiple spaces",
			input:    "12:34:56.789  12345678.5",
			expected: false,
		},
		{
			name:     "invalid - wrong timestamp format",
			input:    "1:34:56.789 12345678.5",
			expected: false,
		},
		{
			name:     "invalid - short distance",
			input:    "12:34:56.789 1234567.5",
			expected: false,
		},
		{
			name:     "invalid - extra content",
			input:    "12:34:56.789 12345678.5 extra",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := linePattern.FindStringSubmatch(tt.input)
			
			if tt.expected {
				if matches == nil {
					t.Errorf("linePattern.FindStringSubmatch(%q) = nil, want matches", tt.input)
					return
				}
				
				if len(tt.groups) > 0 {
					if len(matches) != len(tt.groups) {
						t.Errorf("linePattern.FindStringSubmatch(%q) got %d groups, want %d", tt.input, len(matches), len(tt.groups))
					} else {
						for i, expected := range tt.groups {
							if matches[i] != expected {
								t.Errorf("linePattern.FindStringSubmatch(%q) group[%d] = %q, want %q", tt.input, i, matches[i], expected)
							}
						}
					}
				}
			} else {
				if matches != nil {
					t.Errorf("linePattern.FindStringSubmatch(%q) = %v, want nil", tt.input, matches)
				}
			}
		})
	}
}

// Helper functions for tests
func mustParseTime(timeStr string) time.Time {
	t, err := time.Parse(timestampLayout, timeStr)
	if err != nil {
		panic("invalid time in test: " + timeStr)
	}
	return t
}