// validator_improved_test.go - Enhanced tests for validator functionality
package main

import (
	"testing"
)

func TestBuildPreExecPlan_KubectlApply(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected string
	}{
		{
			name:     "kubectl apply with dry-run",
			command:  "kubectl apply -f deployment.yaml",
			expected: "medium",
		},
		{
			name:     "kubectl create with dry-run",
			command:  "kubectl create deployment nginx --image=nginx",
			expected: "medium",
		},
		{
			name:     "kubectl patch with dry-run",
			command:  "kubectl patch deployment nginx -p '{\"spec\":{\"replicas\":3}}'",
			expected: "medium",
		},
		{
			name:     "kubectl delete with confirmation",
			command:  "kubectl delete deployment nginx",
			expected: "high",
		},
		{
			name:     "dangerous delete all",
			command:  "kubectl delete all --all",
			expected: "critical",
		},
		{
			name:     "helm install with dry-run",
			command:  "helm install myapp ./chart",
			expected: "medium",
		},
		{
			name:     "helm upgrade with dry-run",
			command:  "helm upgrade myapp ./chart",
			expected: "medium",
		},
		{
			name:     "read-only kubectl get",
			command:  "kubectl get pods",
			expected: "low",
		},
		{
			name:     "read-only kubectl describe",
			command:  "kubectl describe pod nginx",
			expected: "low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := BuildPreExecPlan(tt.command)
			if plan.DangerLevel != tt.expected {
				t.Errorf("BuildPreExecPlan() danger level = %v, want %v", plan.DangerLevel, tt.expected)
			}
		})
	}
}

func TestBuildPreExecPlan_DangerousPatterns(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected string
	}{
		{
			name:     "delete all namespaces",
			command:  "kubectl delete all --all-namespaces",
			expected: "critical",
		},
		{
			name:     "force delete",
			command:  "kubectl delete pod nginx --force --grace-period=0",
			expected: "critical",
		},
		{
			name:     "wildcard selector",
			command:  "kubectl delete pods -l app=*",
			expected: "critical",
		},
		{
			name:     "delete nodes",
			command:  "kubectl delete node worker-1",
			expected: "critical",
		},
		{
			name:     "delete namespaces",
			command:  "kubectl delete namespace test",
			expected: "critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := BuildPreExecPlan(tt.command)
			if plan.DangerLevel != tt.expected {
				t.Errorf("BuildPreExecPlan() danger level = %v, want %v", plan.DangerLevel, tt.expected)
			}
		})
	}
}

func TestIsWhitelistedAction_Enhanced(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		{
			name:     "kubectl get pods",
			command:  "kubectl get pods",
			expected: true,
		},
		{
			name:     "kubectl describe pod",
			command:  "kubectl describe pod nginx",
			expected: true,
		},
		{
			name:     "kubectl logs",
			command:  "kubectl logs nginx",
			expected: true,
		},
		{
			name:     "kubectl top nodes",
			command:  "kubectl top nodes",
			expected: true,
		},
		{
			name:     "kubectl api-resources",
			command:  "kubectl api-resources",
			expected: true,
		},
		{
			name:     "kubectl version",
			command:  "kubectl version",
			expected: true,
		},
		{
			name:     "kubectl explain",
			command:  "kubectl explain pod",
			expected: true,
		},
		{
			name:     "helm lint",
			command:  "helm lint ./chart",
			expected: true,
		},
		{
			name:     "helm template",
			command:  "helm template myapp ./chart",
			expected: true,
		},
		{
			name:     "helm version",
			command:  "helm version",
			expected: true,
		},
		{
			name:     "kubectl apply (not whitelisted)",
			command:  "kubectl apply -f deployment.yaml",
			expected: false,
		},
		{
			name:     "kubectl delete (not whitelisted)",
			command:  "kubectl delete pod nginx",
			expected: false,
		},
		{
			name:     "helm install (not whitelisted)",
			command:  "helm install myapp ./chart",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsWhitelistedAction(tt.command)
			if result != tt.expected {
				t.Errorf("IsWhitelistedAction() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBuildPreExecPlan_RequireSecondConfirm(t *testing.T) {
	tests := []struct {
		name                string
		command             string
		expectedSecondConfirm bool
		expectedTypedConfirm  bool
	}{
		{
			name:                "kubectl apply requires second confirm",
			command:             "kubectl apply -f deployment.yaml",
			expectedSecondConfirm: true,
			expectedTypedConfirm:  false,
		},
		{
			name:                "kubectl delete requires second confirm",
			command:             "kubectl delete deployment nginx",
			expectedSecondConfirm: true,
			expectedTypedConfirm:  false,
		},
		{
			name:                "dangerous delete requires typed confirm",
			command:             "kubectl delete all --all",
			expectedSecondConfirm: true,
			expectedTypedConfirm:  true,
		},
		{
			name:                "read-only command no confirm",
			command:             "kubectl get pods",
			expectedSecondConfirm: false,
			expectedTypedConfirm:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := BuildPreExecPlan(tt.command)
			if plan.RequireSecondConfirm != tt.expectedSecondConfirm {
				t.Errorf("BuildPreExecPlan() RequireSecondConfirm = %v, want %v", plan.RequireSecondConfirm, tt.expectedSecondConfirm)
			}
			if plan.RequireTypedConfirm != tt.expectedTypedConfirm {
				t.Errorf("BuildPreExecPlan() RequireTypedConfirm = %v, want %v", plan.RequireTypedConfirm, tt.expectedTypedConfirm)
			}
		})
	}
}
