// context.go
package main

import (
"context"
"encoding/json"
"errors"
"fmt"
"os/exec"
"strings"
"time"
)

type KubeContextSummary struct {
Context           string            `json:"context"`
Namespace         string            `json:"namespace"`
PodPhaseCounts    map[string]int    `json:"pod_phase_counts"`
PodProblemCounts  map[string]int    `json:"pod_problem_counts"` // CrashLoopBackOff, ImagePullBackOff, etc.
DeploymentCount   int               `json:"deployment_count"`
ServiceCount      int               `json:"service_count"`
Warnings          []string          `json:"warnings,omitempty"`
RenderedOneLiner  string            `json:"-"`
}

func runKubectl(timeout time.Duration, args ...string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return "", "", fmt.Errorf("kubectl %v timed out after %s", strings.Join(args, " "), timeout)
	}
	if err != nil {
		// return both streams via CombinedOutput; stderr isn't separate here.
		return "", string(out), fmt.Errorf("kubectl %v failed: %w", strings.Join(args, " "), err)
	}
	return string(out), "", nil
}

func GetCurrentContext() (string, error) {
out, _, err := runKubectl(2*time.Second, "config", "current-context")
if err != nil {
return "", err
}
return strings.TrimSpace(out), nil
}

func GetCurrentNamespace() (string, error) {
out, _, err := runKubectl(2*time.Second, "config", "view", "--minify", "--output", "jsonpath={..namespace}")
if err != nil {
return "", err
}
ns := strings.TrimSpace(out)
if ns == "" {
ns = "default"
}
return ns, nil
}

func getJSON(timeout time.Duration, args ...string) (map[string]any, error) {
out, _, err := runKubectl(timeout, args...)
if err != nil {
// even if kubectl returns non-zero, try to parse out
}
var m map[string]any
if jerr := json.Unmarshal([]byte(out), &m); jerr != nil {
return nil, errors.New("failed to parse kubectl JSON output")
}
return m, nil
}

func countPods(ns string) (phaseCounts map[string]int, problemCounts map[string]int, _ error) {
phaseCounts = map[string]int{}
problemCounts = map[string]int{}
obj, err := getJSON(4*time.Second, "-n", ns, "get", "pods", "-o", "json")
if err != nil {
return phaseCounts, problemCounts, err
}
items, _ := obj["items"].([]any)
for _, it := range items {
m, _ := it.(map[string]any)
status, _ := m["status"].(map[string]any)
phase, _ := status["phase"].(string)
if phase != "" {
phaseCounts[phase]++
}
// scan containerStatuses for common waiting reasons
if css, ok := status["containerStatuses"].([]any); ok {
for _, csi := range css {
csm, _ := csi.(map[string]any)
if st, ok := csm["state"].(map[string]any); ok {
if wait, ok := st["waiting"].(map[string]any); ok {
if reason, _ := wait["reason"].(string); reason != "" {
problemCounts[reason]++
}
}
}
if lst, ok := csm["lastState"].(map[string]any); ok {
if t, ok := lst["terminated"].(map[string]any); ok {
if reason, _ := t["reason"].(string); reason != "" {
problemCounts[reason]++
}
}
}
}
}
}
return phaseCounts, problemCounts, nil
}

func countResources(ns, kind string) (int, error) {
obj, err := getJSON(3*time.Second, "-n", ns, "get", kind, "-o", "json")
if err != nil {
return 0, err
}
items, _ := obj["items"].([]any)
return len(items), nil
}

// BuildContextSummary fetches a compact, token-efficient summary for prompts.
func BuildContextSummary() (*KubeContextSummary, error) {
ctxName, err := GetCurrentContext()
if err != nil {
ctxName = "(unknown)"
}
ns, err := GetCurrentNamespace()
if err != nil {
ns = "default"
}
ph, probs, _ := countPods(ns)
depCount, _ := countResources(ns, "deployments")
svcCount, _ := countResources(ns, "services")

s := &KubeContextSummary{
Context:          ctxName,
Namespace:        ns,
PodPhaseCounts:   ph,
PodProblemCounts: probs,
DeploymentCount:  depCount,
ServiceCount:     svcCount,
}

// Render a tight one-liner for prompt injection.
var phParts []string
for k, v := range ph {
phParts = append(phParts, fmt.Sprintf("%s=%d", k, v))
}
var pbParts []string
for k, v := range probs {
// only show notable ones
if v > 0 && (strings.Contains(k, "BackOff") || strings.Contains(k, "Error")) {
pbParts = append(pbParts, fmt.Sprintf("%s=%d", k, v))
}
}
one := fmt.Sprintf("ctx=%s ns=%s pods:{%s}", s.Context, s.Namespace, strings.Join(phParts, ","))
if len(pbParts) > 0 {
one += fmt.Sprintf(" podProblems:{%s}", strings.Join(pbParts, ","))
}
one += fmt.Sprintf(" deploy=%d svc=%d", depCount, svcCount)
s.RenderedOneLiner = one
return s, nil
}
