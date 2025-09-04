package models

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// DistanceRecord represents a time-stamped distance measurement
type DistanceRecord struct {
	Timestamp time.Time       `json:"timestamp"`
	Distance  decimal.Decimal `json:"distance"`
}

// String implements the Stringer interface for debugging
func (dr DistanceRecord) String() string {
	return fmt.Sprintf("DistanceRecord{Timestamp: %s, Distance: %s}", 
		dr.Timestamp.Format("15:04:05.000"), dr.Distance.String())
}

// FareCalculation represents the result of fare calculations with precise decimal arithmetic
type FareCalculation struct {
	BaseFare     decimal.Decimal `json:"base_fare"`
	DistanceFare decimal.Decimal `json:"distance_fare"`
	TimeFare     decimal.Decimal `json:"time_fare"`
	TotalFare    decimal.Decimal `json:"total_fare"`
}

// String implements the Stringer interface for debugging
func (fc FareCalculation) String() string {
	return fmt.Sprintf("FareCalculation{BaseFare: %s, DistanceFare: %s, TimeFare: %s, TotalFare: %s}",
		fc.BaseFare.String(), fc.DistanceFare.String(), fc.TimeFare.String(), fc.TotalFare.String())
}

// ProcessingResult represents the complete result of processing distance records
type ProcessingResult struct {
	Records     []DistanceRecord `json:"records"`
	Calculation FareCalculation  `json:"calculation"`
	TotalTime   time.Duration    `json:"total_time"`
	Error       error           `json:"error,omitempty"`
}

// String implements the Stringer interface for debugging
func (pr ProcessingResult) String() string {
	errorStr := "nil"
	if pr.Error != nil {
		errorStr = pr.Error.Error()
	}
	return fmt.Sprintf("ProcessingResult{Records: %d, Calculation: %s, TotalTime: %s, Error: %s}",
		len(pr.Records), pr.Calculation.String(), pr.TotalTime.String(), errorStr)
}

// IsValid checks if the ProcessingResult contains valid data
func (pr ProcessingResult) IsValid() bool {
	return pr.Error == nil && len(pr.Records) > 0 && !pr.Calculation.TotalFare.IsNegative()
}