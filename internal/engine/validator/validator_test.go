// validator_test.go
package validator

import (
	"testing"
)

func TestBuildPreExecPlan(t *testing.T) {
	testCases := []struct {
		name                  string
		cmd                   string
		expectedFirstRun      string
		expectedSecondConfirm bool
		expectedNotes         []string
		expectedDangerous     bool
	}{
		{
			name:                  "simple get",
			cmd:                   "kubectl get pods",
			expectedFirstRun:      "kubectl get pods",
			expectedSecondConfirm: false,
		},
		{
			name:                  "dangerous delete",
			cmd:                   "kubectl delete all --all-namespaces",
			expectedFirstRun:      "kubectl delete all --all-namespaces --dry-run=client",
			expectedSecondConfirm: true,
			expectedDangerous:     true,
		},
		{
			name:                  "simple apply",
			cmd:                   "kubectl apply -f my.yaml",
			expectedFirstRun:      "kubectl apply -f my.yaml --dry-run=client",
			expectedSecondConfirm: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			plan := BuildPreExecPlan(tc.cmd)
			if plan.FirstRunCommand != tc.expectedFirstRun {
				t.Errorf("expected first run command to be %q, got %q", tc.expectedFirstRun, plan.FirstRunCommand)
			}
			if plan.RequireSecondConfirm != tc.expectedSecondConfirm {
				t.Errorf("expected second confirm to be %v, got %v", tc.expectedSecondConfirm, plan.RequireSecondConfirm)
			}
			if tc.expectedDangerous {
				found := false
				for _, note := range plan.Notes {
					if note == "🚨 CRITICAL: Bulk delete operation detected!" {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected dangerous command note, but not found. Got notes: %v", plan.Notes)
				}
			}
		})
	}
}
