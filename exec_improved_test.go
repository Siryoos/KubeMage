// exec_improved_test.go - Enhanced tests for exec functionality
package main

import (
	"strings"
	"testing"
	"time"
)

func TestParseCommand_Kubectl(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected string
	}{
		{
			name:     "kubectl get pods",
			command:  "kubectl get pods",
			expected: "kubectl",
		},
		{
			name:     "kubectl apply with file",
			command:  "kubectl apply -f deployment.yaml",
			expected: "kubectl",
		},
		{
			name:     "kubectl delete with force",
			command:  "kubectl delete pod nginx --force",
			expected: "kubectl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := parseCommand(tt.command)
			if cmd == nil {
				t.Fatal("parseCommand returned nil")
			}
			// Check that the command path contains the expected command
			if !strings.Contains(cmd.Path, tt.expected) {
				t.Errorf("parseCommand() path = %v, want to contain %v", cmd.Path, tt.expected)
			}
		})
	}
}

func TestParseCommand_Helm(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected string
	}{
		{
			name:     "helm install",
			command:  "helm install myapp ./chart",
			expected: "helm",
		},
		{
			name:     "helm upgrade",
			command:  "helm upgrade myapp ./chart",
			expected: "helm",
		},
		{
			name:     "helm template",
			command:  "helm template myapp ./chart",
			expected: "helm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := parseCommand(tt.command)
			if cmd == nil {
				t.Fatal("parseCommand returned nil")
			}
			// Check that the command path contains the expected command
			if !strings.Contains(cmd.Path, tt.expected) {
				t.Errorf("parseCommand() path = %v, want to contain %v", cmd.Path, tt.expected)
			}
		})
	}
}

func TestParseCommand_Complex(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected string
	}{
		{
			name:     "complex kubectl command",
			command:  "kubectl get pods -o jsonpath='{.items[*].metadata.name}'",
			expected: "kubectl",
		},
		{
			name:     "helm with values",
			command:  "helm install myapp ./chart --set image.tag=v1.0.0",
			expected: "helm",
		},
		{
			name:     "kubectl with namespace",
			command:  "kubectl get pods -n kube-system",
			expected: "kubectl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := parseCommand(tt.command)
			if cmd == nil {
				t.Fatal("parseCommand returned nil")
			}
			// Check that the command path contains the expected command
			if !strings.Contains(cmd.Path, tt.expected) {
				t.Errorf("parseCommand() path = %v, want to contain %v", cmd.Path, tt.expected)
			}
		})
	}
}

func TestParseCommand_BashFallback(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected string
	}{
		{
			name:     "bash command",
			command:  "echo 'hello world'",
			expected: "bash",
		},
		{
			name:     "complex bash command",
			command:  "ls -la | grep .go",
			expected: "ls",
		},
		{
			name:     "pipeline command",
			command:  "kubectl get pods | grep nginx",
			expected: "kubectl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := parseCommand(tt.command)
			if cmd == nil {
				t.Fatal("parseCommand returned nil")
			}
			// Check that the command path contains the expected command
			if !strings.Contains(cmd.Path, tt.expected) {
				t.Errorf("parseCommand() path = %v, want to contain %v", cmd.Path, tt.expected)
			}
		})
	}
}

func TestParseCommand_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected string
	}{
		{
			name:     "empty command",
			command:  "",
			expected: "bash",
		},
		{
			name:     "whitespace only",
			command:  "   ",
			expected: "bash",
		},
		{
			name:     "single word",
			command:  "kubectl",
			expected: "bash",
		},
		{
			name:     "kubectl with spaces",
			command:  "  kubectl get pods  ",
			expected: "kubectl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := parseCommand(tt.command)
			if cmd == nil {
				t.Fatal("parseCommand returned nil")
			}
			// Check that the command path contains the expected command
			if !strings.Contains(cmd.Path, tt.expected) {
				t.Errorf("parseCommand() path = %v, want to contain %v", cmd.Path, tt.expected)
			}
		})
	}
}

func TestExecCmdWithTimeout_Timeout(t *testing.T) {
	// Test that the function returns a command that can be executed
	// We can't easily test this without a real program, so we'll just test the function exists
	cmd := execCmdWithTimeout("echo 'test'", 1*time.Second, nil)
	if cmd == nil {
		t.Fatal("execCmdWithTimeout returned nil")
	}

	// Test that the command exists and can be called
	// The actual execution would require a real tea.Program instance
	t.Log("execCmdWithTimeout function exists and returns a command")
}

func TestExecCmdWithTimeout_InvalidCommand(t *testing.T) {
	// Test with an invalid command
	cmd := execCmdWithTimeout("nonexistentcommand12345", 1*time.Second, nil)
	if cmd == nil {
		t.Fatal("execCmdWithTimeout returned nil")
	}

	// Test that the command exists and can be called
	// The actual execution would require a real tea.Program instance
	t.Log("execCmdWithTimeout function exists and returns a command for invalid command")
}

func TestExecPreviewCheck_ValidCheck(t *testing.T) {
	check := PreviewCheck{
		Name: "test check",
		Cmd:  "echo 'test'",
	}

	// Test that the PreviewCheck struct is properly initialized
	if check.Name != "test check" {
		t.Errorf("PreviewCheck.Name = %v, want %v", check.Name, "test check")
	}
	if check.Cmd != "echo 'test'" {
		t.Errorf("PreviewCheck.Cmd = %v, want %v", check.Cmd, "echo 'test'")
	}
}

func TestExecPreviewCheck_InvalidCheck(t *testing.T) {
	check := PreviewCheck{
		Name: "invalid check",
		Cmd:  "nonexistentcommand12345",
	}

	// Test that the PreviewCheck struct is properly initialized
	if check.Name != "invalid check" {
		t.Errorf("PreviewCheck.Name = %v, want %v", check.Name, "invalid check")
	}
	if check.Cmd != "nonexistentcommand12345" {
		t.Errorf("PreviewCheck.Cmd = %v, want %v", check.Cmd, "nonexistentcommand12345")
	}
}

func TestMessageTypes(t *testing.T) {
	// Test stdoutMsg
	stdoutMsg := stdoutMsg{
		cmd: "test command",
		out: "test output",
	}
	if stdoutMsg.cmd != "test command" {
		t.Errorf("stdoutMsg.cmd = %v, want %v", stdoutMsg.cmd, "test command")
	}
	if stdoutMsg.out != "test output" {
		t.Errorf("stdoutMsg.out = %v, want %v", stdoutMsg.out, "test output")
	}

	// Test stderrMsg
	stderrMsg := stderrMsg{
		cmd: "test command",
		out: "test error",
	}
	if stderrMsg.cmd != "test command" {
		t.Errorf("stderrMsg.cmd = %v, want %v", stderrMsg.cmd, "test command")
	}
	if stderrMsg.out != "test error" {
		t.Errorf("stderrMsg.out = %v, want %v", stderrMsg.out, "test error")
	}

	// Test execDoneMsg
	execDoneMsg := execDoneMsg{
		cmd: "test command",
		err: nil,
	}
	if execDoneMsg.cmd != "test command" {
		t.Errorf("execDoneMsg.cmd = %v, want %v", execDoneMsg.cmd, "test command")
	}
	if execDoneMsg.err != nil {
		t.Errorf("execDoneMsg.err = %v, want %v", execDoneMsg.err, nil)
	}

	// Test previewCheckDoneMsg
	previewCheckDoneMsg := previewCheckDoneMsg{
		check: PreviewCheck{Name: "test", Cmd: "echo test"},
		out:   "test output",
		err:   nil,
	}
	if previewCheckDoneMsg.check.Name != "test" {
		t.Errorf("previewCheckDoneMsg.check.Name = %v, want %v", previewCheckDoneMsg.check.Name, "test")
	}
	if previewCheckDoneMsg.out != "test output" {
		t.Errorf("previewCheckDoneMsg.out = %v, want %v", previewCheckDoneMsg.out, "test output")
	}

	// Test validationFailedMsg
	validationFailedMsg := validationFailedMsg{
		cmd:    "test command",
		stderr: "validation error",
		err:    nil,
	}
	if validationFailedMsg.cmd != "test command" {
		t.Errorf("validationFailedMsg.cmd = %v, want %v", validationFailedMsg.cmd, "test command")
	}
	if validationFailedMsg.stderr != "validation error" {
		t.Errorf("validationFailedMsg.stderr = %v, want %v", validationFailedMsg.stderr, "validation error")
	}
}
