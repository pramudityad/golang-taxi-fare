package models

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestDistanceRecord(t *testing.T) {
	t.Run("String method", func(t *testing.T) {
		timestamp, _ := time.Parse("15:04:05.000", "14:30:25.123")
		distance := decimal.NewFromFloat(12.5)
		
		dr := DistanceRecord{
			Timestamp: timestamp,
			Distance:  distance,
		}
		
		result := dr.String()
		expected := "DistanceRecord{Timestamp: 14:30:25.123, Distance: 12.5}"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})
	
	t.Run("JSON marshaling", func(t *testing.T) {
		timestamp, _ := time.Parse("15:04:05.000", "14:30:25.123")
		distance := decimal.NewFromFloat(12.5)
		
		dr := DistanceRecord{
			Timestamp: timestamp,
			Distance:  distance,
		}
		
		jsonData, err := json.Marshal(dr)
		if err != nil {
			t.Fatalf("Failed to marshal: %v", err)
		}
		
		var unmarshaled DistanceRecord
		err = json.Unmarshal(jsonData, &unmarshaled)
		if err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}
		
		if !unmarshaled.Distance.Equal(dr.Distance) {
			t.Errorf("Expected distance %s, got %s", dr.Distance, unmarshaled.Distance)
		}
	})
}

func TestFareCalculation(t *testing.T) {
	t.Run("String method", func(t *testing.T) {
		fc := FareCalculation{
			BaseFare:     decimal.NewFromFloat(2.50),
			DistanceFare: decimal.NewFromFloat(10.00),
			TimeFare:     decimal.NewFromFloat(5.25),
			TotalFare:    decimal.NewFromFloat(17.75),
		}
		
		result := fc.String()
		expected := "FareCalculation{BaseFare: 2.5, DistanceFare: 10, TimeFare: 5.25, TotalFare: 17.75}"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})
	
	t.Run("Zero values", func(t *testing.T) {
		fc := FareCalculation{
			BaseFare:     decimal.Zero,
			DistanceFare: decimal.Zero,
			TimeFare:     decimal.Zero,
			TotalFare:    decimal.Zero,
		}
		
		result := fc.String()
		expected := "FareCalculation{BaseFare: 0, DistanceFare: 0, TimeFare: 0, TotalFare: 0}"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})
	
	t.Run("Large numbers precision", func(t *testing.T) {
		baseFare, _ := decimal.NewFromString("999999.99")
		distanceFare, _ := decimal.NewFromString("888888.88")
		timeFare, _ := decimal.NewFromString("777777.77")
		totalFare, _ := decimal.NewFromString("2666666.64")
		
		fc := FareCalculation{
			BaseFare:     baseFare,
			DistanceFare: distanceFare,
			TimeFare:     timeFare,
			TotalFare:    totalFare,
		}
		
		// Verify precision is maintained
		expectedBase, _ := decimal.NewFromString("999999.99")
		expectedTotal, _ := decimal.NewFromString("2666666.64")
		
		if !fc.BaseFare.Equal(expectedBase) {
			t.Errorf("Precision lost in BaseFare")
		}
		if !fc.TotalFare.Equal(expectedTotal) {
			t.Errorf("Precision lost in TotalFare")
		}
	})
}

func TestProcessingResult(t *testing.T) {
	t.Run("String method with no error", func(t *testing.T) {
		records := []DistanceRecord{
			{Timestamp: time.Now(), Distance: decimal.NewFromFloat(10.0)},
			{Timestamp: time.Now(), Distance: decimal.NewFromFloat(15.0)},
		}
		
		fc := FareCalculation{TotalFare: decimal.NewFromFloat(25.50)}
		duration := 5 * time.Minute
		
		pr := ProcessingResult{
			Records:     records,
			Calculation: fc,
			TotalTime:   duration,
			Error:       nil,
		}
		
		result := pr.String()
		if !contains(result, "Records: 2") {
			t.Errorf("Expected Records: 2 in result: %s", result)
		}
		if !contains(result, "TotalTime: 5m0s") {
			t.Errorf("Expected TotalTime: 5m0s in result: %s", result)
		}
		if !contains(result, "Error: nil") {
			t.Errorf("Expected Error: nil in result: %s", result)
		}
	})
	
	t.Run("String method with error", func(t *testing.T) {
		pr := ProcessingResult{
			Records:     []DistanceRecord{},
			Calculation: FareCalculation{},
			TotalTime:   0,
			Error:       errors.New("test error"),
		}
		
		result := pr.String()
		if !contains(result, "Error: test error") {
			t.Errorf("Expected Error: test error in result: %s", result)
		}
	})
	
	t.Run("IsValid method", func(t *testing.T) {
		// Valid result
		validResult := ProcessingResult{
			Records: []DistanceRecord{
				{Timestamp: time.Now(), Distance: decimal.NewFromFloat(10.0)},
			},
			Calculation: FareCalculation{TotalFare: decimal.NewFromFloat(15.50)},
			Error:       nil,
		}
		
		if !validResult.IsValid() {
			t.Error("Expected valid result to be valid")
		}
		
		// Invalid - has error
		invalidResult1 := ProcessingResult{
			Records: []DistanceRecord{
				{Timestamp: time.Now(), Distance: decimal.NewFromFloat(10.0)},
			},
			Calculation: FareCalculation{TotalFare: decimal.NewFromFloat(15.50)},
			Error:       errors.New("some error"),
		}
		
		if invalidResult1.IsValid() {
			t.Error("Expected result with error to be invalid")
		}
		
		// Invalid - no records
		invalidResult2 := ProcessingResult{
			Records:     []DistanceRecord{},
			Calculation: FareCalculation{TotalFare: decimal.NewFromFloat(15.50)},
			Error:       nil,
		}
		
		if invalidResult2.IsValid() {
			t.Error("Expected result with no records to be invalid")
		}
		
		// Invalid - negative fare
		invalidResult3 := ProcessingResult{
			Records: []DistanceRecord{
				{Timestamp: time.Now(), Distance: decimal.NewFromFloat(10.0)},
			},
			Calculation: FareCalculation{TotalFare: decimal.NewFromFloat(-5.00)},
			Error:       nil,
		}
		
		if invalidResult3.IsValid() {
			t.Error("Expected result with negative fare to be invalid")
		}
	})
}

// Helper function to check if string contains substring
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