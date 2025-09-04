// Package farecalculator provides Japanese taxi fare calculation functionality.
// It implements the standard Japanese taxi fare structure with base fare and distance-based rates.
package farecalculator

import (
	"fmt"
	
	"golang-taxi-fare/models"
	"github.com/shopspring/decimal"
)

// Fare rate constants based on Japanese taxi fare structure
var (
	// BaseFare is the initial fare for distances up to and including 1km (400 yen)
	BaseFare = decimal.NewFromInt(400)
	
	// BaseDistance is the distance threshold for base fare (1km = 1000m)
	BaseDistance = decimal.NewFromInt(1000)
	
	// StandardRate is the fare per unit for distances 1-10km (40 yen per 400m)
	StandardRate = decimal.NewFromInt(40)
	StandardUnit = decimal.NewFromInt(400) // meters per fare unit
	
	// StandardThreshold is the distance where extended rate begins (10km = 10000m)
	StandardThreshold = decimal.NewFromInt(10000)
	
	// ExtendedRate is the fare per unit for distances >10km (40 yen per 350m)
	ExtendedRate = decimal.NewFromInt(40)
	ExtendedUnit = decimal.NewFromInt(350) // meters per fare unit
)

// FareBreakdown provides detailed breakdown of fare calculation components
type FareBreakdown struct {
	// BaseFareAmount is the base fare portion (400 yen for ≤1km)
	BaseFareAmount decimal.Decimal `json:"base_fare_amount"`
	
	// StandardFareAmount is the standard rate portion (40 yen/400m for 1-10km)
	StandardFareAmount decimal.Decimal `json:"standard_fare_amount"`
	
	// ExtendedFareAmount is the extended rate portion (40 yen/350m for >10km)
	ExtendedFareAmount decimal.Decimal `json:"extended_fare_amount"`
	
	// TotalFare is the sum of all fare components
	TotalFare decimal.Decimal `json:"total_fare"`
	
	// Distance is the total distance used for calculation
	Distance decimal.Decimal `json:"distance"`
}

// String implements the Stringer interface for debugging
func (fb FareBreakdown) String() string {
	return fmt.Sprintf("FareBreakdown{Distance: %s, Base: %s, Standard: %s, Extended: %s, Total: %s}",
		fb.Distance.StringFixed(1), fb.BaseFareAmount.String(), 
		fb.StandardFareAmount.String(), fb.ExtendedFareAmount.String(), fb.TotalFare.String())
}

// Calculator defines the interface for fare calculation operations
type Calculator interface {
	// CalculateFare calculates the fare for a given distance in meters
	CalculateFare(distanceMeters decimal.Decimal) FareBreakdown
	
	// CalculateFromRecords calculates the cumulative fare from a sequence of distance records
	CalculateFromRecords(records []models.DistanceRecord) models.FareCalculation
}

// TaxiCalculator implements the Calculator interface with Japanese taxi fare logic
type TaxiCalculator struct{}

// NewCalculator creates a new TaxiCalculator instance
func NewCalculator() Calculator {
	return &TaxiCalculator{}
}

// CalculateFare calculates the fare for a given distance in meters using Japanese taxi fare structure
func (tc *TaxiCalculator) CalculateFare(distanceMeters decimal.Decimal) FareBreakdown {
	var baseFareAmount, standardFareAmount, extendedFareAmount decimal.Decimal
	
	// Handle negative or zero distance
	if distanceMeters.IsNegative() || distanceMeters.IsZero() {
		return FareBreakdown{
			Distance: distanceMeters,
			TotalFare: decimal.Zero,
		}
	}
	
	// Base fare: 400 yen for distance ≤ 1km
	if distanceMeters.LessThanOrEqual(BaseDistance) {
		baseFareAmount = BaseFare
	} else {
		baseFareAmount = BaseFare
		remainingDistance := distanceMeters.Sub(BaseDistance)
		
		// Standard rate: 40 yen per 400m for distances 1-10km
		standardDistance := remainingDistance
		if remainingDistance.GreaterThan(StandardThreshold.Sub(BaseDistance)) {
			standardDistance = StandardThreshold.Sub(BaseDistance) // 9km worth
		}
		
		if standardDistance.GreaterThan(decimal.Zero) {
			// Calculate number of 400m units (rounded up)
			standardUnits := standardDistance.Div(StandardUnit).Ceil()
			standardFareAmount = standardUnits.Mul(StandardRate)
		}
		
		// Extended rate: 40 yen per 350m for distances >10km
		if remainingDistance.GreaterThan(StandardThreshold.Sub(BaseDistance)) {
			extendedDistance := remainingDistance.Sub(StandardThreshold.Sub(BaseDistance))
			if extendedDistance.GreaterThan(decimal.Zero) {
				// Calculate number of 350m units (rounded up)
				extendedUnits := extendedDistance.Div(ExtendedUnit).Ceil()
				extendedFareAmount = extendedUnits.Mul(ExtendedRate)
			}
		}
	}
	
	totalFare := baseFareAmount.Add(standardFareAmount).Add(extendedFareAmount)
	
	return FareBreakdown{
		BaseFareAmount:     baseFareAmount,
		StandardFareAmount: standardFareAmount,
		ExtendedFareAmount: extendedFareAmount,
		TotalFare:          totalFare,
		Distance:           distanceMeters,
	}
}

// CalculateFromRecords calculates the cumulative fare from a sequence of distance records
// It uses the maximum distance as the basis for fare calculation (odometer reading)
func (tc *TaxiCalculator) CalculateFromRecords(records []models.DistanceRecord) models.FareCalculation {
	// Handle empty records
	if len(records) == 0 {
		return models.FareCalculation{
			BaseFare:     decimal.Zero,
			DistanceFare: decimal.Zero,
			TimeFare:     decimal.Zero,
			TotalFare:    decimal.Zero,
		}
	}
	
	// Find the maximum distance (assuming odometer readings)
	maxDistance := records[0].Distance
	minDistance := records[0].Distance
	
	for _, record := range records[1:] {
		if record.Distance.GreaterThan(maxDistance) {
			maxDistance = record.Distance
		}
		if record.Distance.LessThan(minDistance) {
			minDistance = record.Distance
		}
	}
	
	// Calculate total travel distance
	travelDistance := maxDistance.Sub(minDistance)
	
	// Convert from kilometers to meters if needed
	// Assuming input is in meters based on the large decimal values in tests
	fareBreakdown := tc.CalculateFare(travelDistance)
	
	// Map to FareCalculation struct
	// Note: Japanese taxi fares typically don't separate time-based charges in this simple model
	// All charges are distance-based, so TimeFare is zero
	return models.FareCalculation{
		BaseFare:     fareBreakdown.BaseFareAmount,
		DistanceFare: fareBreakdown.StandardFareAmount.Add(fareBreakdown.ExtendedFareAmount),
		TimeFare:     decimal.Zero, // No time-based fare in this implementation
		TotalFare:    fareBreakdown.TotalFare,
	}
}