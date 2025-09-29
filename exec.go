package main

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type stdoutMsg struct {
	cmd string
	out string
}
type stderrMsg struct {
	cmd string
	out string
}
type execDoneMsg struct {
	cmd string
	err error
}
type previewCheckDoneMsg struct {
	check PreviewCheck
	out   string
	err   error
}
type validationFailedMsg struct {
	cmd    string
	stderr string
	err    error
}

// parseCommand intelligently parses a command string into exec.Command format
func parseCommand(command string) *exec.Cmd {
	command = strings.TrimSpace(command)

	// Handle kubectl commands directly
	if strings.HasPrefix(command, "kubectl ") {
		args := strings.Fields(command)[1:] // Skip "kubectl"
		return exec.Command("kubectl", args...)
	}

	// Handle helm commands directly
	if strings.HasPrefix(command, "helm ") {
		args := strings.Fields(command)[1:] // Skip "helm"
		return exec.Command("helm", args...)
	}

	// Handle other known commands that don't need shell
	knownCommands := []string{"docker", "git", "ls", "cat", "grep", "awk", "sed"}
	fields := strings.Fields(command)
	if len(fields) > 0 {
		for _, known := range knownCommands {
			if fields[0] == known {
				return exec.Command(fields[0], fields[1:]...)
			}
		}
	}

	// Fallback to bash for complex commands
	return exec.Command("bash", "-c", command)
}

// execCmdWithTimeout executes a command with a timeout
func execCmdWithTimeout(command string, timeout time.Duration, p *tea.Program) tea.Cmd {
	return func() tea.Msg {
		// Track command execution for smart refresh
		if p != nil {
			p.Send(commandTrackedMsg{command: command})
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		cmd := parseCommand(command)
		cmd = exec.CommandContext(ctx, cmd.Path, cmd.Args[1:]...)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return execDoneMsg{cmd: command, err: fmt.Errorf("error creating stdout pipe: %w", err)}
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			return execDoneMsg{cmd: command, err: fmt.Errorf("error creating stderr pipe: %w", err)}
		}

		if err := cmd.Start(); err != nil {
			return execDoneMsg{cmd: command, err: fmt.Errorf("error starting command '%s': %w", command, err)}
		}

		var stderrOutput strings.Builder

		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				p.Send(stdoutMsg{cmd: command, out: scanner.Text()})
			}
		}()

		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				line := scanner.Text()
				stderrOutput.WriteString(line + "\n")
				p.Send(stderrMsg{cmd: command, out: line})
			}
		}()

		err = cmd.Wait()

		// Check for validation failures that need self-correction
		if err != nil && stderrOutput.Len() > 0 {
			return validationFailedMsg{
				cmd:    command,
				stderr: stderrOutput.String(),
				err:    err,
			}
		}

		return execDoneMsg{cmd: command, err: err}
	}
}

// execCmd wraps execCmdWithTimeout with a default timeout
func execCmd(command string, p *tea.Program) tea.Cmd {
	return execCmdWithTimeout(command, 30*time.Second, p)
}

// runPreviewCheck executes a single preview check
func runPreviewCheck(check PreviewCheck, p *tea.Program) tea.Cmd {
	return func() tea.Msg {
		cmd := parseCommand(check.Cmd)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cmd = exec.CommandContext(ctx, cmd.Path, cmd.Args[1:]...)
		output, err := cmd.CombinedOutput()

		return previewCheckDoneMsg{
			check: check,
			out:   string(output),
			err:   err,
		}
	}
}

// runPreviewChecks executes all preview checks for a plan
func runPreviewChecks(plan PreExecPlan, p *tea.Program) []tea.Cmd {
	var cmds []tea.Cmd
	for _, check := range plan.Checks {
		cmds = append(cmds, runPreviewCheck(check, p))
	}
	return cmds
}

// generateCorrectionPrompt creates a prompt for LLM to correct a failed command
func generateCorrectionPrompt(originalCmd, stderr string) string {
	return fmt.Sprintf("The following command failed with an error:\n\nCommand: %s\n\nError output:\n```\n%s\n```\n\nPlease analyze the error and propose a corrected, safe command that addresses the issue. The corrected command should:\n1. Fix the underlying problem indicated by the error\n2. Include appropriate safety measures (--dry-run when applicable)\n3. Be safe to execute in a production environment\n\nRespond with only the corrected command, no explanation.", originalCmd, stderr)
}
