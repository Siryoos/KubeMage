// metrics_improved_test.go - Enhanced tests for metrics functionality
package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestNewSessionMetrics(t *testing.T) {
	metrics := NewSessionMetrics()
	if metrics == nil {
		t.Fatal("NewSessionMetrics returned nil")
	}

	// Test initial values
	if metrics.Suggestions != 0 {
		t.Errorf("Initial suggestions = %v, want 0", metrics.Suggestions)
	}
	if metrics.ValidationsPassed != 0 {
		t.Errorf("Initial validationsPassed = %v, want 0", metrics.ValidationsPassed)
	}
	if metrics.ValidationsFailed != 0 {
		t.Errorf("Initial validationsFailed = %v, want 0", metrics.ValidationsFailed)
	}
	if metrics.EditsSuggested != 0 {
		t.Errorf("Initial editsSuggested = %v, want 0", metrics.EditsSuggested)
	}
	if metrics.EditsApplied != 0 {
		t.Errorf("Initial editsApplied = %v, want 0", metrics.EditsApplied)
	}
	if metrics.Resolutions != 0 {
		t.Errorf("Initial resolutions = %v, want 0", metrics.Resolutions)
	}
	if metrics.SafetyBlocks != 0 {
		t.Errorf("Initial safetyBlocks = %v, want 0", metrics.SafetyBlocks)
	}
}

func TestRecordSuggestion(t *testing.T) {
	metrics := NewSessionMetrics()
	initial := metrics.Suggestions

	metrics.RecordSuggestion()

	if metrics.Suggestions != initial+1 {
		t.Errorf("RecordSuggestion() suggestions = %v, want %v", metrics.Suggestions, initial+1)
	}
}

func TestRecordValidation(t *testing.T) {
	metrics := NewSessionMetrics()

	// Test recording passed validation
	metrics.RecordValidation(true)
	if metrics.ValidationsPassed != 1 {
		t.Errorf("RecordValidation(true) validationsPassed = %v, want 1", metrics.ValidationsPassed)
	}

	// Test recording failed validation
	metrics.RecordValidation(false)
	if metrics.ValidationsFailed != 1 {
		t.Errorf("RecordValidation(false) validationsFailed = %v, want 1", metrics.ValidationsFailed)
	}
	if metrics.ValidationsPassed != 1 {
		t.Errorf("RecordValidation(false) validationsPassed = %v, want 1", metrics.ValidationsPassed)
	}
}

func TestRecordEdit(t *testing.T) {
	metrics := NewSessionMetrics()

	// Test recording edit suggestion
	metrics.RecordEditSuggestion()
	if metrics.EditsSuggested != 1 {
		t.Errorf("RecordEditSuggestion() editsSuggested = %v, want 1", metrics.EditsSuggested)
	}

	// Test recording edit applied
	metrics.RecordEditApplied()
	if metrics.EditsApplied != 1 {
		t.Errorf("RecordEditApplied() editsApplied = %v, want 1", metrics.EditsApplied)
	}
}

func TestRecordResolution(t *testing.T) {
	metrics := NewSessionMetrics()
	initial := metrics.Resolutions

	metrics.RecordResolution()

	if metrics.Resolutions != initial+1 {
		t.Errorf("RecordResolution() resolutions = %v, want %v", metrics.Resolutions, initial+1)
	}
}

func TestRecordSafetyViolation(t *testing.T) {
	metrics := NewSessionMetrics()
	initial := metrics.SafetyBlocks

	metrics.RecordSafetyBlock()

	if metrics.SafetyBlocks != initial+1 {
		t.Errorf("RecordSafetyBlock() safetyBlocks = %v, want %v", metrics.SafetyBlocks, initial+1)
	}
}

func TestGetTaskSuccessRate(t *testing.T) {
	metrics := NewSessionMetrics()

	// Test with no suggestions
	rate := metrics.TSR()
	if rate != 0.0 {
		t.Errorf("TSR() with no suggestions = %v, want 0.0", rate)
	}

	// Test with suggestions and resolutions
	metrics.Suggestions = 10
	metrics.Resolutions = 7
	rate = metrics.TSR()
	expected := 70.0
	if rate != expected {
		t.Errorf("TSR() = %v, want %v", rate, expected)
	}
}

func TestGetCommandAccuracyRate(t *testing.T) {
	metrics := NewSessionMetrics()

	// Test with no validations
	rate := metrics.CAR()
	if rate != 0.0 {
		t.Errorf("CAR() with no validations = %v, want 0.0", rate)
	}

	// Test with validations
	metrics.ValidationsPassed = 8
	metrics.ValidationsFailed = 2
	rate = metrics.CAR()
	expected := 80.0
	if rate != expected {
		t.Errorf("CAR() = %v, want %v", rate, expected)
	}
}

func TestGetEditAccuracyRate(t *testing.T) {
	metrics := NewSessionMetrics()

	// Test with no edits
	rate := metrics.EAR()
	if rate != 0.0 {
		t.Errorf("EAR() with no edits = %v, want 0.0", rate)
	}

	// Test with edits
	metrics.EditsSuggested = 10
	metrics.EditsApplied = 6
	rate = metrics.EAR()
	expected := 60.0
	if rate != expected {
		t.Errorf("EAR() = %v, want %v", rate, expected)
	}
}

func TestGetMeanTurnsToResolution(t *testing.T) {
	metrics := NewSessionMetrics()

	// Test with no resolutions
	turns := metrics.MTR()
	if turns != 0.0 {
		t.Errorf("MTR() with no resolutions = %v, want 0.0", turns)
	}

	// Test with resolutions
	metrics.Resolutions = 5
	metrics.Suggestions = 20
	turns = metrics.MTR()
	expected := 4.0
	if turns != expected {
		t.Errorf("MTR() = %v, want %v", turns, expected)
	}
}

func TestGetSafetyViolationsBlocked(t *testing.T) {
	metrics := NewSessionMetrics()

	// Test initial value
	violations := metrics.SafetyBlocks
	if violations != 0 {
		t.Errorf("SafetyBlocks initial = %v, want 0", violations)
	}

	// Test after recording violations
	metrics.RecordSafetyBlock()
	metrics.RecordSafetyBlock()
	violations = metrics.SafetyBlocks
	if violations != 2 {
		t.Errorf("SafetyBlocks after recording = %v, want 2", violations)
	}
}

func TestDumpJSON(t *testing.T) {
	metrics := NewSessionMetrics()
	metrics.Suggestions = 5
	metrics.ValidationsPassed = 2
	metrics.ValidationsFailed = 1
	metrics.EditsSuggested = 4
	metrics.EditsApplied = 3
	metrics.Resolutions = 2
	metrics.SafetyBlocks = 1

	var buf bytes.Buffer
	metrics.DumpJSON(&buf)

	// Test that the output is valid JSON
	var result map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		t.Errorf("DumpJSON() output is not valid JSON: %v", err)
	}

	// Test that the output contains expected fields
	expectedFields := []string{
		"suggestions",
		"validations",
		"validation_passed",
		"edits",
		"edits_applied",
		"resolutions",
		"safety_violations_blocked",
		"task_success_rate",
		"command_accuracy_rate",
		"edit_accuracy_rate",
		"mean_turns_to_resolution",
	}

	for _, field := range expectedFields {
		if _, exists := result[field]; !exists {
			t.Errorf("DumpJSON() missing field: %s", field)
		}
	}
}

func TestMetricsIntegration(t *testing.T) {
	metrics := NewSessionMetrics()

	// Simulate a complete workflow
	metrics.RecordSuggestion()        // 1 suggestion
	metrics.RecordValidation(true)    // 1 validation passed
	metrics.RecordEditSuggestion()    // 1 edit suggested
	metrics.RecordEditApplied()       // 1 edit applied
	metrics.RecordResolution()        // 1 resolution

	// Test calculated rates
	tsr := metrics.TSR()
	if tsr != 100.0 {
		t.Errorf("Task success rate = %v, want 100.0", tsr)
	}

	car := metrics.CAR()
	if car != 100.0 {
		t.Errorf("Command accuracy rate = %v, want 100.0", car)
	}

	ear := metrics.EAR()
	if ear != 100.0 {
		t.Errorf("Edit accuracy rate = %v, want 100.0", ear)
	}

	mtr := metrics.MTR()
	if mtr != 1.0 {
		t.Errorf("Mean turns to resolution = %v, want 1.0", mtr)
	}
}

func TestMetricsEdgeCases(t *testing.T) {
	metrics := NewSessionMetrics()

	// Test division by zero cases
	tsr := metrics.TSR()
	if tsr != 0.0 {
		t.Errorf("Task success rate with zero suggestions = %v, want 0.0", tsr)
	}

	car := metrics.CAR()
	if car != 0.0 {
		t.Errorf("Command accuracy rate with zero validations = %v, want 0.0", car)
	}

	ear := metrics.EAR()
	if ear != 0.0 {
		t.Errorf("Edit accuracy rate with zero edits = %v, want 0.0", ear)
	}

	mtr := metrics.MTR()
	if mtr != 0.0 {
		t.Errorf("Mean turns to resolution with zero resolutions = %v, want 0.0", mtr)
	}
}
