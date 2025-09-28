// redaction_improved_test.go - Enhanced tests for redaction functionality
package main

import (
	"testing"
)

func TestRedactText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "JWT token redaction",
			input:    "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			expected: "Bearer [REDACTED_JWT_TOKEN]",
		},
		{
			name:     "password redaction",
			input:    "password=secret123",
			expected: "password=[REDACTED]",
		},
		{
			name:     "token redaction",
			input:    "token=abc123def456",
			expected: "token=[REDACTED]",
		},
		{
			name:     "secretKeyRef redaction",
			input:    "secretKeyRef:\n  name: my-secret\n  key: password",
			expected: "secretKeyRef:\n  name: my-secret\n  key: [REDACTED]",
		},
		{
			name:     "base64 redaction",
			input:    "data: SGVsbG8gV29ybGQ=",
			expected: "data: [REDACTED_BASE64]",
		},
		{
			name:     "multiple redactions",
			input:    "password=secret123 token=abc123 data: SGVsbG8gV29ybGQ=",
			expected: "password=[REDACTED] token=[REDACTED] data: [REDACTED_BASE64]",
		},
		{
			name:     "no redaction needed",
			input:    "kubectl get pods",
			expected: "kubectl get pods",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "API key redaction",
			input:    "api_key=sk-1234567890abcdef",
			expected: "api_key=[REDACTED]",
		},
		{
			name:     "Authorization header redaction",
			input:    "Authorization: Bearer sk-1234567890abcdef",
			expected: "Authorization: Bearer [REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactText(tt.input)
			if result != tt.expected {
				t.Errorf("RedactText() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRedactText_PreservesStructure(t *testing.T) {
	input := `apiVersion: v1
kind: Secret
metadata:
  name: my-secret
data:
  password: cGFzc3dvcmQxMjM=
  token: dG9rZW4xMjM=
stringData:
  username: admin
  password: secret123`
	
	result := RedactText(input)
	
	// Should preserve YAML structure
	if !contains(result, "apiVersion: v1") {
		t.Error("RedactText() should preserve YAML structure")
	}
	
	if !contains(result, "kind: Secret") {
		t.Error("RedactText() should preserve YAML structure")
	}
	
	if !contains(result, "metadata:") {
		t.Error("RedactText() should preserve YAML structure")
	}
	
	// Should redact sensitive data
	if contains(result, "cGFzc3dvcmQxMjM=") {
		t.Error("RedactText() should redact base64 encoded passwords")
	}
	
	if contains(result, "secret123") {
		t.Error("RedactText() should redact plaintext passwords")
	}
}

func TestRedactText_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "partial JWT token",
			input:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expected: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", // Should not redact partial JWT
		},
		{
			name:     "password in different case",
			input:    "PASSWORD=secret123",
			expected: "PASSWORD=[REDACTED]",
		},
		{
			name:     "token in different case",
			input:    "TOKEN=abc123",
			expected: "TOKEN=[REDACTED]",
		},
		{
			name:     "mixed case secretKeyRef",
			input:    "secretkeyref:\n  name: my-secret\n  key: password",
			expected: "secretkeyref:\n  name: my-secret\n  key: [REDACTED]",
		},
		{
			name:     "whitespace around equals",
			input:    "password = secret123",
			expected: "password = [REDACTED]",
		},
		{
			name:     "quoted values",
			input:    "password=\"secret123\"",
			expected: "password=\"[REDACTED]\"",
		},
		{
			name:     "single quoted values",
			input:    "password='secret123'",
			expected: "password='[REDACTED]'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactText(tt.input)
			if result != tt.expected {
				t.Errorf("RedactText() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRedactText_Performance(t *testing.T) {
	// Test with a large input to ensure performance is reasonable
	largeInput := ""
	for i := 0; i < 1000; i++ {
		largeInput += "password=secret123 token=abc123 "
	}
	
	result := RedactText(largeInput)
	
	// Should contain redacted values
	if !contains(result, "[REDACTED]") {
		t.Error("RedactText() should redact values in large input")
	}
	
	// Should not contain original secrets
	if contains(result, "secret123") {
		t.Error("RedactText() should not contain original secrets in large input")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsInMiddle(s, substr)))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
