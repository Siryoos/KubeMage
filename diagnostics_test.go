package main

import (
	"strings"
	"testing"
	"time"
)

func TestPlanPodNotReady(t *testing.T) {
	plan := PlanPodNotReady("test-pod", "default")

	// Verify plan structure
	if plan.Title != "Pod test-pod not Ready" {
		t.Errorf("expected title 'Pod test-pod not Ready', got '%s'", plan.Title)
	}

	if len(plan.Steps) != 3 {
		t.Errorf("expected 3 steps, got %d", len(plan.Steps))
	}

	// Verify kubectl commands are properly formatted
	expectedCommands := []string{
		"kubectl describe pod test-pod -n default",
		"kubectl get events -n default --field-selector involvedObject.kind=Pod,involvedObject.name=test-pod --sort-by=.lastTimestamp",
		"kubectl logs test-pod -n default --tail=200 --all-containers",
	}

	for i, expected := range expectedCommands {
		if plan.Steps[i] != expected {
			t.Errorf("step %d: expected '%s', got '%s'", i, expected, plan.Steps[i])
		}
	}

	// Verify summary
	if !strings.Contains(plan.Summary, "root cause") {
		t.Errorf("expected summary to mention 'root cause', got '%s'", plan.Summary)
	}
}

func TestPlanServiceNotReachable(t *testing.T) {
	plan := PlanServiceNotReachable("test-svc", "kube-system")

	if plan.Title != "Service test-svc not reachable" {
		t.Errorf("expected title 'Service test-svc not reachable', got '%s'", plan.Title)
	}

	if len(plan.Steps) != 4 {
		t.Errorf("expected 4 steps, got %d", len(plan.Steps))
	}

	// Check namespace is included in commands
	for _, step := range plan.Steps {
		if !strings.Contains(step, "-n kube-system") {
			t.Errorf("expected step to include namespace flag, got '%s'", step)
		}
	}
}

func TestPlanDeploymentNotReady(t *testing.T) {
	plan := PlanDeploymentNotReady("test-deploy", "production")

	if plan.Title != "Deployment test-deploy not ready" {
		t.Errorf("expected title 'Deployment test-deploy not ready', got '%s'", plan.Title)
	}

	if len(plan.Steps) != 4 {
		t.Errorf("expected 4 steps, got %d", len(plan.Steps))
	}

	// Check namespace is included in commands
	for _, step := range plan.Steps {
		if !strings.Contains(step, "-n production") {
			t.Errorf("expected step to include namespace flag, got '%s'", step)
		}
	}
}

func TestIsWhitelistedAction(t *testing.T) {
	testCases := []struct {
		action   string
		expected bool
	}{
		{"kubectl get pods", true},
		{"kubectl describe pod test", true},
		{"kubectl logs test-pod", true},
		{"kubectl top nodes", true},
		{"kubectl api-resources", true},
		{"kubectl version", true},
		{"kubectl explain pod", true},
		{"helm lint chart/", true},
		{"helm template release chart/", true},
		{"helm version", true},
		{"helm show values chart/", true},
		{"helm get values release", true},
		{"kubectl delete pod test", false},
		{"kubectl apply -f test.yaml", false},
		{"helm install release chart/", false},
		{"helm upgrade release chart/", false},
		{"rm -rf /", false},
		{"curl malicious-site.com", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.action, func(t *testing.T) {
			result := IsWhitelistedAction(tc.action)
			if result != tc.expected {
				t.Errorf("IsWhitelistedAction(%q) = %v, expected %v", tc.action, result, tc.expected)
			}
		})
	}
}

func TestReActSessionBasics(t *testing.T) {
	session := NewReActSession(3)

	if session.MaxSteps != 3 {
		t.Errorf("expected MaxSteps=3, got %d", session.MaxSteps)
	}

	if session.CurrentStep != 0 {
		t.Errorf("expected CurrentStep=0, got %d", session.CurrentStep)
	}

	if session.Completed {
		t.Error("expected session not to be completed initially")
	}
}

func TestReActSessionExecuteAction_NotWhitelisted(t *testing.T) {
	session := NewReActSession(3)

	err := session.ExecuteAction("kubectl delete pod dangerous")
	if err == nil {
		t.Error("expected error for non-whitelisted action")
	}

	if len(session.Steps) != 1 {
		t.Errorf("expected 1 step, got %d", len(session.Steps))
	}

	step := session.Steps[0]
	if step.Allowed {
		t.Error("expected step to be blocked")
	}

	if !strings.Contains(step.Error, "not whitelisted") {
		t.Errorf("expected error to mention 'not whitelisted', got '%s'", step.Error)
	}
}

func TestReActSessionProcessModelResponse_Final(t *testing.T) {
	session := NewReActSession(3)

	response := "After analyzing the data, I can conclude that:\n\nFinal: The pod is failing due to image pull errors and needs valid credentials."

	err := session.ProcessModelResponse(response)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !session.Completed {
		t.Error("expected session to be completed")
	}

	expectedFinal := "The pod is failing due to image pull errors and needs valid credentials."
	if session.FinalAnswer != expectedFinal {
		t.Errorf("expected final answer '%s', got '%s'", expectedFinal, session.FinalAnswer)
	}
}

func TestReActSessionProcessModelResponse_Action(t *testing.T) {
	session := NewReActSession(3)

	response := "Let me check the pod status:\n\nAction: kubectl get pods -n default"

	err := session.ProcessModelResponse(response)
	// This may fail because kubectl might not be available in test environment, but we can check the parsing

	if len(session.Steps) != 1 {
		t.Errorf("expected 1 step, got %d", len(session.Steps))
	}

	step := session.Steps[0]
	if step.Action != "kubectl get pods -n default" {
		t.Errorf("expected action 'kubectl get pods -n default', got '%s'", step.Action)
	}

	if !step.Allowed {
		t.Error("expected action to be allowed")
	}

	// Check if error handling worked properly
	if err != nil {
		t.Logf("Action execution failed as expected in test environment: %v", err)
	}
}

func TestReActSessionProcessModelResponse_NoAction(t *testing.T) {
	session := NewReActSession(3)

	response := "This is just a regular response without any action or final statement."

	err := session.ProcessModelResponse(response)
	if err == nil {
		t.Error("expected error for response without Action: or Final:")
	}

	if !strings.Contains(err.Error(), "no valid Action: or Final: found") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestReActSessionMaxSteps(t *testing.T) {
	session := NewReActSession(1) // Only allow 1 step

	// First action should work (even if it fails execution)
	session.ExecuteAction("kubectl get pods")

	// Second action should fail due to max steps
	err := session.ExecuteAction("kubectl describe pod test")
	if err == nil {
		t.Error("expected error when exceeding max steps")
	}

	if !strings.Contains(err.Error(), "maximum steps") {
		t.Errorf("expected error about max steps, got: %v", err)
	}
}

func TestReActSessionGetLastObservation(t *testing.T) {
	session := NewReActSession(3)

	// No observations initially
	obs := session.GetLastObservation()
	if obs != "" {
		t.Errorf("expected empty observation, got '%s'", obs)
	}

	// Add a blocked action
	session.ExecuteAction("kubectl delete pod test")

	obs = session.GetLastObservation()
	if !strings.Contains(obs, "ERROR:") {
		t.Errorf("expected error in observation for blocked action, got '%s'", obs)
	}
}

func TestReActSessionGetSessionSummary(t *testing.T) {
	session := NewReActSession(3)

	// Test empty session summary
	summary := session.GetSessionSummary()
	if !strings.Contains(summary, "ReAct Session Summary (0/3 steps)") {
		t.Errorf("expected session summary header, got '%s'", summary)
	}

	// Add some actions and check summary updates
	session.ExecuteAction("kubectl get pods")
	session.ExecuteAction("kubectl delete pod test") // This should be blocked

	summary = session.GetSessionSummary()
	if !strings.Contains(summary, "Step 1:") {
		t.Errorf("expected step 1 in summary, got '%s'", summary)
	}

	if !strings.Contains(summary, "Step 2:") {
		t.Errorf("expected step 2 in summary, got '%s'", summary)
	}

	if !strings.Contains(summary, "❌ BLOCKED:") {
		t.Errorf("expected blocked action indicator, got '%s'", summary)
	}
}

func TestRunShell(t *testing.T) {
	// Test with a simple command that should work
	output, err := runShell(2*time.Second, "echo 'test output'", 1024)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "test output") {
		t.Errorf("expected 'test output' in output, got '%s'", output)
	}
}

func TestDiagResultHeuristics(t *testing.T) {
	// Test that heuristic notes are added correctly
	// We can't test RunDiagPlan directly without kubectl, but we can test the heuristic logic
	// by checking how notes are added in the DiagResult processing

	// This is more of an integration test and would need a test environment with kubectl
	// For now, we verify the structure exists
	plan := PlanPodNotReady("test-pod", "default")
	if len(plan.Steps) == 0 {
		t.Error("plan should have steps")
	}
}
