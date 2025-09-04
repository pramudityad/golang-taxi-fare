// Package datavalidator provides comprehensive validation for taxi fare calculation data.
// It validates timing constraints, mileage progression, and data integrity for DistanceRecord sequences.
package datavalidator

import (
	"fmt"
	"time"

	"golang-taxi-fare/models"
)

// ValidationError represents different types of validation errors with context
type ValidationError struct {
	Type        ValidationErrorType
	Message     string
	RecordIndex int    // Index of the record in sequence that failed validation
	Field       string // Field that failed validation (timestamp, distance, etc.)
	Input       string // Input data that caused the error
}

// ValidationErrorType categorizes different validation error types
type ValidationErrorType int

const (
	// ValidationErrorTypeTiming indicates a timing constraint violation
	ValidationErrorTypeTiming ValidationErrorType = iota
	// ValidationErrorTypeFormat indicates a format validation error
	ValidationErrorTypeFormat
	// ValidationErrorTypeMileage indicates a mileage progression error
	ValidationErrorTypeMileage
	// ValidationErrorTypeSequence indicates a sequence-level validation error
	ValidationErrorTypeSequence
	// ValidationErrorTypeConstraint indicates a general constraint violation
	ValidationErrorTypeConstraint
)

// Error implements the error interface
func (ve *ValidationError) Error() string {
	if ve.RecordIndex >= 0 {
		return fmt.Sprintf("validation error at record %d (%s): %s (input: %q)", 
			ve.RecordIndex, ve.Type.String(), ve.Message, ve.Input)
	}
	return fmt.Sprintf("validation error (%s): %s (input: %q)", 
		ve.Type.String(), ve.Message, ve.Input)
}

// String returns a human-readable description of the validation error type
func (vet ValidationErrorType) String() string {
	switch vet {
	case ValidationErrorTypeTiming:
		return "timing"
	case ValidationErrorTypeFormat:
		return "format"
	case ValidationErrorTypeMileage:
		return "mileage"
	case ValidationErrorTypeSequence:
		return "sequence"
	case ValidationErrorTypeConstraint:
		return "constraint"
	default:
		return "unknown"
	}
}

// TimingError creates a ValidationError for timing constraint violations
func TimingError(recordIndex int, message string, input interface{}) *ValidationError {
	return &ValidationError{
		Type:        ValidationErrorTypeTiming,
		Message:     message,
		RecordIndex: recordIndex,
		Field:       "timestamp",
		Input:       fmt.Sprintf("%v", input),
	}
}

// FormatError creates a ValidationError for format validation failures
func FormatError(recordIndex int, field string, message string, input interface{}) *ValidationError {
	return &ValidationError{
		Type:        ValidationErrorTypeFormat,
		Message:     message,
		RecordIndex: recordIndex,
		Field:       field,
		Input:       fmt.Sprintf("%v", input),
	}
}

// MileageError creates a ValidationError for mileage progression violations
func MileageError(recordIndex int, message string, input interface{}) *ValidationError {
	return &ValidationError{
		Type:        ValidationErrorTypeMileage,
		Message:     message,
		RecordIndex: recordIndex,
		Field:       "distance",
		Input:       fmt.Sprintf("%v", input),
	}
}

// SequenceError creates a ValidationError for sequence-level violations
func SequenceError(message string, input interface{}) *ValidationError {
	return &ValidationError{
		Type:        ValidationErrorTypeSequence,
		Message:     message,
		RecordIndex: -1, // Sequence errors don't have a specific record index
		Field:       "sequence",
		Input:       fmt.Sprintf("%v", input),
	}
}

// ConstraintError creates a ValidationError for general constraint violations
func ConstraintError(recordIndex int, field string, message string, input interface{}) *ValidationError {
	return &ValidationError{
		Type:        ValidationErrorTypeConstraint,
		Message:     message,
		RecordIndex: recordIndex,
		Field:       field,
		Input:       fmt.Sprintf("%v", input),
	}
}

// Validator defines the interface for data validation operations
type Validator interface {
	// ValidateRecord validates a single DistanceRecord for basic constraints
	ValidateRecord(record models.DistanceRecord) error
	
	// ValidateSequence validates a complete sequence of DistanceRecord entries
	ValidateSequence(records []models.DistanceRecord) error
}

// DataValidator implements the Validator interface with comprehensive validation rules
type DataValidator struct {
	// MaxInterval defines the maximum allowed time interval between consecutive records
	MaxInterval time.Duration
	
	// AllowIdenticalTimestamps determines if consecutive records can have identical timestamps
	AllowIdenticalTimestamps bool
	
	// AllowIdenticalMileage determines if consecutive records can have identical mileage
	AllowIdenticalMileage bool
}

// NewValidator creates a new DataValidator with default settings
func NewValidator() Validator {
	return &DataValidator{
		MaxInterval:              5 * time.Minute, // 5-minute maximum interval
		AllowIdenticalTimestamps: true,            // Allow identical timestamps
		AllowIdenticalMileage:    true,            // Allow identical mileage readings
	}
}

// NewValidatorWithOptions creates a new DataValidator with custom options
func NewValidatorWithOptions(maxInterval time.Duration, allowIdenticalTimestamps, allowIdenticalMileage bool) Validator {
	return &DataValidator{
		MaxInterval:              maxInterval,
		AllowIdenticalTimestamps: allowIdenticalTimestamps,
		AllowIdenticalMileage:    allowIdenticalMileage,
	}
}

// ValidateRecord validates a single DistanceRecord for basic constraints
func (dv *DataValidator) ValidateRecord(record models.DistanceRecord) error {
	// Validate timestamp is not zero
	if record.Timestamp.IsZero() {
		return FormatError(0, "timestamp", "timestamp cannot be zero", record.Timestamp)
	}
	
	// Validate distance is non-negative
	if record.Distance.IsNegative() {
		return ConstraintError(0, "distance", "distance cannot be negative", record.Distance)
	}
	
	// Additional basic validation can be added here
	
	return nil
}

// ValidateSequence validates a complete sequence of DistanceRecord entries
func (dv *DataValidator) ValidateSequence(records []models.DistanceRecord) error {
	// Handle empty sequence
	if len(records) == 0 {
		return SequenceError("sequence cannot be empty", len(records))
	}
	
	// Single record validation
	if len(records) == 1 {
		return dv.ValidateRecord(records[0])
	}
	
	// Validate each record individually first
	for i, record := range records {
		if err := dv.ValidateRecord(record); err != nil {
			// Update record index for context
			if ve, ok := err.(*ValidationError); ok {
				ve.RecordIndex = i
			}
			return err
		}
	}
	
	// Validate sequence constraints
	for i := 1; i < len(records); i++ {
		current := records[i]
		previous := records[i-1]
		
		// Validate timing constraints
		if err := dv.validateTimingConstraints(previous, current, i); err != nil {
			return err
		}
		
		// Validate mileage progression
		if err := dv.validateMileageProgression(previous, current, i); err != nil {
			return err
		}
	}
	
	return nil
}

// validateTimingConstraints checks timing constraints between consecutive records
func (dv *DataValidator) validateTimingConstraints(previous, current models.DistanceRecord, currentIndex int) error {
	timeDiff := current.Timestamp.Sub(previous.Timestamp)
	
	// Check for non-decreasing timestamps
	if timeDiff < 0 {
		return TimingError(currentIndex, 
			fmt.Sprintf("timestamp must be non-decreasing, got %s before %s", 
				current.Timestamp.Format("15:04:05.000"), 
				previous.Timestamp.Format("15:04:05.000")),
			current.Timestamp)
	}
	
	// Check for identical timestamps if not allowed
	if timeDiff == 0 && !dv.AllowIdenticalTimestamps {
		return TimingError(currentIndex, 
			fmt.Sprintf("identical timestamps not allowed: %s", 
				current.Timestamp.Format("15:04:05.000")),
			current.Timestamp)
	}
	
	// Check maximum interval constraint
	if timeDiff > dv.MaxInterval {
		return TimingError(currentIndex, 
			fmt.Sprintf("time interval exceeds maximum allowed (%v), got %v", 
				dv.MaxInterval, timeDiff),
			timeDiff)
	}
	
	return nil
}

// validateMileageProgression checks mileage progression between consecutive records
func (dv *DataValidator) validateMileageProgression(previous, current models.DistanceRecord, currentIndex int) error {
	mileageDiff := current.Distance.Sub(previous.Distance)
	
	// Check for non-decreasing mileage
	if mileageDiff.IsNegative() {
		return MileageError(currentIndex, 
			fmt.Sprintf("mileage must be non-decreasing, got %s before %s", 
				current.Distance.String(), previous.Distance.String()),
			current.Distance)
	}
	
	// Check for identical mileage if not allowed
	if mileageDiff.IsZero() && !dv.AllowIdenticalMileage {
		return MileageError(currentIndex, 
			fmt.Sprintf("identical mileage readings not allowed: %s", 
				current.Distance.String()),
			current.Distance)
	}
	
	return nil
}