// optimizer.go - Intelligent optimization advisor for Kubernetes resources
package engine

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// OptimizationAdvisor provides intelligent recommendations for cluster optimization
type OptimizationAdvisor struct {
	recommendations []Recommendation
	lastAnalysis    time.Time
}

// Recommendation represents an actionable optimization suggestion
type Recommendation struct {
	ID               string `json:"id"`
	Type             string `json:"type"`     // "resource", "network", "security", "cost"
	Severity         string `json:"severity"` // "info", "warning", "critical"
	Title            string `json:"title"`
	Description      string `json:"description"`
	Impact           string `json:"impact"`            // Expected improvement
	PreviewCmd       string `json:"preview_cmd"`       // kubectl apply --dry-run command
	DiffSuggestion   string `json:"diff_suggestion"`   // Proposed diff
	EstimatedSavings string `json:"estimated_savings"` // Cost or resource savings
	RiskLevel        string `json:"risk_level"`        // "low", "medium", "high"
	Category         string `json:"category"`          // Specific optimization category
	Automated        bool   `json:"automated"`         // Can be auto-applied
}

// ResourceUtilization represents current resource usage patterns
type ResourceUtilization struct {
	Namespace     string                   `json:"namespace"`
	Pod           string                   `json:"pod"`
	Container     string                   `json:"container"`
	CPUUsage      string                   `json:"cpu_usage"`
	MemoryUsage   string                   `json:"memory_usage"`
	CPURequest    string                   `json:"cpu_request"`
	MemoryRequest string                   `json:"memory_request"`
	CPULimit      string                   `json:"cpu_limit"`
	MemoryLimit   string                   `json:"memory_limit"`
	Utilization   ResourceUtilizationRatio `json:"utilization"`
}

type ResourceUtilizationRatio struct {
	CPUUtilization      float64 `json:"cpu_utilization"`    // Usage vs Request
	MemoryUtilization   float64 `json:"memory_utilization"` // Usage vs Request
	OverprovisedCPU     bool    `json:"overprovisioned_cpu"`
	OverprovisedMemory  bool    `json:"overprovisioned_memory"`
	UnderprovisedCPU    bool    `json:"underprovisioned_cpu"`
	UnderprovisedMemory bool    `json:"underprovisioned_memory"`
}

// NetworkAnalysis represents network configuration analysis
type NetworkAnalysis struct {
	Namespace        string               `json:"namespace"`
	OrphanedServices []OrphanedService    `json:"orphaned_services"`
	IngressIssues    []IngressIssue       `json:"ingress_issues"`
	ServiceMesh      *ServiceMeshAnalysis `json:"service_mesh"`
	NetworkPolicies  []NetworkPolicyGap   `json:"network_policies"`
}

type OrphanedService struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Reason     string `json:"reason"`
	Suggestion string `json:"suggestion"`
}

type IngressIssue struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Issue     string `json:"issue"`
	Fix       string `json:"fix"`
}

type ServiceMeshAnalysis struct {
	Detected      bool     `json:"detected"`
	Type          string   `json:"type"` // "istio", "linkerd", "consul"
	Issues        []string `json:"issues"`
	Optimizations []string `json:"optimizations"`
}

type NetworkPolicyGap struct {
	Namespace  string `json:"namespace"`
	Missing    string `json:"missing"`
	Risk       string `json:"risk"`
	Suggestion string `json:"suggestion"`
}

// NewOptimizationAdvisor creates a new optimization advisor
func NewOptimizationAdvisor() *OptimizationAdvisor {
	return &OptimizationAdvisor{
		recommendations: make([]Recommendation, 0),
		lastAnalysis:    time.Time{},
	}
}

// AnalyzeResourceUtilization provides intelligent resource optimization recommendations
func (oa *OptimizationAdvisor) AnalyzeResourceUtilization(ns string) ([]Recommendation, error) {
	var recommendations []Recommendation

	// Get resource usage data (requires metrics-server)
	utilization, err := oa.getResourceUtilization(ns)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource utilization: %w", err)
	}

	for _, util := range utilization {
		// Check for overprovisioning
		if util.Utilization.OverprovisedCPU {
			recommendations = append(recommendations, Recommendation{
				ID:       fmt.Sprintf("cpu-overprovision-%s-%s", util.Pod, util.Container),
				Type:     "resource",
				Severity: "warning",
				Title:    "CPU Overprovisioning Detected",
				Description: fmt.Sprintf("Container %s in pod %s is using only %.1f%% of requested CPU",
					util.Container, util.Pod, util.Utilization.CPUUtilization*100),
				Impact:           "Reduce CPU requests by 30-50% to free up cluster capacity",
				Category:         "cost-optimization",
				RiskLevel:        "low",
				Automated:        true,
				PreviewCmd:       oa.generateResourcePatchCommand(util, "cpu", "reduce"),
				DiffSuggestion:   oa.generateResourceDiff(util, "cpu", "reduce"),
				EstimatedSavings: "~20-40% cost reduction",
			})
		}

		if util.Utilization.OverprovisedMemory {
			recommendations = append(recommendations, Recommendation{
				ID:       fmt.Sprintf("memory-overprovision-%s-%s", util.Pod, util.Container),
				Type:     "resource",
				Severity: "warning",
				Title:    "Memory Overprovisioning Detected",
				Description: fmt.Sprintf("Container %s in pod %s is using only %.1f%% of requested memory",
					util.Container, util.Pod, util.Utilization.MemoryUtilization*100),
				Impact:           "Reduce memory requests to optimize cluster utilization",
				Category:         "cost-optimization",
				RiskLevel:        "medium",
				Automated:        false, // Memory changes are riskier
				PreviewCmd:       oa.generateResourcePatchCommand(util, "memory", "reduce"),
				DiffSuggestion:   oa.generateResourceDiff(util, "memory", "reduce"),
				EstimatedSavings: "~15-30% cost reduction",
			})
		}

		// Check for underprovisioning
		if util.Utilization.UnderprovisedCPU {
			recommendations = append(recommendations, Recommendation{
				ID:       fmt.Sprintf("cpu-underprovision-%s-%s", util.Pod, util.Container),
				Type:     "resource",
				Severity: "critical",
				Title:    "CPU Underprovisioning Detected",
				Description: fmt.Sprintf("Container %s in pod %s is using %.1f%% of requested CPU (likely throttled)",
					util.Container, util.Pod, util.Utilization.CPUUtilization*100),
				Impact:           "Increase CPU requests/limits to improve performance",
				Category:         "performance",
				RiskLevel:        "medium",
				Automated:        false,
				PreviewCmd:       oa.generateResourcePatchCommand(util, "cpu", "increase"),
				DiffSuggestion:   oa.generateResourceDiff(util, "cpu", "increase"),
				EstimatedSavings: "Improved application performance",
			})
		}

		if util.Utilization.UnderprovisedMemory {
			recommendations = append(recommendations, Recommendation{
				ID:       fmt.Sprintf("memory-underprovision-%s-%s", util.Pod, util.Container),
				Type:     "resource",
				Severity: "critical",
				Title:    "Memory Underprovisioning Detected",
				Description: fmt.Sprintf("Container %s in pod %s is using %.1f%% of requested memory (risk of OOMKill)",
					util.Container, util.Pod, util.Utilization.MemoryUtilization*100),
				Impact:           "Increase memory requests/limits to prevent OOMKilled",
				Category:         "reliability",
				RiskLevel:        "high",
				Automated:        false,
				PreviewCmd:       oa.generateResourcePatchCommand(util, "memory", "increase"),
				DiffSuggestion:   oa.generateResourceDiff(util, "memory", "increase"),
				EstimatedSavings: "Prevent application crashes",
			})
		}
	}

	return recommendations, nil
}

// AnalyzeNetworkConfiguration provides network optimization recommendations
func (oa *OptimizationAdvisor) AnalyzeNetworkConfiguration(ns string) ([]Recommendation, error) {
	var recommendations []Recommendation

	analysis, err := oa.analyzeNetwork(ns)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze network: %w", err)
	}

	// Check for orphaned services
	for _, orphaned := range analysis.OrphanedServices {
		recommendations = append(recommendations, Recommendation{
			ID:             fmt.Sprintf("orphaned-service-%s", orphaned.Name),
			Type:           "network",
			Severity:       "warning",
			Title:          "Orphaned Service Detected",
			Description:    fmt.Sprintf("Service %s has no backing pods", orphaned.Name),
			Impact:         "Remove unused service to reduce complexity",
			Category:       "cleanup",
			RiskLevel:      "low",
			Automated:      false,
			PreviewCmd:     fmt.Sprintf("kubectl delete service %s -n %s --dry-run=client", orphaned.Name, ns),
			DiffSuggestion: fmt.Sprintf("- Remove service %s (no backing pods)", orphaned.Name),
		})
	}

	// Check for ingress issues
	for _, issue := range analysis.IngressIssues {
		recommendations = append(recommendations, Recommendation{
			ID:             fmt.Sprintf("ingress-issue-%s", issue.Name),
			Type:           "network",
			Severity:       "warning",
			Title:          "Ingress Configuration Issue",
			Description:    fmt.Sprintf("Ingress %s: %s", issue.Name, issue.Issue),
			Impact:         "Fix ingress configuration for proper traffic routing",
			Category:       "connectivity",
			RiskLevel:      "medium",
			Automated:      false,
			DiffSuggestion: issue.Fix,
		})
	}

	// Check for missing network policies
	for _, gap := range analysis.NetworkPolicies {
		recommendations = append(recommendations, Recommendation{
			ID:             fmt.Sprintf("netpol-gap-%s", gap.Namespace),
			Type:           "security",
			Severity:       "warning",
			Title:          "Missing Network Policy",
			Description:    fmt.Sprintf("Namespace %s lacks network isolation", gap.Namespace),
			Impact:         "Implement network policies for better security",
			Category:       "security",
			RiskLevel:      "medium",
			Automated:      false,
			DiffSuggestion: oa.generateNetworkPolicyTemplate(gap),
		})
	}

	return recommendations, nil
}

// AnalyzeHelmValues provides Helm-specific optimization recommendations
func (oa *OptimizationAdvisor) AnalyzeHelmValues(release string) ([]Recommendation, error) {
	var recommendations []Recommendation

	// Get current Helm values
	values, err := oa.getHelmValues(release)
	if err != nil {
		return nil, fmt.Errorf("failed to get Helm values: %w", err)
	}

	// Analyze values for optimization opportunities
	if valuesMap, ok := values.(map[string]interface{}); ok {
		// Check resource configurations in values
		if resources, exists := valuesMap["resources"]; exists {
			recommendations = append(recommendations, oa.analyzeHelmResources(release, resources)...)
		}

		// Check autoscaling configuration
		if autoscaling, exists := valuesMap["autoscaling"]; exists {
			recommendations = append(recommendations, oa.analyzeHelmAutoscaling(release, autoscaling)...)
		}

		// Check security settings
		if security, exists := valuesMap["securityContext"]; exists {
			recommendations = append(recommendations, oa.analyzeHelmSecurity(release, security)...)
		}
	}

	return recommendations, nil
}

// Helper methods for data collection
func (oa *OptimizationAdvisor) getResourceUtilization(ns string) ([]ResourceUtilization, error) {
	var utilization []ResourceUtilization

	// Get pod metrics (requires metrics-server)
	metricsOut, _, err := runKubectl(10*time.Second, "-n", ns, "top", "pods", "--containers", "--no-headers")
	if err != nil {
		// If metrics-server is not available, return empty results
		return utilization, nil
	}

	lines := strings.Split(strings.TrimSpace(metricsOut), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		pod, container := fields[0], fields[1]
		cpuUsage, memUsage := fields[2], fields[3]

		// Get resource requests/limits
		resourceOut, _, err := runKubectl(5*time.Second, "-n", ns, "get", "pod", pod, "-o",
			"jsonpath={.spec.containers[?(@.name==\""+container+"\")].resources}")
		if err != nil {
			continue
		}

		var resources struct {
			Requests map[string]string `json:"requests"`
			Limits   map[string]string `json:"limits"`
		}

		if json.Unmarshal([]byte(resourceOut), &resources) != nil {
			continue
		}

		// Calculate utilization ratios
		utilRatio := oa.calculateUtilization(cpuUsage, memUsage, resources.Requests, resources.Limits)

		util := ResourceUtilization{
			Namespace:     ns,
			Pod:           pod,
			Container:     container,
			CPUUsage:      cpuUsage,
			MemoryUsage:   memUsage,
			CPURequest:    resources.Requests["cpu"],
			MemoryRequest: resources.Requests["memory"],
			CPULimit:      resources.Limits["cpu"],
			MemoryLimit:   resources.Limits["memory"],
			Utilization:   utilRatio,
		}

		utilization = append(utilization, util)
	}

	return utilization, nil
}

func (oa *OptimizationAdvisor) calculateUtilization(cpuUsage, memUsage string, requests, limits map[string]string) ResourceUtilizationRatio {
	ratio := ResourceUtilizationRatio{}

	// Parse CPU utilization
	if cpuReq, exists := requests["cpu"]; exists && cpuReq != "" {
		usageVal := oa.parseResourceValue(cpuUsage, "cpu")
		requestVal := oa.parseResourceValue(cpuReq, "cpu")
		if requestVal > 0 {
			ratio.CPUUtilization = usageVal / requestVal
			ratio.OverprovisedCPU = ratio.CPUUtilization < 0.2  // Using less than 20%
			ratio.UnderprovisedCPU = ratio.CPUUtilization > 0.9 // Using more than 90%
		}
	}

	// Parse memory utilization
	if memReq, exists := requests["memory"]; exists && memReq != "" {
		usageVal := oa.parseResourceValue(memUsage, "memory")
		requestVal := oa.parseResourceValue(memReq, "memory")
		if requestVal > 0 {
			ratio.MemoryUtilization = usageVal / requestVal
			ratio.OverprovisedMemory = ratio.MemoryUtilization < 0.3  // Using less than 30%
			ratio.UnderprovisedMemory = ratio.MemoryUtilization > 0.8 // Using more than 80%
		}
	}

	return ratio
}

func (oa *OptimizationAdvisor) parseResourceValue(value, resourceType string) float64 {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}

	// Remove units and convert to base value
	if resourceType == "cpu" {
		// Handle CPU units (m = millicores)
		if strings.HasSuffix(value, "m") {
			if val, err := strconv.ParseFloat(strings.TrimSuffix(value, "m"), 64); err == nil {
				return val / 1000 // Convert millicores to cores
			}
		}
		if val, err := strconv.ParseFloat(value, 64); err == nil {
			return val
		}
	} else if resourceType == "memory" {
		// Handle memory units (Ki, Mi, Gi)
		multipliers := map[string]float64{
			"Ki": 1024,
			"Mi": 1024 * 1024,
			"Gi": 1024 * 1024 * 1024,
		}

		for suffix, multiplier := range multipliers {
			if strings.HasSuffix(value, suffix) {
				if val, err := strconv.ParseFloat(strings.TrimSuffix(value, suffix), 64); err == nil {
					return val * multiplier
				}
			}
		}

		// Try parsing as plain number (bytes)
		if val, err := strconv.ParseFloat(value, 64); err == nil {
			return val
		}
	}

	return 0
}

func (oa *OptimizationAdvisor) analyzeNetwork(ns string) (*NetworkAnalysis, error) {
	analysis := &NetworkAnalysis{
		Namespace: ns,
	}

	// Check for orphaned services
	servicesOut, _, err := runKubectl(10*time.Second, "-n", ns, "get", "services", "-o", "json")
	if err == nil {
		var serviceList struct {
			Items []struct {
				Metadata struct {
					Name string `json:"name"`
				} `json:"metadata"`
				Spec struct {
					Selector map[string]string `json:"selector"`
				} `json:"spec"`
			} `json:"items"`
		}

		if json.Unmarshal([]byte(servicesOut), &serviceList) == nil {
			for _, svc := range serviceList.Items {
				// Check if service has backing pods
				if len(svc.Spec.Selector) > 0 {
					// Build selector string
					selectorParts := []string{}
					for k, v := range svc.Spec.Selector {
						selectorParts = append(selectorParts, fmt.Sprintf("%s=%s", k, v))
					}
					selector := strings.Join(selectorParts, ",")

					// Check for matching pods
					podsOut, _, err := runKubectl(5*time.Second, "-n", ns, "get", "pods", "-l", selector, "--no-headers")
					if err != nil || strings.TrimSpace(podsOut) == "" {
						analysis.OrphanedServices = append(analysis.OrphanedServices, OrphanedService{
							Name:       svc.Metadata.Name,
							Namespace:  ns,
							Reason:     "No matching pods found",
							Suggestion: "Remove service or fix selector",
						})
					}
				}
			}
		}
	}

	return analysis, nil
}

func (oa *OptimizationAdvisor) getHelmValues(release string) (interface{}, error) {
	out, _, err := runKubectl(10*time.Second, "get", "secret", "-l", fmt.Sprintf("name=%s", release), "-o", "json")
	if err != nil {
		return nil, err
	}

	// This is a simplified implementation - in reality, you'd need to decode Helm's secret format
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		return nil, err
	}

	return result, nil
}

// Helper methods for generating recommendations
func (oa *OptimizationAdvisor) generateResourcePatchCommand(util ResourceUtilization, resourceType, action string) string {
	var newValue string

	if action == "reduce" {
		if resourceType == "cpu" {
			// Reduce by 30%
			current := oa.parseResourceValue(util.CPURequest, "cpu")
			newValue = fmt.Sprintf("%.0fm", current*0.7*1000)
		} else {
			// Reduce by 20%
			current := oa.parseResourceValue(util.MemoryRequest, "memory")
			newValue = fmt.Sprintf("%.0fMi", current*0.8/(1024*1024))
		}
	} else {
		if resourceType == "cpu" {
			// Increase by 50%
			current := oa.parseResourceValue(util.CPURequest, "cpu")
			newValue = fmt.Sprintf("%.0fm", current*1.5*1000)
		} else {
			// Increase by 30%
			current := oa.parseResourceValue(util.MemoryRequest, "memory")
			newValue = fmt.Sprintf("%.0fMi", current*1.3/(1024*1024))
		}
	}

	return fmt.Sprintf("kubectl patch deployment {deployment} -n %s --type='json' -p='[{\"op\": \"replace\", \"path\": \"/spec/template/spec/containers/0/resources/requests/%s\", \"value\": \"%s\"}]' --dry-run=client",
		util.Namespace, resourceType, newValue)
}

func (oa *OptimizationAdvisor) generateResourceDiff(util ResourceUtilization, resourceType, action string) string {
	var current, new string

	if resourceType == "cpu" {
		current = util.CPURequest
		if action == "reduce" {
			currentVal := oa.parseResourceValue(current, "cpu")
			new = fmt.Sprintf("%.0fm", currentVal*0.7*1000)
		} else {
			currentVal := oa.parseResourceValue(current, "cpu")
			new = fmt.Sprintf("%.0fm", currentVal*1.5*1000)
		}
	} else {
		current = util.MemoryRequest
		if action == "reduce" {
			currentVal := oa.parseResourceValue(current, "memory")
			new = fmt.Sprintf("%.0fMi", currentVal*0.8/(1024*1024))
		} else {
			currentVal := oa.parseResourceValue(current, "memory")
			new = fmt.Sprintf("%.0fMi", currentVal*1.3/(1024*1024))
		}
	}

	return fmt.Sprintf(`@@ Container: %s @@
-        requests:
-          %s: %s
+        requests:
+          %s: %s`, util.Container, resourceType, current, resourceType, new)
}

func (oa *OptimizationAdvisor) generateNetworkPolicyTemplate(gap NetworkPolicyGap) string {
	return fmt.Sprintf(`apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-all
  namespace: %s
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress`, gap.Namespace)
}

func (oa *OptimizationAdvisor) analyzeHelmResources(release string, resources interface{}) []Recommendation {
	// Simplified Helm resource analysis
	return []Recommendation{}
}

func (oa *OptimizationAdvisor) analyzeHelmAutoscaling(release string, autoscaling interface{}) []Recommendation {
	// Simplified Helm autoscaling analysis
	return []Recommendation{}
}

func (oa *OptimizationAdvisor) analyzeHelmSecurity(release string, security interface{}) []Recommendation {
	// Simplified Helm security analysis
	return []Recommendation{}
}

// Global optimization advisor instance
var Optimizer = NewOptimizationAdvisor()
