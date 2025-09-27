// diagnostics.go
package main

import (
	"context"
	"fmt"
	"os/exec"
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
