// Package outputformatter provides structured output formatting for taxi fare calculation results.
// It handles fare display to stdout and record sorting with tabular formatting using text/tabwriter.
package outputformatter

import (
	"fmt"
	"io"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/shopspring/decimal"
	"golang-taxi-fare/models"
)

// OutputFormatter defines the interface for output formatting operations
type OutputFormatter interface {
	// FormatCurrentFare formats and displays the current fare calculation result
	FormatCurrentFare(calculation models.FareCalculation) error
	
	// FormatRecords formats and displays the processed records with sorting
	FormatRecords(records []models.DistanceRecord) error
	
	// FormatProcessingResult formats and displays the complete processing result
	FormatProcessingResult(result models.ProcessingResult) error
	
	// FormatSummaryStatistics formats and displays summary statistics
	FormatSummaryStatistics(records []models.DistanceRecord, calculation models.FareCalculation) error
}

// ConsoleFormatter implements the OutputFormatter interface with console output
type ConsoleFormatter struct {
	output io.Writer
	writer *tabwriter.Writer
}

// NewFormatter creates a new ConsoleFormatter with stdout output
func NewFormatter() OutputFormatter {
	return NewFormatterWithOutput(os.Stdout)
}

// NewFormatterWithOutput creates a new ConsoleFormatter with custom output writer
func NewFormatterWithOutput(output io.Writer) OutputFormatter {
	writer := tabwriter.NewWriter(output, 0, 8, 1, '\t', 0)
	return &ConsoleFormatter{
		output: output,
		writer: writer,
	}
}

// FormatCurrentFare formats and displays the current fare calculation result
func (cf *ConsoleFormatter) FormatCurrentFare(calculation models.FareCalculation) error {
	// Convert decimal to integer for display (rounded)
	totalFareInt := calculation.TotalFare.Round(0).IntPart()
	
	fmt.Fprintf(cf.output, "%d\n", totalFareInt)
	return nil
}

// FormatRecords formats and displays the processed records with sorting
func (cf *ConsoleFormatter) FormatRecords(records []models.DistanceRecord) error {
	if len(records) == 0 {
		fmt.Fprint(cf.output, "No records to display\n")
		return nil
	}
	
	// Sort records by mileage difference (descending)
	sortedRecords := make([]RecordWithDifference, 0, len(records))
	
	for i, record := range records {
		diff := decimal.Zero
		if i > 0 {
			diff = record.Distance.Sub(records[i-1].Distance)
		}
		
		sortedRecords = append(sortedRecords, RecordWithDifference{
			Record:          record,
			MileageDiff:     diff,
			Index:           i,
		})
	}
	
	// Sort by mileage difference in descending order
	sort.Slice(sortedRecords, func(i, j int) bool {
		return sortedRecords[i].MileageDiff.GreaterThan(sortedRecords[j].MileageDiff)
	})
	
	// Format output using tabwriter
	fmt.Fprintln(cf.writer, "Index\tTimestamp\tDistance\tMileage Diff")
	fmt.Fprintln(cf.writer, "-----\t---------\t--------\t------------")
	
	for _, item := range sortedRecords {
		fmt.Fprintf(cf.writer, "%d\t%s\t%s\t%s\n",
			item.Index,
			item.Record.Timestamp.Format("15:04:05.000"),
			item.Record.Distance.StringFixed(1),
			item.MileageDiff.StringFixed(1),
		)
	}
	
	return cf.writer.Flush()
}

// FormatProcessingResult formats and displays the complete processing result
func (cf *ConsoleFormatter) FormatProcessingResult(result models.ProcessingResult) error {
	if result.Error != nil {
		fmt.Fprintf(cf.output, "Processing failed: %v\n", result.Error)
		return nil
	}
	
	if !result.IsValid() {
		fmt.Fprint(cf.output, "Invalid processing result\n")
		return nil
	}
	
	// Display fare calculation
	if err := cf.FormatCurrentFare(result.Calculation); err != nil {
		return fmt.Errorf("error formatting fare: %w", err)
	}
	
	// Display processing summary
	fmt.Fprintf(cf.output, "\nProcessing Summary:\n")
	fmt.Fprintf(cf.output, "Records processed: %d\n", len(result.Records))
	fmt.Fprintf(cf.output, "Processing time: %v\n", result.TotalTime)
	fmt.Fprintf(cf.output, "Total fare: %d yen\n", result.Calculation.TotalFare.Round(0).IntPart())
	
	return nil
}

// FormatSummaryStatistics formats and displays summary statistics
func (cf *ConsoleFormatter) FormatSummaryStatistics(records []models.DistanceRecord, calculation models.FareCalculation) error {
	if len(records) == 0 {
		fmt.Fprint(cf.output, "No data for statistics\n")
		return nil
	}
	
	// Calculate statistics
	stats := calculateStatistics(records, calculation)
	
	// Format statistics using tabwriter
	fmt.Fprintln(cf.writer, "\nSummary Statistics")
	fmt.Fprintln(cf.writer, "------------------")
	fmt.Fprintf(cf.writer, "Total Records:\t%d\n", stats.TotalRecords)
	fmt.Fprintf(cf.writer, "Total Distance:\t%s km\n", stats.TotalDistance.StringFixed(3))
	fmt.Fprintf(cf.writer, "Average Distance:\t%s km\n", stats.AverageDistance.StringFixed(3))
	fmt.Fprintf(cf.writer, "Min Distance:\t%s km\n", stats.MinDistance.StringFixed(3))
	fmt.Fprintf(cf.writer, "Max Distance:\t%s km\n", stats.MaxDistance.StringFixed(3))
	fmt.Fprintf(cf.writer, "Base Fare:\t%d yen\n", calculation.BaseFare.Round(0).IntPart())
	fmt.Fprintf(cf.writer, "Distance Fare:\t%d yen\n", calculation.DistanceFare.Round(0).IntPart())
	fmt.Fprintf(cf.writer, "Time Fare:\t%d yen\n", calculation.TimeFare.Round(0).IntPart())
	fmt.Fprintf(cf.writer, "Total Fare:\t%d yen\n", calculation.TotalFare.Round(0).IntPart())
	
	return cf.writer.Flush()
}

// RecordWithDifference represents a record with its mileage difference
type RecordWithDifference struct {
	Record      models.DistanceRecord
	MileageDiff decimal.Decimal
	Index       int
}

// Statistics holds summary statistics for processed records
type Statistics struct {
	TotalRecords    int
	TotalDistance   decimal.Decimal
	AverageDistance decimal.Decimal
	MinDistance     decimal.Decimal
	MaxDistance     decimal.Decimal
}

// calculateStatistics computes summary statistics from records
func calculateStatistics(records []models.DistanceRecord, calculation models.FareCalculation) Statistics {
	if len(records) == 0 {
		return Statistics{}
	}
	
	stats := Statistics{
		TotalRecords:  len(records),
		MinDistance:   records[0].Distance,
		MaxDistance:   records[0].Distance,
		TotalDistance: decimal.Zero,
	}
	
	// Calculate min, max, and total
	for _, record := range records {
		stats.TotalDistance = stats.TotalDistance.Add(record.Distance)
		
		if record.Distance.LessThan(stats.MinDistance) {
			stats.MinDistance = record.Distance
		}
		
		if record.Distance.GreaterThan(stats.MaxDistance) {
			stats.MaxDistance = record.Distance
		}
	}
	
	// Calculate average
	if len(records) > 0 {
		stats.AverageDistance = stats.TotalDistance.Div(decimal.NewFromInt(int64(len(records))))
	}
	
	return stats
}

// CompactFormatter provides a minimal output format for production use
type CompactFormatter struct {
	output io.Writer
}

// NewCompactFormatter creates a formatter with minimal output
func NewCompactFormatter() OutputFormatter {
	return NewCompactFormatterWithOutput(os.Stdout)
}

// NewCompactFormatterWithOutput creates a compact formatter with custom output
func NewCompactFormatterWithOutput(output io.Writer) OutputFormatter {
	return &CompactFormatter{output: output}
}

// FormatCurrentFare formats the fare as a single integer
func (cf *CompactFormatter) FormatCurrentFare(calculation models.FareCalculation) error {
	totalFareInt := calculation.TotalFare.Round(0).IntPart()
	fmt.Fprintf(cf.output, "%d\n", totalFareInt)
	return nil
}

// FormatRecords formats records in a compact format
func (cf *CompactFormatter) FormatRecords(records []models.DistanceRecord) error {
	fmt.Fprintf(cf.output, "Records: %d\n", len(records))
	if len(records) > 0 {
		first := records[0]
		last := records[len(records)-1]
		totalDistance := last.Distance.Sub(first.Distance)
		fmt.Fprintf(cf.output, "Distance: %s\n", totalDistance.StringFixed(1))
	}
	return nil
}

// FormatProcessingResult formats the result compactly
func (cf *CompactFormatter) FormatProcessingResult(result models.ProcessingResult) error {
	if result.Error != nil {
		return result.Error
	}
	
	return cf.FormatCurrentFare(result.Calculation)
}

// FormatSummaryStatistics formats statistics compactly
func (cf *CompactFormatter) FormatSummaryStatistics(records []models.DistanceRecord, calculation models.FareCalculation) error {
	if len(records) == 0 {
		return nil
	}
	
	fmt.Fprintf(cf.output, "Records: %d, Fare: %d yen\n", 
		len(records), 
		calculation.TotalFare.Round(0).IntPart())
	return nil
}

// DebugFormatter provides detailed output for debugging purposes
type DebugFormatter struct {
	output io.Writer
	writer *tabwriter.Writer
}

// NewDebugFormatter creates a formatter with debug output
func NewDebugFormatter() OutputFormatter {
	return NewDebugFormatterWithOutput(os.Stdout)
}

// NewDebugFormatterWithOutput creates a debug formatter with custom output
func NewDebugFormatterWithOutput(output io.Writer) OutputFormatter {
	writer := tabwriter.NewWriter(output, 0, 8, 1, '\t', 0)
	return &DebugFormatter{
		output: output,
		writer: writer,
	}
}

// FormatCurrentFare formats the fare with detailed breakdown
func (df *DebugFormatter) FormatCurrentFare(calculation models.FareCalculation) error {
	fmt.Fprintln(df.writer, "Fare Breakdown:")
	fmt.Fprintln(df.writer, "Component\tAmount (yen)")
	fmt.Fprintln(df.writer, "---------\t-----------")
	fmt.Fprintf(df.writer, "Base Fare\t%d\n", calculation.BaseFare.Round(0).IntPart())
	fmt.Fprintf(df.writer, "Distance Fare\t%d\n", calculation.DistanceFare.Round(0).IntPart())
	fmt.Fprintf(df.writer, "Time Fare\t%d\n", calculation.TimeFare.Round(0).IntPart())
	fmt.Fprintln(df.writer, "---------\t-----------")
	fmt.Fprintf(df.writer, "Total\t%d\n", calculation.TotalFare.Round(0).IntPart())
	
	return df.writer.Flush()
}

// FormatRecords formats records with full details
func (df *DebugFormatter) FormatRecords(records []models.DistanceRecord) error {
	if len(records) == 0 {
		fmt.Fprint(df.output, "No records to display\n")
		return nil
	}
	
	fmt.Fprintln(df.writer, "\nDetailed Record Information:")
	fmt.Fprintln(df.writer, "Index\tTimestamp\tDistance\tMileage Diff\tCumulative")
	fmt.Fprintln(df.writer, "-----\t---------\t--------\t------------\t----------")
	
	for i, record := range records {
		diff := decimal.Zero
		if i > 0 {
			diff = record.Distance.Sub(records[i-1].Distance)
		}
		
		cumulative := record.Distance.Sub(records[0].Distance)
		
		fmt.Fprintf(df.writer, "%d\t%s\t%s\t%s\t%s\n",
			i,
			record.Timestamp.Format("15:04:05.000"),
			record.Distance.StringFixed(3),
			diff.StringFixed(3),
			cumulative.StringFixed(3),
		)
	}
	
	return df.writer.Flush()
}

// FormatProcessingResult formats the result with debug information
func (df *DebugFormatter) FormatProcessingResult(result models.ProcessingResult) error {
	fmt.Fprintf(df.output, "Processing Result Debug Information:\n")
	fmt.Fprintf(df.output, "=====================================\n")
	
	if result.Error != nil {
		fmt.Fprintf(df.output, "Error: %v\n", result.Error)
		return nil
	}
	
	fmt.Fprintf(df.output, "Records processed: %d\n", len(result.Records))
	fmt.Fprintf(df.output, "Processing time: %v\n", result.TotalTime)
	fmt.Fprintf(df.output, "Valid result: %t\n", result.IsValid())
	
	// Display fare breakdown
	if err := df.FormatCurrentFare(result.Calculation); err != nil {
		return fmt.Errorf("error formatting fare breakdown: %w", err)
	}
	
	// Display records
	return df.FormatRecords(result.Records)
}

// FormatSummaryStatistics formats statistics with debug details
func (df *DebugFormatter) FormatSummaryStatistics(records []models.DistanceRecord, calculation models.FareCalculation) error {
	if len(records) == 0 {
		fmt.Fprint(df.output, "No data for debug statistics\n")
		return nil
	}
	
	stats := calculateStatistics(records, calculation)
	
	fmt.Fprintln(df.writer, "\nDebug Statistics:")
	fmt.Fprintln(df.writer, "=================")
	fmt.Fprintf(df.writer, "Record Count:\t%d\n", stats.TotalRecords)
	fmt.Fprintf(df.writer, "Distance Range:\t%s - %s km\n", 
		stats.MinDistance.StringFixed(3), stats.MaxDistance.StringFixed(3))
	fmt.Fprintf(df.writer, "Total Distance:\t%s km\n", stats.TotalDistance.StringFixed(3))
	fmt.Fprintf(df.writer, "Average Distance:\t%s km\n", stats.AverageDistance.StringFixed(3))
	fmt.Fprintf(df.writer, "Distance Span:\t%s km\n", 
		stats.MaxDistance.Sub(stats.MinDistance).StringFixed(3))
	
	// Fare calculation details
	fmt.Fprintln(df.writer, "\nFare Calculation Details:")
	fmt.Fprintf(df.writer, "Base Component:\t%s yen\n", calculation.BaseFare.StringFixed(2))
	fmt.Fprintf(df.writer, "Distance Component:\t%s yen\n", calculation.DistanceFare.StringFixed(2))
	fmt.Fprintf(df.writer, "Time Component:\t%s yen\n", calculation.TimeFare.StringFixed(2))
	fmt.Fprintf(df.writer, "Total (precise):\t%s yen\n", calculation.TotalFare.StringFixed(2))
	fmt.Fprintf(df.writer, "Total (display):\t%d yen\n", calculation.TotalFare.Round(0).IntPart())
	
	return df.writer.Flush()
}