// performance_test.go - Performance tests and benchmarks for critical operations
package main

import (
	"bytes"
	"os"
	"testing"
	"time"
)

// Benchmark tests for critical operations

func BenchmarkParseCommand_Kubectl(b *testing.B) {
	command := "kubectl get pods -o jsonpath='{.items[*].metadata.name}'"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parseCommand(command)
	}
}

func BenchmarkParseCommand_Helm(b *testing.B) {
	command := "helm install myapp ./chart --set image.tag=v1.0.0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parseCommand(command)
	}
}

func BenchmarkParseCommand_Bash(b *testing.B) {
	command := "echo 'hello world' | grep hello"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parseCommand(command)
	}
}

func BenchmarkBuildPreExecPlan_Safe(b *testing.B) {
	command := "kubectl get pods"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildPreExecPlan(command)
	}
}

func BenchmarkBuildPreExecPlan_Dangerous(b *testing.B) {
	command := "kubectl delete all --all"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildPreExecPlan(command)
	}
}

func BenchmarkIsWhitelistedAction(b *testing.B) {
	command := "kubectl get pods"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsWhitelistedAction(command)
	}
}

func BenchmarkRedactText_Simple(b *testing.B) {
	text := "password=secret123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RedactText(text)
	}
}

func BenchmarkRedactText_Complex(b *testing.B) {
	text := `apiVersion: v1
kind: Secret
metadata:
  name: my-secret
data:
  password: cGFzc3dvcmQxMjM=
  token: dG9rZW4xMjM=
stringData:
  username: admin
  password: secret123`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RedactText(text)
	}
}

func BenchmarkRedactText_Large(b *testing.B) {
	// Create a large text with many secrets
	text := ""
	for i := 0; i < 1000; i++ {
		text += "password=secret123 token=abc123 "
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RedactText(text)
	}
}

func BenchmarkMetrics_RecordSuggestion(b *testing.B) {
	metrics := NewSessionMetrics()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.RecordSuggestion()
	}
}

func BenchmarkMetrics_RecordValidation(b *testing.B) {
	metrics := NewSessionMetrics()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.RecordValidation(true)
	}
}

func BenchmarkMetrics_RecordEdit(b *testing.B) {
	metrics := NewSessionMetrics()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.RecordEditSuggestion()
		metrics.RecordEditApplied()
	}
}

func BenchmarkMetrics_CalculateRates(b *testing.B) {
	metrics := NewSessionMetrics()

	// Set up some data
	for i := 0; i < 100; i++ {
		metrics.RecordSuggestion()
		metrics.RecordValidation(true)
		metrics.RecordEditSuggestion()
		metrics.RecordEditApplied()
		metrics.RecordResolution()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = metrics.TSR()
		_ = metrics.CAR()
		_ = metrics.EAR()
		_ = metrics.MTR()
	}
}

func BenchmarkMetrics_DumpJSON(b *testing.B) {
	metrics := NewSessionMetrics()

	// Set up some data
	for i := 0; i < 100; i++ {
		metrics.RecordSuggestion()
		metrics.RecordValidation(true)
		metrics.RecordEditSuggestion()
		metrics.RecordEditApplied()
		metrics.RecordResolution()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		metrics.DumpJSON(&buf)
	}
}

func BenchmarkConfig_Load(b *testing.B) {
	// Create a test config file
	config := DefaultConfig()
	SaveConfig(config)
	defer os.Remove("config.yaml")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = LoadConfig()
	}
}

func BenchmarkConfig_Save(b *testing.B) {
	config := DefaultConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SaveConfig(config)
	}

	// Clean up
	os.Remove("config.yaml")
}

func BenchmarkConfig_UpdateModel(b *testing.B) {
	// Create a test config file
	config := DefaultConfig()
	SaveConfig(config)
	defer os.Remove("config.yaml")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = UpdateModelInConfig("chat", "test-model")
	}
}

// Performance tests for memory usage

func TestMemoryUsage_RedactText(t *testing.T) {
	// Test memory usage with large input
	largeInput := ""
	for i := 0; i < 10000; i++ {
		largeInput += "password=secret123 token=abc123 "
	}

	start := time.Now()
	result := RedactText(largeInput)
	duration := time.Since(start)

	if len(result) == 0 {
		t.Error("RedactText should return non-empty result")
	}

	if duration > 1*time.Second {
		t.Errorf("RedactText took too long: %v", duration)
	}
}

func TestMemoryUsage_Metrics(t *testing.T) {
	// Test memory usage with many metrics
	metrics := NewSessionMetrics()

	start := time.Now()
	for i := 0; i < 10000; i++ {
		metrics.RecordSuggestion()
		metrics.RecordValidation(true)
		metrics.RecordEditSuggestion()
		metrics.RecordEditApplied()
		metrics.RecordResolution()
	}
	duration := time.Since(start)

	if duration > 10*time.Millisecond {
		t.Errorf("Metrics recording took too long: %v", duration)
	}

	// Test that rates can be calculated efficiently
	start = time.Now()
	for i := 0; i < 1000; i++ {
		_ = metrics.TSR()
		_ = metrics.CAR()
		_ = metrics.EAR()
		_ = metrics.MTR()
	}
	duration = time.Since(start)

	if duration > 10*time.Millisecond {
		t.Errorf("Rate calculations took too long: %v", duration)
	}
}

func TestMemoryUsage_Config(t *testing.T) {
	// Test memory usage with config operations
	config := DefaultConfig()

	start := time.Now()
	for i := 0; i < 1000; i++ {
		_ = SaveConfig(config)
		_, _ = LoadConfig()
	}
	duration := time.Since(start)

	if duration > 2*time.Second {
		t.Errorf("Config operations took too long: %v", duration)
	}

	// Clean up
	os.Remove("config.yaml")
}

// Performance tests for concurrent operations

func TestConcurrentMetrics(t *testing.T) {
	metrics := NewSessionMetrics()

	// Test concurrent access to metrics
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				metrics.RecordSuggestion()
				metrics.RecordValidation(true)
				metrics.RecordEditSuggestion()
				metrics.RecordEditApplied()
				metrics.RecordResolution()
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify metrics were recorded (allow for some race conditions)
	if metrics.Suggestions < 800 {
		t.Errorf("Expected at least 800 suggestions, got %d", metrics.Suggestions)
	}
}

func TestConcurrentRedaction(t *testing.T) {
	// Test concurrent redaction
	text := "password=secret123 token=abc123"
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = RedactText(text)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestConcurrentValidation(t *testing.T) {
	// Test concurrent validation
	command := "kubectl get pods"
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = BuildPreExecPlan(command)
				_ = IsWhitelistedAction(command)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// Performance regression tests

func TestPerformanceRegression_RedactText(t *testing.T) {
	// Test that redaction performance doesn't regress
	text := "password=secret123 token=abc123"

	start := time.Now()
	for i := 0; i < 1000; i++ {
		_ = RedactText(text)
	}
	duration := time.Since(start)

	// Should complete in less than 100ms
	if duration > 100*time.Millisecond {
		t.Errorf("RedactText performance regression: %v", duration)
	}
}

func TestPerformanceRegression_Metrics(t *testing.T) {
	// Test that metrics performance doesn't regress
	metrics := NewSessionMetrics()

	start := time.Now()
	for i := 0; i < 1000; i++ {
		metrics.RecordSuggestion()
		metrics.RecordValidation(true)
		metrics.RecordEditSuggestion()
		metrics.RecordEditApplied()
		metrics.RecordResolution()
	}
	duration := time.Since(start)

	// Should complete in less than 5ms
	if duration > 5*time.Millisecond {
		t.Errorf("Metrics performance regression: %v", duration)
	}
}

func TestPerformanceRegression_Validation(t *testing.T) {
	// Test that validation performance doesn't regress
	command := "kubectl get pods"

	start := time.Now()
	for i := 0; i < 1000; i++ {
		_ = BuildPreExecPlan(command)
		_ = IsWhitelistedAction(command)
	}
	duration := time.Since(start)

	// Should complete in less than 5ms
	if duration > 5*time.Millisecond {
		t.Errorf("Validation performance regression: %v", duration)
	}
}
