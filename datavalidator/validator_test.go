package datavalidator

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"golang-taxi-fare/models"
)

func TestValidationError(t *testing.T) {
	t.Run("ValidationError with record index", func(t *testing.T) {
		err := &ValidationError{
			Type:        ValidationErrorTypeTiming,
			Message:     "timestamp out of sequence",
			RecordIndex: 5,
			Field:       "timestamp",
			Input:       "12:30:45.123",
		}
		
		expected := `validation error at record 5 (timing): timestamp out of sequence (input: "12:30:45.123")`
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("ValidationError without record index", func(t *testing.T) {
		err := &ValidationError{
			Type:        ValidationErrorTypeSequence,
			Message:     "empty sequence",
			RecordIndex: -1,
			Field:       "sequence",
			Input:       "0",
		}
		
		expected := `validation error (sequence): empty sequence (input: "0")`
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})
}

func TestValidationErrorType_String(t *testing.T) {
	tests := []struct {
		name     string
		errorType ValidationErrorType
		expected  string
	}{
		{"timing", ValidationErrorTypeTiming, "timing"},
		{"format", ValidationErrorTypeFormat, "format"},
		{"mileage", ValidationErrorTypeMileage, "mileage"},
		{"sequence", ValidationErrorTypeSequence, "sequence"},
		{"constraint", ValidationErrorTypeConstraint, "constraint"},
		{"unknown", ValidationErrorType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.errorType.String(); got != tt.expected {
				t.Errorf("ValidationErrorType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestErrorConstructors(t *testing.T) {
	t.Run("TimingError", func(t *testing.T) {
		err := TimingError(3, "time out of sequence", "12:30:45.123")
		
		if err.Type != ValidationErrorTypeTiming {
			t.Errorf("Expected timing error type, got %v", err.Type)
		}
		if err.RecordIndex != 3 {
			t.Errorf("Expected record index 3, got %d", err.RecordIndex)
		}
		if err.Field != "timestamp" {
			t.Errorf("Expected timestamp field, got %s", err.Field)
		}
	})

	t.Run("FormatError", func(t *testing.T) {
		err := FormatError(1, "distance", "invalid format", "abc.def")
		
		if err.Type != ValidationErrorTypeFormat {
			t.Errorf("Expected format error type, got %v", err.Type)
		}
		if err.Field != "distance" {
			t.Errorf("Expected distance field, got %s", err.Field)
		}
	})

	t.Run("MileageError", func(t *testing.T) {
		err := MileageError(2, "negative mileage", "-123.45")
		
		if err.Type != ValidationErrorTypeMileage {
			t.Errorf("Expected mileage error type, got %v", err.Type)
		}
		if err.Field != "distance" {
			t.Errorf("Expected distance field, got %s", err.Field)
		}
	})

	t.Run("SequenceError", func(t *testing.T) {
		err := SequenceError("empty sequence", 0)
		
		if err.Type != ValidationErrorTypeSequence {
			t.Errorf("Expected sequence error type, got %v", err.Type)
		}
		if err.RecordIndex != -1 {
			t.Errorf("Expected record index -1, got %d", err.RecordIndex)
		}
	})

	t.Run("ConstraintError", func(t *testing.T) {
		err := ConstraintError(4, "timestamp", "zero timestamp", "0001-01-01T00:00:00Z")
		
		if err.Type != ValidationErrorTypeConstraint {
			t.Errorf("Expected constraint error type, got %v", err.Type)
		}
		if err.Field != "timestamp" {
			t.Errorf("Expected timestamp field, got %s", err.Field)
		}
	})
}

func TestNewValidator(t *testing.T) {
	validator := NewValidator()
	
	// Test that we get a DataValidator with default settings
	dv, ok := validator.(*DataValidator)
	if !ok {
		t.Fatalf("Expected *DataValidator, got %T", validator)
	}
	
	if dv.MaxInterval != 5*time.Minute {
		t.Errorf("Expected 5 minute max interval, got %v", dv.MaxInterval)
	}
	
	if !dv.AllowIdenticalTimestamps {
		t.Error("Expected identical timestamps to be allowed by default")
	}
	
	if !dv.AllowIdenticalMileage {
		t.Error("Expected identical mileage to be allowed by default")
	}
}

func TestNewValidatorWithOptions(t *testing.T) {
	maxInterval := 3 * time.Minute
	validator := NewValidatorWithOptions(maxInterval, false, false)
	
	dv, ok := validator.(*DataValidator)
	if !ok {
		t.Fatalf("Expected *DataValidator, got %T", validator)
	}
	
	if dv.MaxInterval != maxInterval {
		t.Errorf("Expected %v max interval, got %v", maxInterval, dv.MaxInterval)
	}
	
	if dv.AllowIdenticalTimestamps {
		t.Error("Expected identical timestamps to be disallowed")
	}
	
	if dv.AllowIdenticalMileage {
		t.Error("Expected identical mileage to be disallowed")
	}
}

func TestDataValidator_ValidateRecord(t *testing.T) {
	validator := NewValidator().(*DataValidator)
	
	t.Run("valid record", func(t *testing.T) {
		record := models.DistanceRecord{
			Timestamp: time.Now(),
			Distance:  decimal.NewFromFloat(12345678.9),
		}
		
		err := validator.ValidateRecord(record)
		if err != nil {
			t.Errorf("Expected no error for valid record, got %v", err)
		}
	})

	t.Run("zero timestamp", func(t *testing.T) {
		record := models.DistanceRecord{
			Timestamp: time.Time{},
			Distance:  decimal.NewFromFloat(12345678.9),
		}
		
		err := validator.ValidateRecord(record)
		if err == nil {
			t.Error("Expected error for zero timestamp")
		}
		
		ve, ok := err.(*ValidationError)
		if !ok {
			t.Errorf("Expected ValidationError, got %T", err)
		} else if ve.Type != ValidationErrorTypeFormat {
			t.Errorf("Expected format error, got %v", ve.Type)
		}
	})

	t.Run("negative distance", func(t *testing.T) {
		record := models.DistanceRecord{
			Timestamp: time.Now(),
			Distance:  decimal.NewFromFloat(-123.45),
		}
		
		err := validator.ValidateRecord(record)
		if err == nil {
			t.Error("Expected error for negative distance")
		}
		
		ve, ok := err.(*ValidationError)
		if !ok {
			t.Errorf("Expected ValidationError, got %T", err)
		} else if ve.Type != ValidationErrorTypeConstraint {
			t.Errorf("Expected constraint error, got %v", ve.Type)
		}
	})
}

func TestDataValidator_ValidateSequence(t *testing.T) {
	validator := NewValidator().(*DataValidator)
	baseTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	
	t.Run("empty sequence", func(t *testing.T) {
		err := validator.ValidateSequence([]models.DistanceRecord{})
		if err == nil {
			t.Error("Expected error for empty sequence")
		}
		
		ve, ok := err.(*ValidationError)
		if !ok {
			t.Errorf("Expected ValidationError, got %T", err)
		} else if ve.Type != ValidationErrorTypeSequence {
			t.Errorf("Expected sequence error, got %v", ve.Type)
		}
	})

	t.Run("single record", func(t *testing.T) {
		records := []models.DistanceRecord{
			{
				Timestamp: baseTime,
				Distance:  decimal.NewFromFloat(12345678.9),
			},
		}
		
		err := validator.ValidateSequence(records)
		if err != nil {
			t.Errorf("Expected no error for single valid record, got %v", err)
		}
	})

	t.Run("valid sequence", func(t *testing.T) {
		records := []models.DistanceRecord{
			{
				Timestamp: baseTime,
				Distance:  decimal.NewFromFloat(12345678.9),
			},
			{
				Timestamp: baseTime.Add(1 * time.Minute),
				Distance:  decimal.NewFromFloat(12345679.5),
			},
			{
				Timestamp: baseTime.Add(2 * time.Minute),
				Distance:  decimal.NewFromFloat(12345680.1),
			},
		}
		
		err := validator.ValidateSequence(records)
		if err != nil {
			t.Errorf("Expected no error for valid sequence, got %v", err)
		}
	})

	t.Run("timing constraint violation - decreasing timestamp", func(t *testing.T) {
		records := []models.DistanceRecord{
			{
				Timestamp: baseTime,
				Distance:  decimal.NewFromFloat(12345678.9),
			},
			{
				Timestamp: baseTime.Add(-1 * time.Minute), // Goes backwards
				Distance:  decimal.NewFromFloat(12345679.5),
			},
		}
		
		err := validator.ValidateSequence(records)
		if err == nil {
			t.Error("Expected error for decreasing timestamp")
		}
		
		ve, ok := err.(*ValidationError)
		if !ok {
			t.Errorf("Expected ValidationError, got %T", err)
		} else if ve.Type != ValidationErrorTypeTiming {
			t.Errorf("Expected timing error, got %v", ve.Type)
		}
	})

	t.Run("timing constraint violation - exceeds max interval", func(t *testing.T) {
		records := []models.DistanceRecord{
			{
				Timestamp: baseTime,
				Distance:  decimal.NewFromFloat(12345678.9),
			},
			{
				Timestamp: baseTime.Add(6 * time.Minute), // Exceeds 5-minute limit
				Distance:  decimal.NewFromFloat(12345679.5),
			},
		}
		
		err := validator.ValidateSequence(records)
		if err == nil {
			t.Error("Expected error for exceeding max interval")
		}
		
		ve, ok := err.(*ValidationError)
		if !ok {
			t.Errorf("Expected ValidationError, got %T", err)
		} else if ve.Type != ValidationErrorTypeTiming {
			t.Errorf("Expected timing error, got %v", ve.Type)
		}
	})

	t.Run("mileage constraint violation - decreasing mileage", func(t *testing.T) {
		records := []models.DistanceRecord{
			{
				Timestamp: baseTime,
				Distance:  decimal.NewFromFloat(12345678.9),
			},
			{
				Timestamp: baseTime.Add(1 * time.Minute),
				Distance:  decimal.NewFromFloat(12345678.5), // Decreases
			},
		}
		
		err := validator.ValidateSequence(records)
		if err == nil {
			t.Error("Expected error for decreasing mileage")
		}
		
		ve, ok := err.(*ValidationError)
		if !ok {
			t.Errorf("Expected ValidationError, got %T", err)
		} else if ve.Type != ValidationErrorTypeMileage {
			t.Errorf("Expected mileage error, got %v", ve.Type)
		}
	})

	t.Run("identical timestamps allowed by default", func(t *testing.T) {
		records := []models.DistanceRecord{
			{
				Timestamp: baseTime,
				Distance:  decimal.NewFromFloat(12345678.9),
			},
			{
				Timestamp: baseTime, // Same timestamp
				Distance:  decimal.NewFromFloat(12345679.5),
			},
		}
		
		err := validator.ValidateSequence(records)
		if err != nil {
			t.Errorf("Expected no error for identical timestamps (allowed by default), got %v", err)
		}
	})

	t.Run("identical mileage allowed by default", func(t *testing.T) {
		records := []models.DistanceRecord{
			{
				Timestamp: baseTime,
				Distance:  decimal.NewFromFloat(12345678.9),
			},
			{
				Timestamp: baseTime.Add(1 * time.Minute),
				Distance:  decimal.NewFromFloat(12345678.9), // Same distance
			},
		}
		
		err := validator.ValidateSequence(records)
		if err != nil {
			t.Errorf("Expected no error for identical mileage (allowed by default), got %v", err)
		}
	})
}

func TestDataValidator_ValidateSequenceWithStrictOptions(t *testing.T) {
	validator := NewValidatorWithOptions(5*time.Minute, false, false).(*DataValidator)
	baseTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	
	t.Run("identical timestamps not allowed", func(t *testing.T) {
		records := []models.DistanceRecord{
			{
				Timestamp: baseTime,
				Distance:  decimal.NewFromFloat(12345678.9),
			},
			{
				Timestamp: baseTime, // Same timestamp
				Distance:  decimal.NewFromFloat(12345679.5),
			},
		}
		
		err := validator.ValidateSequence(records)
		if err == nil {
			t.Error("Expected error for identical timestamps when not allowed")
		}
		
		ve, ok := err.(*ValidationError)
		if !ok {
			t.Errorf("Expected ValidationError, got %T", err)
		} else if ve.Type != ValidationErrorTypeTiming {
			t.Errorf("Expected timing error, got %v", ve.Type)
		}
	})

	t.Run("identical mileage not allowed", func(t *testing.T) {
		records := []models.DistanceRecord{
			{
				Timestamp: baseTime,
				Distance:  decimal.NewFromFloat(12345678.9),
			},
			{
				Timestamp: baseTime.Add(1 * time.Minute),
				Distance:  decimal.NewFromFloat(12345678.9), // Same distance
			},
		}
		
		err := validator.ValidateSequence(records)
		if err == nil {
			t.Error("Expected error for identical mileage when not allowed")
		}
		
		ve, ok := err.(*ValidationError)
		if !ok {
			t.Errorf("Expected ValidationError, got %T", err)
		} else if ve.Type != ValidationErrorTypeMileage {
			t.Errorf("Expected mileage error, got %v", ve.Type)
		}
	})
}