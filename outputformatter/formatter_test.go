package outputformatter

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"golang-taxi-fare/models"
)

func TestNewFormatter(t *testing.T) {
	formatter := NewFormatter()
	if formatter == nil {
		t.Error("Expected non-nil formatter")
	}
	
	// Test that it implements the OutputFormatter interface
	_, ok := formatter.(OutputFormatter)
	if !ok {
		t.Error("Formatter should implement OutputFormatter interface")
	}
}

func TestNewFormatterWithOutput(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewFormatterWithOutput(&buf)
	
	if formatter == nil {
		t.Error("Expected non-nil formatter")
	}
	
	// Test that it uses the custom output
	calculation := models.FareCalculation{
		TotalFare: decimal.NewFromInt(1250),
	}
	
	err := formatter.FormatCurrentFare(calculation)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, "1250") {
		t.Errorf("Expected output to contain '1250', got: %s", output)
	}
}

func TestConsoleFormatter_FormatCurrentFare(t *testing.T) {
	tests := []struct {
		name           string
		calculation    models.FareCalculation
		expectedOutput string
	}{
		{
			name: "integer fare",
			calculation: models.FareCalculation{
				TotalFare: decimal.NewFromInt(1000),
			},
			expectedOutput: "1000\n",
		},
		{
			name: "decimal fare rounded",
			calculation: models.FareCalculation{
				TotalFare: decimal.NewFromFloat(1234.7),
			},
			expectedOutput: "1235\n", // Rounded up
		},
		{
			name: "zero fare",
			calculation: models.FareCalculation{
				TotalFare: decimal.Zero,
			},
			expectedOutput: "0\n",
		},
		{
			name: "large fare",
			calculation: models.FareCalculation{
				TotalFare: decimal.NewFromInt(99999),
			},
			expectedOutput: "99999\n",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := NewFormatterWithOutput(&buf)
			
			err := formatter.FormatCurrentFare(tt.calculation)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			output := buf.String()
			if output != tt.expectedOutput {
				t.Errorf("Expected output %q, got %q", tt.expectedOutput, output)
			}
		})
	}
}

func TestConsoleFormatter_FormatRecords(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewFormatterWithOutput(&buf)
	baseTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	
	t.Run("empty records", func(t *testing.T) {
		buf.Reset()
		err := formatter.FormatRecords([]models.DistanceRecord{})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		
		output := buf.String()
		if !strings.Contains(output, "No records to display") {
			t.Errorf("Expected 'No records to display' message, got: %s", output)
		}
	})
	
	t.Run("single record", func(t *testing.T) {
		buf.Reset()
		records := []models.DistanceRecord{
			{Timestamp: baseTime, Distance: decimal.NewFromFloat(12345.6)},
		}
		
		err := formatter.FormatRecords(records)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		
		output := buf.String()
		expectedContains := []string{"Index", "Timestamp", "Distance", "Mileage Diff", "12:00:00.000", "12345.6", "0.0"}
		
		for _, expected := range expectedContains {
			if !strings.Contains(output, expected) {
				t.Errorf("Expected output to contain %q, got: %s", expected, output)
			}
		}
	})
	
	t.Run("multiple records with sorting", func(t *testing.T) {
		buf.Reset()
		records := []models.DistanceRecord{
			{Timestamp: baseTime, Distance: decimal.NewFromFloat(12345.0)},
			{Timestamp: baseTime.Add(time.Minute), Distance: decimal.NewFromFloat(12346.5)}, // diff: 1.5
			{Timestamp: baseTime.Add(2 * time.Minute), Distance: decimal.NewFromFloat(12349.0)}, // diff: 2.5
			{Timestamp: baseTime.Add(3 * time.Minute), Distance: decimal.NewFromFloat(12350.0)}, // diff: 1.0
		}
		
		err := formatter.FormatRecords(records)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		
		output := buf.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")
		
		// Check that records are sorted by mileage difference (descending)
		// The record with diff 2.5 should be first (after header)
		dataLines := lines[2:] // Skip header lines
		if len(dataLines) < 4 {
			t.Errorf("Expected at least 4 data lines, got %d", len(dataLines))
			return
		}
		
		// First data line should have the highest diff (2.5)
		if !strings.Contains(dataLines[0], "2.5") {
			t.Errorf("First data line should contain '2.5', got: %s", dataLines[0])
		}
	})
}

func TestConsoleFormatter_FormatProcessingResult(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewFormatterWithOutput(&buf)
	baseTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	
	t.Run("successful result", func(t *testing.T) {
		buf.Reset()
		result := models.ProcessingResult{
			Records: []models.DistanceRecord{
				{Timestamp: baseTime, Distance: decimal.NewFromFloat(12345.0)},
				{Timestamp: baseTime.Add(time.Minute), Distance: decimal.NewFromFloat(12346.0)},
			},
			Calculation: models.FareCalculation{
				BaseFare:     decimal.NewFromInt(400),
				DistanceFare: decimal.NewFromInt(80),
				TimeFare:     decimal.Zero,
				TotalFare:    decimal.NewFromInt(480),
			},
			TotalTime: 250 * time.Millisecond,
			Error:     nil,
		}
		
		err := formatter.FormatProcessingResult(result)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		
		output := buf.String()
		expectedContains := []string{"480", "Processing Summary", "Records processed: 2", "Total fare: 480 yen"}
		
		for _, expected := range expectedContains {
			if !strings.Contains(output, expected) {
				t.Errorf("Expected output to contain %q, got: %s", expected, output)
			}
		}
	})
	
	t.Run("error result", func(t *testing.T) {
		buf.Reset()
		result := models.ProcessingResult{
			Error: fmt.Errorf("processing failed"),
		}
		
		err := formatter.FormatProcessingResult(result)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		
		output := buf.String()
		if !strings.Contains(output, "Processing failed: processing failed") {
			t.Errorf("Expected error message, got: %s", output)
		}
	})
	
	t.Run("invalid result", func(t *testing.T) {
		buf.Reset()
		result := models.ProcessingResult{
			Records: []models.DistanceRecord{},
			Calculation: models.FareCalculation{
				TotalFare: decimal.NewFromInt(-100), // Negative fare makes it invalid
			},
			Error: nil,
		}
		
		err := formatter.FormatProcessingResult(result)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		
		output := buf.String()
		if !strings.Contains(output, "Invalid processing result") {
			t.Errorf("Expected invalid result message, got: %s", output)
		}
	})
}

func TestConsoleFormatter_FormatSummaryStatistics(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewFormatterWithOutput(&buf)
	baseTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	
	t.Run("empty records", func(t *testing.T) {
		buf.Reset()
		calculation := models.FareCalculation{TotalFare: decimal.Zero}
		
		err := formatter.FormatSummaryStatistics([]models.DistanceRecord{}, calculation)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		
		output := buf.String()
		if !strings.Contains(output, "No data for statistics") {
			t.Errorf("Expected no data message, got: %s", output)
		}
	})
	
	t.Run("multiple records", func(t *testing.T) {
		buf.Reset()
		records := []models.DistanceRecord{
			{Timestamp: baseTime, Distance: decimal.NewFromFloat(12345.0)},
			{Timestamp: baseTime.Add(time.Minute), Distance: decimal.NewFromFloat(12346.0)},
			{Timestamp: baseTime.Add(2 * time.Minute), Distance: decimal.NewFromFloat(12347.0)},
		}
		
		calculation := models.FareCalculation{
			BaseFare:     decimal.NewFromInt(400),
			DistanceFare: decimal.NewFromInt(120),
			TimeFare:     decimal.NewFromInt(50),
			TotalFare:    decimal.NewFromInt(570),
		}
		
		err := formatter.FormatSummaryStatistics(records, calculation)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		
		output := buf.String()
		expectedContains := []string{
			"Summary Statistics",
			"Total Records:",
			"3",
			"Base Fare:",
			"400 yen",
			"Distance Fare:",
			"120 yen",
			"Time Fare:",
			"50 yen",
			"Total Fare:",
			"570 yen",
		}
		
		for _, expected := range expectedContains {
			if !strings.Contains(output, expected) {
				t.Errorf("Expected output to contain %q, got: %s", expected, output)
			}
		}
	})
}

func TestCalculateStatistics(t *testing.T) {
	baseTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	
	tests := []struct {
		name     string
		records  []models.DistanceRecord
		expected Statistics
	}{
		{
			name:    "empty records",
			records: []models.DistanceRecord{},
			expected: Statistics{
				TotalRecords:    0,
				TotalDistance:   decimal.Zero,
				AverageDistance: decimal.Zero,
				MinDistance:     decimal.Zero,
				MaxDistance:     decimal.Zero,
			},
		},
		{
			name: "single record",
			records: []models.DistanceRecord{
				{Timestamp: baseTime, Distance: decimal.NewFromFloat(100.0)},
			},
			expected: Statistics{
				TotalRecords:    1,
				TotalDistance:   decimal.NewFromFloat(100.0),
				AverageDistance: decimal.NewFromFloat(100.0),
				MinDistance:     decimal.NewFromFloat(100.0),
				MaxDistance:     decimal.NewFromFloat(100.0),
			},
		},
		{
			name: "multiple records",
			records: []models.DistanceRecord{
				{Timestamp: baseTime, Distance: decimal.NewFromFloat(100.0)},
				{Timestamp: baseTime.Add(time.Minute), Distance: decimal.NewFromFloat(200.0)},
				{Timestamp: baseTime.Add(2 * time.Minute), Distance: decimal.NewFromFloat(150.0)},
			},
			expected: Statistics{
				TotalRecords:    3,
				TotalDistance:   decimal.NewFromFloat(450.0),
				AverageDistance: decimal.NewFromFloat(150.0),
				MinDistance:     decimal.NewFromFloat(100.0),
				MaxDistance:     decimal.NewFromFloat(200.0),
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calculation := models.FareCalculation{} // Not used in statistics calculation
			result := calculateStatistics(tt.records, calculation)
			
			if result.TotalRecords != tt.expected.TotalRecords {
				t.Errorf("TotalRecords = %d, want %d", result.TotalRecords, tt.expected.TotalRecords)
			}
			
			if !result.TotalDistance.Equal(tt.expected.TotalDistance) {
				t.Errorf("TotalDistance = %s, want %s", result.TotalDistance.String(), tt.expected.TotalDistance.String())
			}
			
			if !result.AverageDistance.Equal(tt.expected.AverageDistance) {
				t.Errorf("AverageDistance = %s, want %s", result.AverageDistance.String(), tt.expected.AverageDistance.String())
			}
			
			if len(tt.records) > 0 {
				if !result.MinDistance.Equal(tt.expected.MinDistance) {
					t.Errorf("MinDistance = %s, want %s", result.MinDistance.String(), tt.expected.MinDistance.String())
				}
				
				if !result.MaxDistance.Equal(tt.expected.MaxDistance) {
					t.Errorf("MaxDistance = %s, want %s", result.MaxDistance.String(), tt.expected.MaxDistance.String())
				}
			}
		})
	}
}

func TestCompactFormatter(t *testing.T) {
	t.Run("FormatCurrentFare", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := NewCompactFormatterWithOutput(&buf)
		
		calculation := models.FareCalculation{
			TotalFare: decimal.NewFromFloat(1234.7),
		}
		
		err := formatter.FormatCurrentFare(calculation)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		
		output := buf.String()
		if output != "1235\n" {
			t.Errorf("Expected '1235\\n', got %q", output)
		}
	})
	
	t.Run("FormatRecords", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := NewCompactFormatterWithOutput(&buf)
		baseTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		
		records := []models.DistanceRecord{
			{Timestamp: baseTime, Distance: decimal.NewFromFloat(100.0)},
			{Timestamp: baseTime.Add(time.Minute), Distance: decimal.NewFromFloat(102.5)},
		}
		
		err := formatter.FormatRecords(records)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		
		output := buf.String()
		expectedContains := []string{"Records: 2", "Distance: 2.5"}
		
		for _, expected := range expectedContains {
			if !strings.Contains(output, expected) {
				t.Errorf("Expected output to contain %q, got: %s", expected, output)
			}
		}
	})
	
	t.Run("FormatProcessingResult", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := NewCompactFormatterWithOutput(&buf)
		
		result := models.ProcessingResult{
			Calculation: models.FareCalculation{TotalFare: decimal.NewFromInt(500)},
			Error:       nil,
		}
		
		err := formatter.FormatProcessingResult(result)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		
		output := buf.String()
		if !strings.Contains(output, "500") {
			t.Errorf("Expected output to contain '500', got: %s", output)
		}
	})
}

func TestDebugFormatter(t *testing.T) {
	t.Run("FormatCurrentFare", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := NewDebugFormatterWithOutput(&buf)
		
		calculation := models.FareCalculation{
			BaseFare:     decimal.NewFromInt(400),
			DistanceFare: decimal.NewFromInt(120),
			TimeFare:     decimal.NewFromInt(30),
			TotalFare:    decimal.NewFromInt(550),
		}
		
		err := formatter.FormatCurrentFare(calculation)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		
		output := buf.String()
		expectedContains := []string{"Fare Breakdown", "Base Fare", "400", "Distance Fare", "120", "Time Fare", "30", "Total", "550"}
		
		for _, expected := range expectedContains {
			if !strings.Contains(output, expected) {
				t.Errorf("Expected output to contain %q, got: %s", expected, output)
			}
		}
	})
	
	t.Run("FormatRecords", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := NewDebugFormatterWithOutput(&buf)
		baseTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		
		records := []models.DistanceRecord{
			{Timestamp: baseTime, Distance: decimal.NewFromFloat(100.000)},
			{Timestamp: baseTime.Add(time.Minute), Distance: decimal.NewFromFloat(102.500)},
		}
		
		err := formatter.FormatRecords(records)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		
		output := buf.String()
		expectedContains := []string{"Detailed Record Information", "Cumulative", "100.000", "102.500", "2.500"}
		
		for _, expected := range expectedContains {
			if !strings.Contains(output, expected) {
				t.Errorf("Expected output to contain %q, got: %s", expected, output)
			}
		}
	})
}

// Benchmark tests for performance validation
func BenchmarkConsoleFormatter_FormatCurrentFare(b *testing.B) {
	var buf bytes.Buffer
	formatter := NewFormatterWithOutput(&buf)
	calculation := models.FareCalculation{TotalFare: decimal.NewFromInt(1234)}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		formatter.FormatCurrentFare(calculation)
	}
}

func BenchmarkConsoleFormatter_FormatRecords(b *testing.B) {
	var buf bytes.Buffer
	formatter := NewFormatterWithOutput(&buf)
	baseTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	
	// Create a moderate number of records for realistic benchmarking
	records := make([]models.DistanceRecord, 100)
	for i := range records {
		records[i] = models.DistanceRecord{
			Timestamp: baseTime.Add(time.Duration(i) * time.Minute),
			Distance:  decimal.NewFromInt(int64(12345000 + i*100)),
		}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		formatter.FormatRecords(records)
	}
}

func BenchmarkCalculateStatistics(b *testing.B) {
	baseTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	calculation := models.FareCalculation{TotalFare: decimal.NewFromInt(1000)}
	
	// Create records for benchmarking
	records := make([]models.DistanceRecord, 1000)
	for i := range records {
		records[i] = models.DistanceRecord{
			Timestamp: baseTime.Add(time.Duration(i) * time.Second),
			Distance:  decimal.NewFromInt(int64(12345000 + i*10)),
		}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculateStatistics(records, calculation)
	}
}