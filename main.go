package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang-taxi-fare/datavalidator"
	"golang-taxi-fare/errorhandler"
	"golang-taxi-fare/farecalculator"
	"golang-taxi-fare/inputparser"
	"golang-taxi-fare/loggingsystem"
	"golang-taxi-fare/models"
	"golang-taxi-fare/outputformatter"
)

// Application represents the main taxi fare calculator application
type Application struct {
	logger       loggingsystem.Logger
	errorHandler errorhandler.ErrorHandler
	parser       inputparser.Parser
	validator    datavalidator.Validator
	calculator   farecalculator.Calculator
	formatter    outputformatter.OutputFormatter
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewApplication creates and initializes a new Application instance
func NewApplication() *Application {
	ctx, cancel := context.WithCancel(context.Background())
	
	logger := loggingsystem.NewLogger()
	errorHandler := errorhandler.NewErrorHandler()
	parser := inputparser.NewParser()
	validator := datavalidator.NewValidator()
	calculator := farecalculator.NewCalculator()
	formatter := outputformatter.NewFormatter()

	return &Application{
		logger:       logger,
		errorHandler: errorHandler,
		parser:       parser,
		validator:    validator,
		calculator:   calculator,
		formatter:    formatter,
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Run executes the main application processing loop
func (app *Application) Run() error {
	startTime := time.Now()
	
	// Setup signal handling for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		sig := <-signalChan
		app.logger.WithComponent("main").Info("Received shutdown signal",
			"signal", sig.String(),
		)
		app.cancel()
	}()
	
	app.logger.WithComponent("main").Info("Starting taxi fare calculation processing")
	loggingsystem.LogProcessingStart(app.logger.WithComponent("main"), 0)
	
	// Parse input records from stdin
	parseResultChan, err := app.parser.ParseStream(app.ctx, os.Stdin)
	if err != nil {
		app.logger.WithComponent("parser").Error("Failed to start parsing stream", "error", err.Error())
		app.errorHandler.HandleError(err)
		return err
	}
	
	var records []models.DistanceRecord
	var processingErrors []error
	recordCount := 0
	
	// Process records from the input stream
	for {
		select {
		case <-app.ctx.Done():
			app.logger.WithComponent("main").Info("Processing cancelled by user")
			return app.ctx.Err()
			
		case parseResult, ok := <-parseResultChan:
			if !ok {
				// Channel closed, processing complete
				goto ProcessComplete
			}
			
			recordCount++
			
			// Check for parsing error
			if parseResult.Error != nil {
				loggingsystem.LogParsingError(app.logger.WithComponent("parser"), 
					parseResult.Line, "parsing_error", parseResult.Error.Error())
				processingErrors = append(processingErrors, parseResult.Error)
				
				// Handle critical parsing errors
				if app.isCriticalError(parseResult.Error) {
					app.errorHandler.HandleError(parseResult.Error)
					return parseResult.Error
				}
				continue
			}
			
			// Validate individual record
			if err := app.validator.ValidateRecord(parseResult.Record); err != nil {
				loggingsystem.LogValidationError(app.logger.WithComponent("validator"), 
					recordCount-1, "record_validation", err.Error())
				processingErrors = append(processingErrors, err)
				continue
			}
			
			records = append(records, parseResult.Record)
		}
	}

ProcessComplete:
	processingTime := time.Since(startTime)
	
	app.logger.WithComponent("main").Info("Input processing completed", 
		"total_records", len(records),
		"processing_errors", len(processingErrors),
		"processing_time_ms", processingTime.Milliseconds(),
	)
	
	// Validate the complete sequence of records
	if len(records) == 0 {
		err := errors.New("insufficient data: no valid records processed")
		app.errorHandler.HandleError(err)
		return err
	}
	
	if err := app.validator.ValidateSequence(records); err != nil {
		loggingsystem.LogValidationError(app.logger.WithComponent("validator"), 
			-1, "sequence_validation", err.Error())
		app.errorHandler.HandleError(err)
		return err
	}
	
	// Calculate fare from processed records
	calculation := app.calculator.CalculateFromRecords(records)
	
	loggingsystem.LogCalculationResult(app.logger.WithComponent("calculator"), 
		calculation.TotalFare, len(records))
	
	// Create processing result
	result := models.ProcessingResult{
		Records:     records,
		Calculation: calculation,
		TotalTime:   processingTime,
		Error:       nil,
	}
	
	// Format and display the result
	if err := app.formatter.FormatProcessingResult(result); err != nil {
		app.logger.WithComponent("formatter").Error("Output formatting failed", "error", err.Error())
		app.errorHandler.HandleError(err)
		return err
	}
	
	loggingsystem.LogProcessingComplete(app.logger.WithComponent("main"), 
		len(records), processingTime)
	
	return nil
}

// isCriticalError determines if an error should stop processing
func (app *Application) isCriticalError(err error) bool {
	switch err.(type) {
	case *inputparser.ParsingError:
		// Continue processing on parsing errors for individual lines
		return false
	case *datavalidator.ValidationError:
		// Continue processing on validation errors for individual records
		return false
	default:
		// Stop processing on unknown errors
		return true
	}
}

// Cleanup performs graceful cleanup of application resources
func (app *Application) Cleanup() {
	app.logger.WithComponent("main").Info("Performing application cleanup")
	
	if app.cancel != nil {
		app.cancel()
	}
	
	// Additional cleanup logic could go here
	// For example: closing database connections, flushing buffers, etc.
}

func main() {
	app := NewApplication()
	defer app.Cleanup()
	
	// Run the application
	if err := app.Run(); err != nil {
		// Error handling is managed by the error handler which calls os.Exit
		// This should not be reached in normal circumstances
		app.logger.WithComponent("main").Error("Application terminated with error", "error", err.Error())
	}
	
	app.logger.WithComponent("main").Info("Application completed successfully")
}