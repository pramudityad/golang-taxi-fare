package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewApplication(t *testing.T) {
	app := NewApplication()
	
	if app == nil {
		t.Fatal("Expected non-nil Application")
	}
	
	if app.logger == nil {
		t.Error("Expected logger to be initialized")
	}
	
	if app.errorHandler == nil {
		t.Error("Expected errorHandler to be initialized")
	}
	
	if app.parser == nil {
		t.Error("Expected parser to be initialized")
	}
	
	if app.validator == nil {
		t.Error("Expected validator to be initialized")
	}
	
	if app.calculator == nil {
		t.Error("Expected calculator to be initialized")
	}
	
	if app.formatter == nil {
		t.Error("Expected formatter to be initialized")
	}
	
	if app.ctx == nil {
		t.Error("Expected context to be initialized")
	}
	
	if app.cancel == nil {
		t.Error("Expected cancel function to be initialized")
	}
}

func TestApplicationCleanup(t *testing.T) {
	app := NewApplication()
	
	// Test that cleanup doesn't panic
	app.Cleanup()
	
	// Test that context is cancelled
	select {
	case <-app.ctx.Done():
		// Expected - context was cancelled
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected context to be cancelled after cleanup")
	}
}

func TestIsCriticalError(t *testing.T) {
	app := NewApplication()
	
	tests := []struct {
		name        string
		err         error
		expectCritical bool
	}{
		{
			name:           "nil error",
			err:            nil,
			expectCritical: true, // Unknown error type
		},
		{
			name:           "generic error",
			err:            errors.New("generic error"),
			expectCritical: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := app.isCriticalError(tt.err)
			if result != tt.expectCritical {
				t.Errorf("Expected isCriticalError(%v) = %v, got %v", 
					tt.err, tt.expectCritical, result)
			}
		})
	}
}

func TestMainIntegration(t *testing.T) {
	// Redirect stdout to capture application output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// Redirect stderr to capture log output
	oldStderr := os.Stderr
	r2, w2, _ := os.Pipe()
	os.Stderr = w2
	
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()
	
	// Create test input
	testInput := `12:34:56.789 12345678.5
12:34:57.123 12345679.1
12:34:58.456 12345680.3`
	
	// Redirect stdin
	oldStdin := os.Stdin
	r3, w3, _ := os.Pipe()
	os.Stdin = r3
	go func() {
		defer w3.Close()
		w3.Write([]byte(testInput))
	}()
	defer func() {
		os.Stdin = oldStdin
	}()
	
	// Run main in goroutine to avoid os.Exit
	done := make(chan bool)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Handle any panics
				t.Logf("Main function panicked: %v", r)
			}
			done <- true
		}()
		
		app := NewApplication()
		err := app.Run()
		if err != nil {
			t.Logf("Application returned error: %v", err)
		}
	}()
	
	// Wait for completion with timeout
	select {
	case <-done:
		// Application completed
	case <-time.After(5 * time.Second):
		t.Fatal("Application timed out")
	}
	
	// Capture stdout
	w.Close()
	var stdout bytes.Buffer
	stdout.ReadFrom(r)
	
	// Capture stderr (logs)
	w2.Close()
	var stderr bytes.Buffer
	stderr.ReadFrom(r2)
	
	stdoutStr := stdout.String()
	stderrStr := stderr.String()
	
	// Verify fare calculation output
	if !strings.Contains(stdoutStr, "400") {
		t.Errorf("Expected stdout to contain fare '400', got: %s", stdoutStr)
	}
	
	// Verify processing summary
	if !strings.Contains(stdoutStr, "Processing Summary") {
		t.Errorf("Expected stdout to contain 'Processing Summary', got: %s", stdoutStr)
	}
	
	if !strings.Contains(stdoutStr, "Records processed: 3") {
		t.Errorf("Expected stdout to contain 'Records processed: 3', got: %s", stdoutStr)
	}
	
	// Verify structured logging
	if !strings.Contains(stderrStr, "\"level\":\"INFO\"") {
		t.Errorf("Expected stderr to contain structured JSON logs, got: %s", stderrStr)
	}
	
	if !strings.Contains(stderrStr, "\"component\":\"main\"") {
		t.Errorf("Expected stderr to contain main component logs, got: %s", stderrStr)
	}
	
	if !strings.Contains(stderrStr, "Starting taxi fare calculation processing") {
		t.Errorf("Expected stderr to contain startup message, got: %s", stderrStr)
	}
	
	if !strings.Contains(stderrStr, "Application completed successfully") {
		t.Errorf("Expected stderr to contain completion message, got: %s", stderrStr)
	}
}

func TestApplicationWithInvalidInput(t *testing.T) {
	app := NewApplication()
	
	// Redirect stdin with invalid input
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	go func() {
		defer w.Close()
		w.Write([]byte("invalid input\n"))
	}()
	defer func() {
		os.Stdin = oldStdin
	}()
	
	// This should handle errors gracefully
	err := app.Run()
	if err == nil {
		t.Error("Expected error when processing invalid input")
	}
}

func TestApplicationWithEmptyInput(t *testing.T) {
	app := NewApplication()
	
	// Redirect stdin with empty input
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	w.Close() // Close immediately to simulate empty input
	defer func() {
		os.Stdin = oldStdin
	}()
	
	// This should return error for insufficient data
	err := app.Run()
	if err == nil {
		t.Error("Expected error when processing empty input")
	}
	
	if !strings.Contains(err.Error(), "insufficient data") {
		t.Errorf("Expected error to mention insufficient data, got: %v", err)
	}
}

func TestApplicationContextCancellation(t *testing.T) {
	app := NewApplication()
	
	// Cancel context immediately
	app.cancel()
	
	// Run should return context error
	err := app.Run()
	if err == nil {
		t.Error("Expected error when context is cancelled")
	}
	
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got: %v", err)
	}
}