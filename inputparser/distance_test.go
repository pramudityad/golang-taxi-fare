package inputparser

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestParseDistance(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		expectedErr string
		expected    string // decimal value as string for comparison
	}{
		{
			name:     "valid distance with single decimal",
			input:    "12345678.5",
			wantErr:  false,
			expected: "12345678.5",
		},
		{
			name:     "valid distance with multiple decimals",
			input:    "87654321.123",
			wantErr:  false,
			expected: "87654321.123",
		},
		{
			name:     "valid distance with zero",
			input:    "00000000.0",
			wantErr:  false,
			expected: "0",
		},
		{
			name:     "valid large distance",
			input:    "999999999.999999",
			wantErr:  false,
			expected: "999999999.999999",
		},
		{
			name:        "empty distance",
			input:       "",
			wantErr:     true,
			expectedErr: "empty distance",
		},
		{
			name:        "negative distance",
			input:       "-12345678.5",
			wantErr:     true,
			expectedErr: "distance cannot be negative",
		},
		{
			name:        "invalid decimal format",
			input:       "not_a_number",
			wantErr:     true,
			expectedErr: "invalid distance format",
		},
		{
			name:        "no decimal point (accepted by parseDistance, rejected by validation)",
			input:       "12345678",
			wantErr:     false, // parseDistance accepts this, validation rejects it
			expected:    "12345678",
		},
		{
			name:        "multiple decimal points",
			input:       "123.456.789",
			wantErr:     true,
			expectedErr: "invalid distance format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDistance(tt.input)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseDistance() expected error, got nil")
					return
				}
				
				if tt.expectedErr != "" && !contains(err.Error(), tt.expectedErr) {
					t.Errorf("parseDistance() error = %v, expected to contain %v", err.Error(), tt.expectedErr)
				}
				
				// Verify it's a ParsingError with correct type
				if pe, ok := err.(*ParsingError); ok {
					if pe.Type != ErrorTypeDistance {
						t.Errorf("parseDistance() error type = %v, want %v", pe.Type, ErrorTypeDistance)
					}
					if pe.Input != tt.input {
						t.Errorf("parseDistance() error input = %v, want %v", pe.Input, tt.input)
					}
				} else {
					t.Errorf("parseDistance() error is not ParsingError type")
				}
				
				// Verify zero value is returned on error
				if !result.IsZero() {
					t.Errorf("parseDistance() expected zero decimal on error, got %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("parseDistance() unexpected error = %v", err)
					return
				}
				
				// Verify the parsed distance matches expected
				expected, _ := decimal.NewFromString(tt.expected)
				if !result.Equal(expected) {
					t.Errorf("parseDistance() = %v, want %v", result, expected)
				}
			}
		})
	}
}

func TestValidateDistanceFormat(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		expectedErr string
	}{
		{
			name:    "valid format - exactly 8 digits",
			input:   "12345678.5",
			wantErr: false,
		},
		{
			name:    "valid format - more than 8 digits",
			input:   "123456789.123",
			wantErr: false,
		},
		{
			name:    "valid format - multiple fractional digits",
			input:   "87654321.123456",
			wantErr: false,
		},
		{
			name:    "valid format - single fractional digit",
			input:   "99999999.9",
			wantErr: false,
		},
		{
			name:        "invalid - less than 8 digits",
			input:       "1234567.5",
			wantErr:     true,
			expectedErr: "invalid distance format",
		},
		{
			name:        "invalid - no decimal point",
			input:       "12345678",
			wantErr:     true,
			expectedErr: "invalid distance format",
		},
		{
			name:        "invalid - no fractional part",
			input:       "12345678.",
			wantErr:     true,
			expectedErr: "invalid distance format",
		},
		{
			name:        "invalid - non-numeric characters",
			input:       "1234567a.5",
			wantErr:     true,
			expectedErr: "invalid distance format",
		},
		{
			name:        "invalid - negative sign",
			input:       "-12345678.5",
			wantErr:     true,
			expectedErr: "invalid distance format",
		},
		{
			name:        "valid - leading zeros are acceptable",
			input:       "01234567.5", // 8 digits with leading zero
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDistanceFormat(tt.input)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateDistanceFormat() expected error, got nil")
					return
				}
				
				if tt.expectedErr != "" && !contains(err.Error(), tt.expectedErr) {
					t.Errorf("validateDistanceFormat() error = %v, expected to contain %v", err.Error(), tt.expectedErr)
				}
			} else {
				if err != nil {
					t.Errorf("validateDistanceFormat() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestParseDistanceWithValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid distance passes both validation and parsing",
			input:   "12345678.123",
			wantErr: false,
		},
		{
			name:    "large valid distance",
			input:   "999999999999.999",
			wantErr: false,
		},
		{
			name:    "fails validation - wrong format",
			input:   "1234567.5",
			wantErr: true,
		},
		{
			name:    "fails parsing - negative value",
			input:   "-87654321.5",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDistanceWithValidation(tt.input)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseDistanceWithValidation() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("parseDistanceWithValidation() unexpected error = %v", err)
					return
				}
				
				// Verify the result is not zero
				if result.IsZero() && tt.input != "00000000.0" {
					t.Errorf("parseDistanceWithValidation() returned zero for non-zero input")
				}
				
				// Verify precision is maintained
				expected, _ := decimal.NewFromString(tt.input)
				if !result.Equal(expected) {
					t.Errorf("parseDistanceWithValidation() = %v, want %v", result, expected)
				}
			}
		})
	}
}

func TestDistancePattern(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Valid patterns
		{"8 digits with 1 decimal", "12345678.5", true},
		{"9 digits with 2 decimals", "123456789.12", true},
		{"10 digits with 3 decimals", "1234567890.123", true},
		{"many digits with many decimals", "12345678901234567890.123456789", true},
		
		// Invalid patterns
		{"7 digits", "1234567.5", false},
		{"no decimal point", "12345678", false},
		{"no fractional part", "12345678.", false},
		{"negative", "-12345678.5", false},
		{"letters", "1234567a.5", false},
		{"multiple decimal points", "12345678.5.5", false},
		{"empty", "", false},
		{"only decimal", ".5", false},
		{"decimal at start", ".12345678", false},
		{"spaces", "12345678 .5", false},
		{"leading zero allowed", "01234567.5", true}, // Leading zeros should be acceptable
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := distancePattern.MatchString(tt.input)
			if result != tt.expected {
				t.Errorf("distancePattern.MatchString(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDecimalPrecision(t *testing.T) {
	t.Run("high precision maintenance", func(t *testing.T) {
		input := "12345678.123456789012345"
		result, err := parseDistance(input)
		if err != nil {
			t.Fatalf("parseDistance() unexpected error = %v", err)
		}
		
		// Verify precision is maintained by converting back to string
		resultStr := result.String()
		if resultStr != input {
			t.Errorf("parseDistance() precision lost: got %v, want %v", resultStr, input)
		}
	})
	
	t.Run("large number precision", func(t *testing.T) {
		input := "999999999999999999999.999999999999999999"
		result, err := parseDistance(input)
		if err != nil {
			t.Fatalf("parseDistance() unexpected error = %v", err)
		}
		
		// Verify the number is correctly parsed (shopspring/decimal should handle this)
		if result.IsZero() {
			t.Errorf("parseDistance() returned zero for large number")
		}
		
		// Verify it's actually larger than a reasonable threshold
		threshold := decimal.NewFromInt(999999999999999999)
		if result.LessThan(threshold) {
			t.Errorf("parseDistance() result %v is less than expected threshold %v", result, threshold)
		}
	})
}