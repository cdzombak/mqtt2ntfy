package main

import (
	"testing"
)

func TestSetupLogger(t *testing.T) {
	// Test that logger is created without error
	logger := SetupLogger(false)
	if logger == nil {
		t.Error("SetupLogger() returned nil logger")
	}

	verboseLogger := SetupLogger(true)
	if verboseLogger == nil {
		t.Error("SetupLogger() returned nil logger for verbose")
	}

	// Both should be valid loggers
	if logger.Handler() == nil {
		t.Error("Logger handler is nil")
	}
}
