package main

import "fmt"

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