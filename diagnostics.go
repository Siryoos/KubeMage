package main

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type DiagResult struct {
	Step    string
	Command string
	Output  string // truncated for UI
	Notes   []string
}

type DiagPlan struct {
	Title   string
	Steps   []string // kubectl commands (read-only)
	Summary string   // optional LLM prompt suffix / human hint
}

// Execs a command with timeout and returns combined output (truncated).
func runShell(timeout time.Duration, command string, maxBytes int) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	out, err := cmd.CombinedOutput()
	text := string(out)
	if len(text) > maxBytes {
		text = text[len(text)-maxBytes:]
		text = "(…truncated…)\n" + text
	}
	if ctx.Err() == context.DeadlineExceeded {
		return text, fmt.Errorf("timeout after %s", timeout)
	}
	return text, err
}

// Heuristic plan for a pod that isn't Ready / is CrashLooping.
func PlanPodNotReady(pod, ns string) DiagPlan {
	base := fmt.Sprintf("-n %s", ns)
	return DiagPlan{
		Title: fmt.Sprintf("Pod %s not Ready", pod),
		Steps: []string{
			fmt.Sprintf("kubectl describe pod %s %s", pod, base),
			fmt.Sprintf("kubectl get events %s --field-selector involvedObject.kind=Pod,involvedObject.name=%s --sort-by=.lastTimestamp", base, pod),
			fmt.Sprintf("kubectl logs %s %s --tail=200 --all-containers", pod, base),
		},
		Summary: "Analyze describe/events/logs to infer root cause (ImagePullBackOff, CrashLoopBackOff, OOMKilled, probe failures, etc.).",
	}
}

// Runs a DiagPlan and returns per-step outputs.
func RunDiagPlan(p DiagPlan) ([]DiagResult, error) {
	results := make([]DiagResult, 0, len(p.Steps))
	for _, c := range p.Steps {
		out, err := runShell(8*time.Second, c, 32*1024) // cap to 32KB per step
		step := c
		if i := strings.Index(c, " "); i > 0 {
			step = c[:i]
		}
		dr := DiagResult{Step: step, Command: c, Output: out}
		if err != nil {
			dr.Notes = append(dr.Notes, "error: "+err.Error())
		}
		// quick heuristics
		if strings.Contains(out, "ImagePullBackOff") {
			dr.Notes = append(dr.Notes, "Detected ImagePullBackOff — check image name/registry/credentials/network.")
		}
		if strings.Contains(out, "CrashLoopBackOff") {
			dr.Notes = append(dr.Notes, "Detected CrashLoopBackOff — inspect container logs; check readiness/liveness probes and app startup.")
		}
		if strings.Contains(out, "OOMKilled") {
			dr.Notes = append(dr.Notes, "Container OOMKilled — consider limits/requests and memory usage.")
		}
		results = append(results, dr)
	}
	return results, nil
}

// Convenience wrapper for a single pod.
func DiagnosePodNotReady(pod, ns string) ([]DiagResult, error) {
	plan := PlanPodNotReady(pod, ns)
	return RunDiagPlan(plan)
}

// PlanServiceNotReachable creates a diagnostic plan for service connectivity issues
func PlanServiceNotReachable(svc, ns string) DiagPlan {
	base := fmt.Sprintf("-n %s", ns)
	return DiagPlan{
		Title: fmt.Sprintf("Service %s not reachable", svc),
		Steps: []string{
			fmt.Sprintf("kubectl describe service %s %s", svc, base),
			fmt.Sprintf("kubectl get endpoints %s %s", svc, base),
			fmt.Sprintf("kubectl get pods %s --show-labels", base),
			fmt.Sprintf("kubectl get events %s --field-selector involvedObject.kind=Service,involvedObject.name=%s", base, svc),
		},
		Summary: "Check service endpoints, pod selectors, and networking issues.",
	}
}

// PlanDeploymentNotReady creates a diagnostic plan for deployment issues
func PlanDeploymentNotReady(deploy, ns string) DiagPlan {
	base := fmt.Sprintf("-n %s", ns)
	return DiagPlan{
		Title: fmt.Sprintf("Deployment %s not ready", deploy),
		Steps: []string{
			fmt.Sprintf("kubectl describe deployment %s %s", deploy, base),
			fmt.Sprintf("kubectl get rs %s", base),
			fmt.Sprintf("kubectl get pods %s --show-labels", base),
			fmt.Sprintf("kubectl get events %s --field-selector involvedObject.kind=Deployment,involvedObject.name=%s", base, deploy),
		},
		Summary: "Analyze deployment rollout, replica sets, and pod readiness.",
	}
}

// Convenience wrapper for service issues.
func DiagnoseServiceNotReachable(svc, ns string) ([]DiagResult, error) {
	plan := PlanServiceNotReachable(svc, ns)
	return RunDiagPlan(plan)
}

// Convenience wrapper for deployment issues.
func DiagnoseDeploymentNotReady(deploy, ns string) ([]DiagResult, error) {
	plan := PlanDeploymentNotReady(deploy, ns)
	return RunDiagPlan(plan)
}

// ReAct-lite agent functionality
var (
	// Whitelisted read-only kubectl commands
	reWhitelistedKubectl = regexp.MustCompile(`^kubectl\s+(get|describe|logs|top|api-resources|version|explain)\b`)
	// Whitelisted read-only helm commands
	reWhitelistedHelm = regexp.MustCompile(`^helm\s+(lint|template|version|show|get)\b`)
	// Action pattern matching
	reActionPattern = regexp.MustCompile(`(?i)^Action:\s*(.+)$`)
	// Final pattern matching
	reFinalPattern = regexp.MustCompile(`(?i)^Final:\s*(.+)$`)
)

// ReActStep represents a single step in the ReAct-lite loop
type ReActStep struct {
	Action      string
	Observation string
	Allowed     bool
	Error       string
}

// ReActSession manages a ReAct-lite diagnostic session
type ReActSession struct {
	MaxSteps    int
	Steps       []ReActStep
	CurrentStep int
	Completed   bool
	FinalAnswer string
}

// NewReActSession creates a new ReAct-lite session
func NewReActSession(maxSteps int) *ReActSession {
	return &ReActSession{
		MaxSteps: maxSteps,
		Steps:    make([]ReActStep, 0),
	}
}

// IsWhitelistedAction checks if an action is safe to execute
func IsWhitelistedAction(action string) bool {
	action = strings.TrimSpace(action)
	return reWhitelistedKubectl.MatchString(action) || reWhitelistedHelm.MatchString(action)
}

// ProcessModelResponse parses model output for Action: or Final: statements
func (rs *ReActSession) ProcessModelResponse(response string) error {
	if rs.Completed {
		return fmt.Errorf("session already completed")
	}

	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check for Final: statement
		if matches := reFinalPattern.FindStringSubmatch(line); len(matches) > 1 {
			rs.FinalAnswer = strings.TrimSpace(matches[1])
			rs.Completed = true
			return nil
		}

		// Check for Action: statement
		if matches := reActionPattern.FindStringSubmatch(line); len(matches) > 1 {
			action := strings.TrimSpace(matches[1])
			return rs.ExecuteAction(action)
		}
	}

	return fmt.Errorf("no valid Action: or Final: found in response")
}

// ExecuteAction safely executes a whitelisted action
func (rs *ReActSession) ExecuteAction(action string) error {
	if rs.CurrentStep >= rs.MaxSteps {
		return fmt.Errorf("maximum steps (%d) reached", rs.MaxSteps)
	}

	step := ReActStep{Action: action}

	// Check if action is whitelisted
	if !IsWhitelistedAction(action) {
		step.Allowed = false
		step.Error = "Action not whitelisted - only read-only kubectl (get|describe|logs|top|api-resources|version|explain) and helm (lint|template|version|show|get) commands allowed"
		rs.Steps = append(rs.Steps, step)
		return fmt.Errorf("%s", step.Error)
	}

	step.Allowed = true

	// Execute the action safely
	output, err := runShell(8*time.Second, action, 32*1024)
	if err != nil {
		step.Error = err.Error()
		step.Observation = fmt.Sprintf("Error executing command: %v\nOutput: %s", err, output)
	} else {
		step.Observation = output
	}

	rs.Steps = append(rs.Steps, step)
	rs.CurrentStep++

	return nil
}

// GetLastObservation returns the most recent observation for feeding back to the model
func (rs *ReActSession) GetLastObservation() string {
	if len(rs.Steps) == 0 {
		return ""
	}

	lastStep := rs.Steps[len(rs.Steps)-1]
	if !lastStep.Allowed {
		return fmt.Sprintf("ERROR: %s", lastStep.Error)
	}

	observation := "Observation:\n" + lastStep.Observation
	if lastStep.Error != "" {
		observation += fmt.Sprintf("\nError: %s", lastStep.Error)
	}

	return observation
}

// GetSessionSummary returns a summary of the entire ReAct session
func (rs *ReActSession) GetSessionSummary() string {
	var summary strings.Builder

	summary.WriteString(fmt.Sprintf("ReAct Session Summary (%d/%d steps)\n", len(rs.Steps), rs.MaxSteps))
	summary.WriteString("==========================================\n\n")

	for i, step := range rs.Steps {
		summary.WriteString(fmt.Sprintf("Step %d:\n", i+1))
		summary.WriteString(fmt.Sprintf("Action: %s\n", step.Action))

		if !step.Allowed {
			summary.WriteString(fmt.Sprintf("❌ BLOCKED: %s\n", step.Error))
		} else {
			summary.WriteString("✅ EXECUTED\n")
			if step.Error != "" {
				summary.WriteString(fmt.Sprintf("⚠️  Error: %s\n", step.Error))
			}
			// Truncate observation for summary
			obs := step.Observation
			if len(obs) > 500 {
				obs = obs[:500] + "\n...(truncated for summary)..."
			}
			summary.WriteString(fmt.Sprintf("Observation: %s\n", obs))
		}
		summary.WriteString("\n")
	}

	if rs.Completed {
		summary.WriteString(fmt.Sprintf("✅ COMPLETED\nFinal Answer: %s\n", rs.FinalAnswer))
	} else {
		summary.WriteString("🔄 Session ongoing...\n")
	}

	return summary.String()
}