package execx

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Runner defines the interface for command execution
type Runner interface {
	Run(ctx context.Context, name string, args ...string) (stdout string, stderr string, err error)
	RunCommand(ctx context.Context, command string) (stdout string, stderr string, err error)
}

// OSRunner implements Runner using os/exec
type OSRunner struct{}

// NewOSRunner creates a new OS command runner
func NewOSRunner() *OSRunner {
	return &OSRunner{}
}

// Run executes a command with the given name and arguments
func (r *OSRunner) Run(ctx context.Context, name string, args ...string) (string, string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	
	return stdout.String(), stderr.String(), err
}

// RunCommand parses and executes a command string
func (r *OSRunner) RunCommand(ctx context.Context, command string) (string, string, error) {
	command = strings.TrimSpace(command)
	
	// Parse the command intelligently
	var cmd *exec.Cmd
	
	// Handle kubectl commands directly
	if strings.HasPrefix(command, "kubectl ") {
		args := strings.Fields(command)[1:] // Skip "kubectl"
		cmd = exec.CommandContext(ctx, "kubectl", args...)
	} else if strings.HasPrefix(command, "helm ") {
		// Handle helm commands directly
		args := strings.Fields(command)[1:] // Skip "helm"
		cmd = exec.CommandContext(ctx, "helm", args...)
	} else {
		// For other commands, use shell
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	
	return stdout.String(), stderr.String(), err
}

// MockRunner implements Runner for testing
type MockRunner struct {
	Commands []struct {
		Name   string
		Args   []string
		Stdout string
		Stderr string
		Err    error
	}
	CallCount int
}

// NewMockRunner creates a new mock runner for testing
func NewMockRunner() *MockRunner {
	return &MockRunner{
		Commands: make([]struct {
			Name   string
			Args   []string
			Stdout string
			Stderr string
			Err    error
		}, 0),
	}
}

// Run executes a mocked command
func (m *MockRunner) Run(ctx context.Context, name string, args ...string) (string, string, error) {
	if m.CallCount >= len(m.Commands) {
		return "", "", fmt.Errorf("unexpected command: %s %v", name, args)
	}
	
	cmd := m.Commands[m.CallCount]
	m.CallCount++
	
	return cmd.Stdout, cmd.Stderr, cmd.Err
}

// RunCommand executes a mocked command string
func (m *MockRunner) RunCommand(ctx context.Context, command string) (string, string, error) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", "", fmt.Errorf("empty command")
	}
	
	return m.Run(ctx, parts[0], parts[1:]...)
}