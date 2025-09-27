package main

import (
	"strings"
	"testing"
)

func TestBuildPreExecPlan_ReadOnlyCommands(t *testing.T) {
	tests := []struct {
		name     string
		cmd      string
		expected struct {
			dangerLevel          string
			requireSecondConfirm bool
			requireTypedConfirm  bool
		}
	}{
		{
			name: "kubectl get pods",
			cmd:  "kubectl get pods",
			expected: struct {
				dangerLevel          string
				requireSecondConfirm bool
				requireTypedConfirm  bool
			}{
				dangerLevel:          "low",
				requireSecondConfirm: false,
				requireTypedConfirm:  false,
			},
		},
		{
			name: "kubectl describe deployment",
			cmd:  "kubectl describe deployment nginx",
			expected: struct {
				dangerLevel          string
				requireSecondConfirm bool
				requireTypedConfirm  bool
			}{
				dangerLevel:          "low",
				requireSecondConfirm: false,
				requireTypedConfirm:  false,
			},
		},
		{
			name: "kubectl logs",
			cmd:  "kubectl logs pod-name",
			expected: struct {
				dangerLevel          string
				requireSecondConfirm bool
				requireTypedConfirm  bool
			}{
				dangerLevel:          "low",
				requireSecondConfirm: false,
				requireTypedConfirm:  false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := BuildPreExecPlan(tt.cmd)

			if plan.DangerLevel != tt.expected.dangerLevel {
				t.Errorf("Expected danger level %s, got %s", tt.expected.dangerLevel, plan.DangerLevel)
			}

			if plan.RequireSecondConfirm != tt.expected.requireSecondConfirm {
				t.Errorf("Expected RequireSecondConfirm %v, got %v", tt.expected.requireSecondConfirm, plan.RequireSecondConfirm)
			}

			if plan.RequireTypedConfirm != tt.expected.requireTypedConfirm {
				t.Errorf("Expected RequireTypedConfirm %v, got %v", tt.expected.requireTypedConfirm, plan.RequireTypedConfirm)
			}

			if plan.FirstRunCommand != tt.cmd {
				t.Errorf("Expected FirstRunCommand to be %s, got %s", tt.cmd, plan.FirstRunCommand)
			}
		})
	}
}

func TestBuildPreExecPlan_MutatingCommands(t *testing.T) {
	tests := []struct {
		name     string
		cmd      string
		expected struct {
			dangerLevel          string
			requireSecondConfirm bool
			requireTypedConfirm  bool
			shouldHaveDryRun     bool
		}
	}{
		{
			name: "kubectl apply",
			cmd:  "kubectl apply -f deployment.yaml",
			expected: struct {
				dangerLevel          string
				requireSecondConfirm bool
				requireTypedConfirm  bool
				shouldHaveDryRun     bool
			}{
				dangerLevel:          "medium",
				requireSecondConfirm: true,
				requireTypedConfirm:  false,
				shouldHaveDryRun:     true,
			},
		},
		{
			name: "kubectl delete pod",
			cmd:  "kubectl delete pod my-pod",
			expected: struct {
				dangerLevel          string
				requireSecondConfirm bool
				requireTypedConfirm  bool
				shouldHaveDryRun     bool
			}{
				dangerLevel:          "medium",
				requireSecondConfirm: true,
				requireTypedConfirm:  false,
				shouldHaveDryRun:     true,
			},
		},
		{
			name: "kubectl create",
			cmd:  "kubectl create deployment nginx --image=nginx",
			expected: struct {
				dangerLevel          string
				requireSecondConfirm bool
				requireTypedConfirm  bool
				shouldHaveDryRun     bool
			}{
				dangerLevel:          "medium",
				requireSecondConfirm: true,
				requireTypedConfirm:  false,
				shouldHaveDryRun:     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := BuildPreExecPlan(tt.cmd)

			if plan.DangerLevel != tt.expected.dangerLevel {
				t.Errorf("Expected danger level %s, got %s", tt.expected.dangerLevel, plan.DangerLevel)
			}

			if plan.RequireSecondConfirm != tt.expected.requireSecondConfirm {
				t.Errorf("Expected RequireSecondConfirm %v, got %v", tt.expected.requireSecondConfirm, plan.RequireSecondConfirm)
			}

			if plan.RequireTypedConfirm != tt.expected.requireTypedConfirm {
				t.Errorf("Expected RequireTypedConfirm %v, got %v", tt.expected.requireTypedConfirm, plan.RequireTypedConfirm)
			}

			if tt.expected.shouldHaveDryRun && !strings.Contains(plan.FirstRunCommand, "--dry-run") {
				t.Errorf("Expected FirstRunCommand to contain --dry-run, got %s", plan.FirstRunCommand)
			}
		})
	}
}

func TestBuildPreExecPlan_DangerousCommands(t *testing.T) {
	tests := []struct {
		name     string
		cmd      string
		expected struct {
			dangerLevel          string
			requireSecondConfirm bool
			requireTypedConfirm  bool
		}
	}{
		{
			name: "delete all resources",
			cmd:  "kubectl delete all --all",
			expected: struct {
				dangerLevel          string
				requireSecondConfirm bool
				requireTypedConfirm  bool
			}{
				dangerLevel:          "critical",
				requireSecondConfirm: true,
				requireTypedConfirm:  true,
			},
		},
		{
			name: "delete with wildcard selector",
			cmd:  "kubectl delete pods -l '*'",
			expected: struct {
				dangerLevel          string
				requireSecondConfirm bool
				requireTypedConfirm  bool
			}{
				dangerLevel:          "critical",
				requireSecondConfirm: true,
				requireTypedConfirm:  true,
			},
		},
		{
			name: "force delete",
			cmd:  "kubectl delete pod my-pod --force",
			expected: struct {
				dangerLevel          string
				requireSecondConfirm bool
				requireTypedConfirm  bool
			}{
				dangerLevel:          "high",
				requireSecondConfirm: true,
				requireTypedConfirm:  true,
			},
		},
		{
			name: "cross-namespace delete",
			cmd:  "kubectl delete pods --all-namespaces",
			expected: struct {
				dangerLevel          string
				requireSecondConfirm bool
				requireTypedConfirm  bool
			}{
				dangerLevel:          "critical",
				requireSecondConfirm: true,
				requireTypedConfirm:  true,
			},
		},
		{
			name: "delete cluster-level resource",
			cmd:  "kubectl delete node worker-1",
			expected: struct {
				dangerLevel          string
				requireSecondConfirm bool
				requireTypedConfirm  bool
			}{
				dangerLevel:          "critical",
				requireSecondConfirm: true,
				requireTypedConfirm:  true,
			},
		},
		{
			name: "delete namespace",
			cmd:  "kubectl delete namespace production",
			expected: struct {
				dangerLevel          string
				requireSecondConfirm bool
				requireTypedConfirm  bool
			}{
				dangerLevel:          "critical",
				requireSecondConfirm: true,
				requireTypedConfirm:  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := BuildPreExecPlan(tt.cmd)

			if plan.DangerLevel != tt.expected.dangerLevel {
				t.Errorf("Expected danger level %s, got %s", tt.expected.dangerLevel, plan.DangerLevel)
			}

			if plan.RequireSecondConfirm != tt.expected.requireSecondConfirm {
				t.Errorf("Expected RequireSecondConfirm %v, got %v", tt.expected.requireSecondConfirm, plan.RequireSecondConfirm)
			}

			if plan.RequireTypedConfirm != tt.expected.requireTypedConfirm {
				t.Errorf("Expected RequireTypedConfirm %v, got %v", tt.expected.requireTypedConfirm, plan.RequireTypedConfirm)
			}

			// Dangerous commands should have safety checks
			if len(plan.SafetyChecks) == 0 {
				t.Error("Expected safety checks for dangerous command")
			}

			// Should have warning notes
			if len(plan.Notes) == 0 {
				t.Error("Expected warning notes for dangerous command")
			}
		})
	}
}

func TestBuildPreExecPlan_HelmCommands(t *testing.T) {
	tests := []struct {
		name     string
		cmd      string
		expected struct {
			dangerLevel          string
			requireSecondConfirm bool
			requireTypedConfirm  bool
			shouldHaveChecks     bool
		}
	}{
		{
			name: "helm install",
			cmd:  "helm install myapp ./chart",
			expected: struct {
				dangerLevel          string
				requireSecondConfirm bool
				requireTypedConfirm  bool
				shouldHaveChecks     bool
			}{
				dangerLevel:          "medium",
				requireSecondConfirm: true,
				requireTypedConfirm:  false,
				shouldHaveChecks:     true,
			},
		},
		{
			name: "helm upgrade",
			cmd:  "helm upgrade myapp ./chart",
			expected: struct {
				dangerLevel          string
				requireSecondConfirm bool
				requireTypedConfirm  bool
				shouldHaveChecks     bool
			}{
				dangerLevel:          "medium",
				requireSecondConfirm: true,
				requireTypedConfirm:  false,
				shouldHaveChecks:     true,
			},
		},
		{
			name: "helm install production",
			cmd:  "helm install myapp ./chart --namespace production",
			expected: struct {
				dangerLevel          string
				requireSecondConfirm bool
				requireTypedConfirm  bool
				shouldHaveChecks     bool
			}{
				dangerLevel:          "high",
				requireSecondConfirm: true,
				requireTypedConfirm:  true,
				shouldHaveChecks:     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := BuildPreExecPlan(tt.cmd)

			if plan.DangerLevel != tt.expected.dangerLevel {
				t.Errorf("Expected danger level %s, got %s", tt.expected.dangerLevel, plan.DangerLevel)
			}

			if plan.RequireSecondConfirm != tt.expected.requireSecondConfirm {
				t.Errorf("Expected RequireSecondConfirm %v, got %v", tt.expected.requireSecondConfirm, plan.RequireSecondConfirm)
			}

			if plan.RequireTypedConfirm != tt.expected.requireTypedConfirm {
				t.Errorf("Expected RequireTypedConfirm %v, got %v", tt.expected.requireTypedConfirm, plan.RequireTypedConfirm)
			}

			if tt.expected.shouldHaveChecks && len(plan.Checks) == 0 {
				t.Error("Expected validation checks for helm command")
			}

			if !strings.Contains(plan.FirstRunCommand, "--dry-run") {
				t.Errorf("Expected FirstRunCommand to contain --dry-run, got %s", plan.FirstRunCommand)
			}
		})
	}
}

func TestBuildPreExecPlan_UnknownCommands(t *testing.T) {
	tests := []struct {
		name     string
		cmd      string
		expected struct {
			dangerLevel          string
			requireSecondConfirm bool
			requireTypedConfirm  bool
		}
	}{
		{
			name: "safe unknown command",
			cmd:  "echo hello",
			expected: struct {
				dangerLevel          string
				requireSecondConfirm bool
				requireTypedConfirm  bool
			}{
				dangerLevel:          "medium",
				requireSecondConfirm: true,
				requireTypedConfirm:  false,
			},
		},
		{
			name: "dangerous unknown command",
			cmd:  "rm -rf /important/data",
			expected: struct {
				dangerLevel          string
				requireSecondConfirm bool
				requireTypedConfirm  bool
			}{
				dangerLevel:          "high",
				requireSecondConfirm: true,
				requireTypedConfirm:  true,
			},
		},
		{
			name: "database drop command",
			cmd:  "psql -c 'DROP DATABASE production'",
			expected: struct {
				dangerLevel          string
				requireSecondConfirm bool
				requireTypedConfirm  bool
			}{
				dangerLevel:          "high",
				requireSecondConfirm: true,
				requireTypedConfirm:  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := BuildPreExecPlan(tt.cmd)

			if plan.DangerLevel != tt.expected.dangerLevel {
				t.Errorf("Expected danger level %s, got %s", tt.expected.dangerLevel, plan.DangerLevel)
			}

			if plan.RequireSecondConfirm != tt.expected.requireSecondConfirm {
				t.Errorf("Expected RequireSecondConfirm %v, got %v", tt.expected.requireSecondConfirm, plan.RequireSecondConfirm)
			}

			if plan.RequireTypedConfirm != tt.expected.requireTypedConfirm {
				t.Errorf("Expected RequireTypedConfirm %v, got %v", tt.expected.requireTypedConfirm, plan.RequireTypedConfirm)
			}
		})
	}
}

func TestPreExecPlan_RequiresTypedConfirmation(t *testing.T) {
	tests := []struct {
		name     string
		plan     PreExecPlan
		expected bool
	}{
		{
			name: "dangerous command with DANGEROUS note",
			plan: PreExecPlan{
				RequireTypedConfirm: true,
				Notes:               []string{"DANGEROUS COMMAND!"},
			},
			expected: true,
		},
		{
			name: "safe command",
			plan: PreExecPlan{
				RequireTypedConfirm: false,
				Notes:               []string{"Safe operation"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.plan.RequiresTypedConfirmation()
			if result != tt.expected {
				t.Errorf("Expected RequiresTypedConfirmation %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPreExecPlan_GetDangerLevelEmoji(t *testing.T) {
	tests := []struct {
		dangerLevel string
		expected    string
	}{
		{"low", "✅"},
		{"medium", "🔶"},
		{"high", "⚠️"},
		{"critical", "🚨"},
		{"unknown", "✅"}, // default case
	}

	for _, tt := range tests {
		t.Run(tt.dangerLevel, func(t *testing.T) {
			plan := PreExecPlan{DangerLevel: tt.dangerLevel}
			result := plan.GetDangerLevelEmoji()
			if result != tt.expected {
				t.Errorf("Expected emoji %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestPreExecPlan_GetSafetyReport(t *testing.T) {
	plan := PreExecPlan{
		DangerLevel: "high",
		SafetyChecks: []string{
			"⚠️ This affects production environment",
			"⚠️ Consider backup before proceeding",
		},
		Checks: []PreviewCheck{
			{Name: "helm lint", Cmd: "helm lint ."},
			{Name: "kubectl validation", Cmd: "kubectl apply --dry-run=client -f -"},
		},
		RequireTypedConfirm: true,
	}

	report := plan.GetSafetyReport()

	// Check that report contains expected elements
	if !strings.Contains(report, "⚠️ Danger Level: HIGH") {
		t.Error("Report should contain danger level")
	}

	if !strings.Contains(report, "🔍 Safety Analysis:") {
		t.Error("Report should contain safety analysis section")
	}

	if !strings.Contains(report, "🧪 Pre-execution Validation:") {
		t.Error("Report should contain validation section")
	}

	if !strings.Contains(report, "⚠️ This command requires typing 'yes' to confirm execution.") {
		t.Error("Report should contain typed confirmation requirement")
	}

	for _, check := range plan.SafetyChecks {
		if !strings.Contains(report, check) {
			t.Errorf("Report should contain safety check: %s", check)
		}
	}
}

func TestExtractHelmChartPath(t *testing.T) {
	tests := []struct {
		name     string
		cmd      string
		expected string
	}{
		{
			name:     "current directory",
			cmd:      "helm install myapp .",
			expected: ".",
		},
		{
			name:     "relative path",
			cmd:      "helm install myapp ./charts/webapp",
			expected: "./charts/webapp",
		},
		{
			name:     "absolute path",
			cmd:      "helm install myapp /home/user/charts/webapp",
			expected: "/home/user/charts/webapp",
		},
		{
			name:     "with flags",
			cmd:      "helm install myapp ./chart --namespace production --set image.tag=v1.0.0",
			expected: "./chart",
		},
		{
			name:     "no explicit path",
			cmd:      "helm install myapp",
			expected: ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractHelmChartPath(tt.cmd)
			if result != tt.expected {
				t.Errorf("Expected chart path %s, got %s", tt.expected, result)
			}
		})
	}
}