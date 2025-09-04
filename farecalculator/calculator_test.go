package farecalculator

import (
	"testing"
	"time"

	"golang-taxi-fare/models"
	"github.com/shopspring/decimal"
)

func TestFareConstants(t *testing.T) {
	// Test that fare constants are correctly defined
	if !BaseFare.Equal(decimal.NewFromInt(400)) {
		t.Errorf("Expected BaseFare to be 400, got %s", BaseFare.String())
	}
	
	if !BaseDistance.Equal(decimal.NewFromInt(1000)) {
		t.Errorf("Expected BaseDistance to be 1000m, got %s", BaseDistance.String())
	}
	
	if !StandardRate.Equal(decimal.NewFromInt(40)) {
		t.Errorf("Expected StandardRate to be 40, got %s", StandardRate.String())
	}
	
	if !StandardUnit.Equal(decimal.NewFromInt(400)) {
		t.Errorf("Expected StandardUnit to be 400m, got %s", StandardUnit.String())
	}
	
	if !StandardThreshold.Equal(decimal.NewFromInt(10000)) {
		t.Errorf("Expected StandardThreshold to be 10000m, got %s", StandardThreshold.String())
	}
	
	if !ExtendedRate.Equal(decimal.NewFromInt(40)) {
		t.Errorf("Expected ExtendedRate to be 40, got %s", ExtendedRate.String())
	}
	
	if !ExtendedUnit.Equal(decimal.NewFromInt(350)) {
		t.Errorf("Expected ExtendedUnit to be 350m, got %s", ExtendedUnit.String())
	}
}

func TestNewCalculator(t *testing.T) {
	calc := NewCalculator()
	if calc == nil {
		t.Error("Expected non-nil calculator")
	}
	
	// Test that it implements the Calculator interface
	_, ok := calc.(Calculator)
	if !ok {
		t.Error("Calculator should implement Calculator interface")
	}
}

func TestTaxiCalculator_CalculateFare(t *testing.T) {
	calc := NewCalculator().(*TaxiCalculator)
	
	tests := []struct {
		name             string
		distance         decimal.Decimal
		expectedBase     decimal.Decimal
		expectedStandard decimal.Decimal
		expectedExtended decimal.Decimal
		expectedTotal    decimal.Decimal
	}{
		{
			name:             "zero distance",
			distance:         decimal.Zero,
			expectedBase:     decimal.Zero,
			expectedStandard: decimal.Zero,
			expectedExtended: decimal.Zero,
			expectedTotal:    decimal.Zero,
		},
		{
			name:             "negative distance",
			distance:         decimal.NewFromInt(-100),
			expectedBase:     decimal.Zero,
			expectedStandard: decimal.Zero,
			expectedExtended: decimal.Zero,
			expectedTotal:    decimal.Zero,
		},
		{
			name:             "exactly 500m (base fare)",
			distance:         decimal.NewFromInt(500),
			expectedBase:     decimal.NewFromInt(400),
			expectedStandard: decimal.Zero,
			expectedExtended: decimal.Zero,
			expectedTotal:    decimal.NewFromInt(400),
		},
		{
			name:             "exactly 1km (base fare boundary)",
			distance:         decimal.NewFromInt(1000),
			expectedBase:     decimal.NewFromInt(400),
			expectedStandard: decimal.Zero,
			expectedExtended: decimal.Zero,
			expectedTotal:    decimal.NewFromInt(400),
		},
		{
			name:             "1.5km (base + standard)",
			distance:         decimal.NewFromInt(1500),
			expectedBase:     decimal.NewFromInt(400),
			expectedStandard: decimal.NewFromInt(80), // 500m = 2 units of 400m = 2 * 40 = 80
			expectedExtended: decimal.Zero,
			expectedTotal:    decimal.NewFromInt(480),
		},
		{
			name:             "2km (base + standard)",
			distance:         decimal.NewFromInt(2000),
			expectedBase:     decimal.NewFromInt(400),
			expectedStandard: decimal.NewFromInt(120), // 1000m = 3 units of 400m = 3 * 40 = 120
			expectedExtended: decimal.Zero,
			expectedTotal:    decimal.NewFromInt(520),
		},
		{
			name:             "exactly 10km (base + standard, no extended)",
			distance:         decimal.NewFromInt(10000),
			expectedBase:     decimal.NewFromInt(400),
			expectedStandard: decimal.NewFromInt(920), // 9000m = 22.5 units, rounded up = 23 units * 40 = 920
			expectedExtended: decimal.Zero,
			expectedTotal:    decimal.NewFromInt(1320), // 400 + 920 = 1320
		},
		{
			name:             "12km (base + standard + extended)",
			distance:         decimal.NewFromInt(12000),
			expectedBase:     decimal.NewFromInt(400),
			expectedStandard: decimal.NewFromInt(920), // 9000m = 23 units of 400m = 920
			expectedExtended: decimal.NewFromInt(240), // 2000m = 6 units of 350m = 6 * 40 = 240
			expectedTotal:    decimal.NewFromInt(1560), // 400 + 920 + 240 = 1560
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.CalculateFare(tt.distance)
			
			if !result.Distance.Equal(tt.distance) {
				t.Errorf("Expected distance %s, got %s", tt.distance.String(), result.Distance.String())
			}
			
			if !result.BaseFareAmount.Equal(tt.expectedBase) {
				t.Errorf("Expected base fare %s, got %s", tt.expectedBase.String(), result.BaseFareAmount.String())
			}
			
			if !result.StandardFareAmount.Equal(tt.expectedStandard) {
				t.Errorf("Expected standard fare %s, got %s", tt.expectedStandard.String(), result.StandardFareAmount.String())
			}
			
			if !result.ExtendedFareAmount.Equal(tt.expectedExtended) {
				t.Errorf("Expected extended fare %s, got %s", tt.expectedExtended.String(), result.ExtendedFareAmount.String())
			}
			
			if !result.TotalFare.Equal(tt.expectedTotal) {
				t.Errorf("Expected total fare %s, got %s", tt.expectedTotal.String(), result.TotalFare.String())
			}
		})
	}
}

func TestTaxiCalculator_CalculateFareBoundaryConditions(t *testing.T) {
	calc := NewCalculator().(*TaxiCalculator)
	
	// Test exact boundary at 1km
	result1km := calc.CalculateFare(decimal.NewFromInt(1000))
	expected1km := decimal.NewFromInt(400)
	if !result1km.TotalFare.Equal(expected1km) {
		t.Errorf("At exactly 1km, expected %s, got %s", expected1km.String(), result1km.TotalFare.String())
	}
	
	// Test just over 1km
	result1001m := calc.CalculateFare(decimal.NewFromInt(1001))
	expectedOver1km := decimal.NewFromInt(440) // 400 base + 40 for first 400m unit
	if !result1001m.TotalFare.Equal(expectedOver1km) {
		t.Errorf("At 1001m, expected %s, got %s", expectedOver1km.String(), result1001m.TotalFare.String())
	}
	
	// Test exact boundary at 10km
	result10km := calc.CalculateFare(decimal.NewFromInt(10000))
	// Base: 400, Standard: 9000m = 23 units of 400m (rounded up) = 23 * 40 = 920
	// Actually: 9000 / 400 = 22.5, rounded up = 23
	expectedUnits := decimal.NewFromFloat(9000.0).Div(decimal.NewFromInt(400)).Ceil()
	expectedStandardAt10km := expectedUnits.Mul(decimal.NewFromInt(40))
	expected10km := decimal.NewFromInt(400).Add(expectedStandardAt10km)
	
	if !result10km.TotalFare.Equal(expected10km) {
		t.Errorf("At exactly 10km, expected %s, got %s", expected10km.String(), result10km.TotalFare.String())
	}
}

func TestTaxiCalculator_CalculateFromRecords(t *testing.T) {
	calc := NewCalculator().(*TaxiCalculator)
	baseTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	
	tests := []struct {
		name     string
		records  []models.DistanceRecord
		expected models.FareCalculation
	}{
		{
			name:    "empty records",
			records: []models.DistanceRecord{},
			expected: models.FareCalculation{
				BaseFare:     decimal.Zero,
				DistanceFare: decimal.Zero,
				TimeFare:     decimal.Zero,
				TotalFare:    decimal.Zero,
			},
		},
		{
			name: "single record",
			records: []models.DistanceRecord{
				{Timestamp: baseTime, Distance: decimal.NewFromInt(12345678)}, // Large number representing odometer
			},
			expected: models.FareCalculation{
				BaseFare:     decimal.Zero,
				DistanceFare: decimal.Zero,
				TimeFare:     decimal.Zero,
				TotalFare:    decimal.Zero,
			},
		},
		{
			name: "short trip (500m)",
			records: []models.DistanceRecord{
				{Timestamp: baseTime, Distance: decimal.NewFromInt(12345000)},
				{Timestamp: baseTime.Add(time.Minute), Distance: decimal.NewFromInt(12345500)},
			},
			expected: models.FareCalculation{
				BaseFare:     decimal.NewFromInt(400),
				DistanceFare: decimal.Zero,
				TimeFare:     decimal.Zero,
				TotalFare:    decimal.NewFromInt(400),
			},
		},
		{
			name: "medium trip (2km)",
			records: []models.DistanceRecord{
				{Timestamp: baseTime, Distance: decimal.NewFromInt(12345000)},
				{Timestamp: baseTime.Add(time.Minute), Distance: decimal.NewFromInt(12346000)},
				{Timestamp: baseTime.Add(2 * time.Minute), Distance: decimal.NewFromInt(12347000)},
			},
			expected: models.FareCalculation{
				BaseFare:     decimal.NewFromInt(400),
				DistanceFare: decimal.NewFromInt(120), // 1000m over base = 3 units * 40 = 120
				TimeFare:     decimal.Zero,
				TotalFare:    decimal.NewFromInt(520),
			},
		},
		{
			name: "long trip (12km)",
			records: []models.DistanceRecord{
				{Timestamp: baseTime, Distance: decimal.NewFromInt(12345000)},
				{Timestamp: baseTime.Add(5 * time.Minute), Distance: decimal.NewFromInt(12357000)},
			},
			expected: models.FareCalculation{
				BaseFare:     decimal.NewFromInt(400),
				DistanceFare: decimal.NewFromInt(1160), // Standard: 920 + Extended: 240
				TimeFare:     decimal.Zero,
				TotalFare:    decimal.NewFromInt(1560),
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.CalculateFromRecords(tt.records)
			
			if !result.BaseFare.Equal(tt.expected.BaseFare) {
				t.Errorf("Expected base fare %s, got %s", tt.expected.BaseFare.String(), result.BaseFare.String())
			}
			
			if !result.DistanceFare.Equal(tt.expected.DistanceFare) {
				t.Errorf("Expected distance fare %s, got %s", tt.expected.DistanceFare.String(), result.DistanceFare.String())
			}
			
			if !result.TimeFare.Equal(tt.expected.TimeFare) {
				t.Errorf("Expected time fare %s, got %s", tt.expected.TimeFare.String(), result.TimeFare.String())
			}
			
			if !result.TotalFare.Equal(tt.expected.TotalFare) {
				t.Errorf("Expected total fare %s, got %s", tt.expected.TotalFare.String(), result.TotalFare.String())
			}
		})
	}
}

func TestFareBreakdown_String(t *testing.T) {
	breakdown := FareBreakdown{
		Distance:           decimal.NewFromFloat(1500.0),
		BaseFareAmount:     decimal.NewFromInt(400),
		StandardFareAmount: decimal.NewFromInt(80),
		ExtendedFareAmount: decimal.Zero,
		TotalFare:          decimal.NewFromInt(480),
	}
	
	str := breakdown.String()
	if str == "" {
		t.Error("String representation should not be empty")
	}
	
	// Check that all components are included in the string
	if !containsString(str, "1500.0") || !containsString(str, "400") || !containsString(str, "80") || !containsString(str, "480") {
		t.Errorf("String representation missing components: %s", str)
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestFareMonotonicity(t *testing.T) {
	calc := NewCalculator().(*TaxiCalculator)
	
	// Property-based test: fare should never decrease as distance increases
	distances := []int{0, 500, 1000, 1500, 2000, 5000, 10000, 12000, 15000, 20000}
	
	var prevFare decimal.Decimal
	for i, dist := range distances {
		result := calc.CalculateFare(decimal.NewFromInt(int64(dist)))
		
		if i > 0 && result.TotalFare.LessThan(prevFare) {
			t.Errorf("Fare monotonicity violated: distance %dm has fare %s, but previous distance had fare %s",
				dist, result.TotalFare.String(), prevFare.String())
		}
		
		prevFare = result.TotalFare
	}
}

func TestDecimalPrecision(t *testing.T) {
	calc := NewCalculator().(*TaxiCalculator)
	
	// Test with fractional meters to ensure decimal precision is maintained
	result := calc.CalculateFare(decimal.NewFromFloat(1500.7))
	
	if result.Distance.IsZero() {
		t.Error("Distance should be preserved with decimal precision")
	}
	
	// Ensure calculations are still accurate with decimal inputs
	expected := decimal.NewFromInt(480) // 400 base + 80 standard
	if !result.TotalFare.Equal(expected) {
		t.Errorf("Expected %s for 1500.7m, got %s", expected.String(), result.TotalFare.String())
	}
}