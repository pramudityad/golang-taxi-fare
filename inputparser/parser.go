// Package inputparser provides functionality for parsing time-stamped distance records from stdin.
// It supports streaming processing with robust error handling and precise decimal arithmetic.
package inputparser

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"golang-taxi-fare/models"
)

// Parser defines the interface for parsing time-stamped distance records
type Parser interface {
	// ParseStream reads from the provided reader and returns a channel of DistanceRecord
	// The channel is closed when EOF is reached or an unrecoverable error occurs
	ParseStream(ctx context.Context, reader io.Reader) (<-chan ParseResult, error)
	
	// ParseLine parses a single line in the format "hh:mm:ss.fff xxxxxxxx.f"
	ParseLine(line string) (models.DistanceRecord, error)
}

// ParseResult represents the result of parsing a single line
type ParseResult struct {
	Record models.DistanceRecord
	Error  error
	Line   int // Line number for error reporting
}

// StreamParser implements the Parser interface with streaming capabilities
type StreamParser struct {
	// Configuration options can be added here in the future
}

// NewParser creates a new StreamParser instance
func NewParser() Parser {
	return &StreamParser{}
}

// ParsingError represents different types of parsing errors
type ParsingError struct {
	Type    ErrorType
	Message string
	Line    int
	Input   string
}

// ErrorType categorizes different parsing error types
type ErrorType int

const (
	// ErrorTypeFormat indicates a format validation error
	ErrorTypeFormat ErrorType = iota
	// ErrorTypeTimestamp indicates a timestamp parsing error
	ErrorTypeTimestamp
	// ErrorTypeDistance indicates a distance parsing error
	ErrorTypeDistance
	// ErrorTypeIO indicates an I/O error
	ErrorTypeIO
)

// Error implements the error interface
func (pe *ParsingError) Error() string {
	return fmt.Sprintf("parsing error at line %d: %s (input: %q)", 
		pe.Line, pe.Message, pe.Input)
}

// String returns a human-readable description of the error type
func (et ErrorType) String() string {
	switch et {
	case ErrorTypeFormat:
		return "format"
	case ErrorTypeTimestamp:
		return "timestamp"
	case ErrorTypeDistance:
		return "distance"
	case ErrorTypeIO:
		return "io"
	default:
		return "unknown"
	}
}

// timestampLayout defines the expected timestamp format
const timestampLayout = "15:04:05.000"

// distancePattern defines the regex pattern for distance validation (8+ digits, decimal point, 1+ fractional digits)
var distancePattern = regexp.MustCompile(`^\d{8,}\.\d+$`)

// linePattern defines the complete line format: timestamp single-space distance
var linePattern = regexp.MustCompile(`^(\d{2}:\d{2}:\d{2}\.\d{3}) (\d{8,}\.\d+)$`)

// parseTimestamp parses a timestamp string in the format "hh:mm:ss.fff"
func parseTimestamp(timestampStr string) (time.Time, error) {
	if timestampStr == "" {
		return time.Time{}, &ParsingError{
			Type:    ErrorTypeTimestamp,
			Message: "empty timestamp",
			Input:   timestampStr,
		}
	}
	
	// Parse using the expected layout
	parsedTime, err := time.Parse(timestampLayout, timestampStr)
	if err != nil {
		return time.Time{}, &ParsingError{
			Type:    ErrorTypeTimestamp,
			Message: fmt.Sprintf("invalid timestamp format, expected hh:mm:ss.fff: %v", err),
			Input:   timestampStr,
		}
	}
	
	return parsedTime, nil
}

// validateTimestampFormat performs additional validation on timestamp format
func validateTimestampFormat(timestampStr string) error {
	if len(timestampStr) != len(timestampLayout) {
		return &ParsingError{
			Type:    ErrorTypeTimestamp,
			Message: fmt.Sprintf("invalid timestamp length, expected %d characters, got %d", 
				len(timestampLayout), len(timestampStr)),
			Input:   timestampStr,
		}
	}
	
	// Check for required separators
	if len(timestampStr) >= 3 && timestampStr[2] != ':' {
		return &ParsingError{
			Type:    ErrorTypeTimestamp,
			Message: "missing colon separator at position 2",
			Input:   timestampStr,
		}
	}
	
	if len(timestampStr) >= 6 && timestampStr[5] != ':' {
		return &ParsingError{
			Type:    ErrorTypeTimestamp,
			Message: "missing colon separator at position 5",
			Input:   timestampStr,
		}
	}
	
	if len(timestampStr) >= 9 && timestampStr[8] != '.' {
		return &ParsingError{
			Type:    ErrorTypeTimestamp,
			Message: "missing dot separator at position 8",
			Input:   timestampStr,
		}
	}
	
	return nil
}

// parseTimestampWithValidation combines format validation and parsing
func parseTimestampWithValidation(timestampStr string) (time.Time, error) {
	// First validate the format structure
	if err := validateTimestampFormat(timestampStr); err != nil {
		return time.Time{}, err
	}
	
	// Then parse the timestamp
	return parseTimestamp(timestampStr)
}

// parseDistance parses a distance string using decimal.NewFromString for precision
func parseDistance(distanceStr string) (decimal.Decimal, error) {
	if distanceStr == "" {
		return decimal.Zero, &ParsingError{
			Type:    ErrorTypeDistance,
			Message: "empty distance",
			Input:   distanceStr,
		}
	}
	
	// Parse using decimal.NewFromString for precision
	distance, err := decimal.NewFromString(distanceStr)
	if err != nil {
		return decimal.Zero, &ParsingError{
			Type:    ErrorTypeDistance,
			Message: fmt.Sprintf("invalid distance format: %v", err),
			Input:   distanceStr,
		}
	}
	
	// Validate that distance is non-negative
	if distance.IsNegative() {
		return decimal.Zero, &ParsingError{
			Type:    ErrorTypeDistance,
			Message: "distance cannot be negative",
			Input:   distanceStr,
		}
	}
	
	return distance, nil
}

// validateDistanceFormat performs format validation on distance string
func validateDistanceFormat(distanceStr string) error {
	if !distancePattern.MatchString(distanceStr) {
		return &ParsingError{
			Type:    ErrorTypeDistance,
			Message: "invalid distance format, expected xxxxxxxx.f (8+ digits, decimal point, 1+ fractional digits)",
			Input:   distanceStr,
		}
	}
	return nil
}

// parseDistanceWithValidation combines format validation and parsing
func parseDistanceWithValidation(distanceStr string) (decimal.Decimal, error) {
	// First validate the format structure
	if err := validateDistanceFormat(distanceStr); err != nil {
		return decimal.Zero, err
	}
	
	// Then parse the distance
	return parseDistance(distanceStr)
}

// parseLine parses a single line in the format "hh:mm:ss.fff xxxxxxxx.f"
func parseLine(line string, lineNum int) (models.DistanceRecord, error) {
	// Skip blank lines
	line = strings.TrimSpace(line)
	if line == "" {
		return models.DistanceRecord{}, &ParsingError{
			Type:    ErrorTypeFormat,
			Message: "blank line",
			Line:    lineNum,
			Input:   line,
		}
	}
	
	// Validate overall line format
	matches := linePattern.FindStringSubmatch(line)
	if len(matches) != 3 {
		return models.DistanceRecord{}, &ParsingError{
			Type:    ErrorTypeFormat,
			Message: "invalid line format, expected 'hh:mm:ss.fff xxxxxxxx.f'",
			Line:    lineNum,
			Input:   line,
		}
	}
	
	timestampStr := matches[1]
	distanceStr := matches[2]
	
	// Parse timestamp using existing function
	timestamp, err := parseTimestampWithValidation(timestampStr)
	if err != nil {
		// Convert to include line number
		if pe, ok := err.(*ParsingError); ok {
			pe.Line = lineNum
		}
		return models.DistanceRecord{}, err
	}
	
	// Parse distance using existing function
	distance, err := parseDistanceWithValidation(distanceStr)
	if err != nil {
		// Convert to include line number
		if pe, ok := err.(*ParsingError); ok {
			pe.Line = lineNum
		}
		return models.DistanceRecord{}, err
	}
	
	return models.DistanceRecord{
		Timestamp: timestamp,
		Distance:  distance,
	}, nil
}

// ParseLine implements single line parsing for the Parser interface
func (sp *StreamParser) ParseLine(line string) (models.DistanceRecord, error) {
	return parseLine(line, 0) // Line number 0 for standalone parsing
}

// ParseStream implements streaming parsing with context support
func (sp *StreamParser) ParseStream(ctx context.Context, reader io.Reader) (<-chan ParseResult, error) {
	resultChan := make(chan ParseResult, 10) // Buffered channel for better performance
	
	go func() {
		defer close(resultChan)
		
		scanner := bufio.NewScanner(reader)
		lineNum := 0
		
		for scanner.Scan() {
			lineNum++
			
			// Check for context cancellation
			select {
			case <-ctx.Done():
				resultChan <- ParseResult{
					Record: models.DistanceRecord{},
					Error:  ctx.Err(),
					Line:   lineNum,
				}
				return
			default:
				// Continue processing
			}
			
			line := scanner.Text()
			
			// Skip blank lines silently
			if strings.TrimSpace(line) == "" {
				continue
			}
			
			// Parse the line
			record, err := parseLine(line, lineNum)
			
			result := ParseResult{
				Record: record,
				Error:  err,
				Line:   lineNum,
			}
			
			// Send result to channel
			select {
			case resultChan <- result:
				// Successfully sent
			case <-ctx.Done():
				// Context cancelled while sending
				return
			}
		}
		
		// Check for scanner errors
		if err := scanner.Err(); err != nil {
			select {
			case resultChan <- ParseResult{
				Record: models.DistanceRecord{},
				Error: &ParsingError{
					Type:    ErrorTypeIO,
					Message: fmt.Sprintf("scanner error: %v", err),
					Line:    lineNum,
					Input:   "",
				},
				Line: lineNum,
			}:
			case <-ctx.Done():
				// Context cancelled
			}
		}
	}()
	
	return resultChan, nil
}