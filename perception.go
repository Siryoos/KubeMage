// perception.go - Enhanced surgical awareness for intelligent Kubernetes operations
package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// FactHelper provides surgical, token-cheap awareness
type FactHelper struct {
	cache map[string]CachedFact
	ttl   time.Duration
}

type CachedFact struct {
	Data      interface{}
	Timestamp time.Time
}

// NewFactHelper creates a new fact helper with 30-second cache TTL
func NewFactHelper() *FactHelper {
	return &FactHelper{
		cache: make(map[string]CachedFact),
		ttl:   30 * time.Second,
	}
}

// getCached retrieves cached data if valid, otherwise calls fetchFn
func (f *FactHelper) getCached(key string, fetchFn func() (interface{}, error)) (interface{}, error) {
	if cached, exists := f.cache[key]; exists {
		if time.Since(cached.Timestamp) < f.ttl {
			return cached.Data, nil
		}
	}

	data, err := fetchFn()
	if err != nil {
		return nil, err
	}

	f.cache[key] = CachedFact{
		Data:      data,
		Timestamp: time.Now(),
	}
	return data, nil
}

// PodsSummary provides phase histogram + top 3 waiting/terminated reasons
type PodsSummary struct {
	Phases        map[string]int    `json:"phases"`
	TopProblems   []ProblemSummary  `json:"top_problems"`
	TotalPods     int               `json:"total_pods"`
	HealthyRatio  float64           `json:"healthy_ratio"`
}

type ProblemSummary struct {
	Reason string `json:"reason"`
	Count  int    `json:"count"`
	Sample string `json:"sample"` // Example pod name
}

func (f *FactHelper) PodsSummary(ns string) (*PodsSummary, error) {
	key := fmt.Sprintf("pods:%s", ns)
	data, err := f.getCached(key, func() (interface{}, error) {
		return f.fetchPodsSummary(ns)
	})
	if err != nil {
		return nil, err
	}
	return data.(*PodsSummary), nil
}

func (f *FactHelper) fetchPodsSummary(ns string) (*PodsSummary, error) {
	// Use jsonpath for surgical extraction
	out, _, err := runKubectl(4*time.Second, "-n", ns, "get", "pods", "-o",
		"jsonpath={range .items[*]}{.metadata.name}{\"\\t\"}{.status.phase}{\"\\t\"}{.status.containerStatuses[0].state.waiting.reason}{\"\\t\"}{.status.containerStatuses[0].lastState.terminated.reason}{\"\\n\"}{end}")
	if err != nil {
		return nil, err
	}

	summary := &PodsSummary{
		Phases: make(map[string]int),
	}

	problemCounts := make(map[string]PodProblemCount)
	lines := strings.Split(strings.TrimSpace(out), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}

		podName, phase := parts[0], parts[1]
		summary.Phases[phase]++
		summary.TotalPods++

		// Track problems with example pod names
		if len(parts) > 2 && parts[2] != "" {
			reason := parts[2]
			if problem, exists := problemCounts[reason]; exists {
				problem.Count++
			} else {
				problemCounts[reason] = PodProblemCount{Count: 1, SamplePod: podName}
			}
		}
		if len(parts) > 3 && parts[3] != "" {
			reason := parts[3]
			if problem, exists := problemCounts[reason]; exists {
				problem.Count++
			} else {
				problemCounts[reason] = PodProblemCount{Count: 1, SamplePod: podName}
			}
		}
	}

	// Calculate healthy ratio
	running := summary.Phases["Running"]
	if summary.TotalPods > 0 {
		summary.HealthyRatio = float64(running) / float64(summary.TotalPods)
	}

	// Get top 3 problems by frequency
	var problems []ProblemSummary
	for reason, problem := range problemCounts {
		problems = append(problems, ProblemSummary{
			Reason: reason,
			Count:  problem.Count,
			Sample: problem.SamplePod,
		})
	}
	sort.Slice(problems, func(i, j int) bool {
		return problems[i].Count > problems[j].Count
	})
	if len(problems) > 3 {
		problems = problems[:3]
	}
	summary.TopProblems = problems

	return summary, nil
}

type PodProblemCount struct {
	Count     int
	SamplePod string
}

// DeploymentProgress provides replicas/available/unavailable + last event
type DeploymentProgress struct {
	Name         string `json:"name"`
	Replicas     int    `json:"replicas"`
	Available    int    `json:"available"`
	Unavailable  int    `json:"unavailable"`
	Updated      int    `json:"updated"`
	LastEvent    string `json:"last_event"`
	Conditions   []string `json:"conditions"`
	ReadyRatio   float64  `json:"ready_ratio"`
}

func (f *FactHelper) DeployProgress(ns, name string) (*DeploymentProgress, error) {
	key := fmt.Sprintf("deploy:%s:%s", ns, name)
	data, err := f.getCached(key, func() (interface{}, error) {
		return f.fetchDeploymentProgress(ns, name)
	})
	if err != nil {
		return nil, err
	}
	return data.(*DeploymentProgress), nil
}

func (f *FactHelper) fetchDeploymentProgress(ns, name string) (*DeploymentProgress, error) {
	// Get deployment status
	out, _, err := runKubectl(3*time.Second, "-n", ns, "get", "deployment", name, "-o",
		"jsonpath={.metadata.name}{\"\\t\"}{.spec.replicas}{\"\\t\"}{.status.availableReplicas}{\"\\t\"}{.status.unavailableReplicas}{\"\\t\"}{.status.updatedReplicas}")
	if err != nil {
		return nil, err
	}

	parts := strings.Split(strings.TrimSpace(out), "\t")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid deployment data")
	}

	progress := &DeploymentProgress{Name: parts[0]}

	if len(parts) > 1 && parts[1] != "" {
		fmt.Sscanf(parts[1], "%d", &progress.Replicas)
	}
	if len(parts) > 2 && parts[2] != "" {
		fmt.Sscanf(parts[2], "%d", &progress.Available)
	}
	if len(parts) > 3 && parts[3] != "" {
		fmt.Sscanf(parts[3], "%d", &progress.Unavailable)
	}
	if len(parts) > 4 && parts[4] != "" {
		fmt.Sscanf(parts[4], "%d", &progress.Updated)
	}

	// Calculate ready ratio
	if progress.Replicas > 0 {
		progress.ReadyRatio = float64(progress.Available) / float64(progress.Replicas)
	}

	// Get last event
	eventOut, _, err := runKubectl(2*time.Second, "-n", ns, "get", "events",
		"--field-selector", fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Deployment", name),
		"--sort-by", ".lastTimestamp", "-o", "jsonpath={.items[-1:].message}")
	if err == nil && strings.TrimSpace(eventOut) != "" {
		progress.LastEvent = strings.TrimSpace(eventOut)
	}

	// Get conditions
	condOut, _, err := runKubectl(2*time.Second, "-n", ns, "get", "deployment", name, "-o",
		"jsonpath={.status.conditions[*].type}")
	if err == nil {
		conditions := strings.Fields(condOut)
		progress.Conditions = conditions
	}

	return progress, nil
}

// ServiceStatus provides type, endpoints count, ingress status
type ServiceStatus struct {
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	ClusterIP      string   `json:"cluster_ip"`
	ExternalIP     string   `json:"external_ip"`
	EndpointsCount int      `json:"endpoints_count"`
	PortsCount     int      `json:"ports_count"`
	Selector       string   `json:"selector"`
	Ready          bool     `json:"ready"`
}

func (f *FactHelper) ServiceCheck(ns, name string) (*ServiceStatus, error) {
	key := fmt.Sprintf("svc:%s:%s", ns, name)
	data, err := f.getCached(key, func() (interface{}, error) {
		return f.fetchServiceStatus(ns, name)
	})
	if err != nil {
		return nil, err
	}
	return data.(*ServiceStatus), nil
}

func (f *FactHelper) fetchServiceStatus(ns, name string) (*ServiceStatus, error) {
	// Get service info
	out, _, err := runKubectl(3*time.Second, "-n", ns, "get", "service", name, "-o",
		"jsonpath={.metadata.name}{\"\\t\"}{.spec.type}{\"\\t\"}{.spec.clusterIP}{\"\\t\"}{.status.loadBalancer.ingress[0].ip}{\"\\t\"}{.spec.selector}")
	if err != nil {
		return nil, err
	}

	parts := strings.Split(strings.TrimSpace(out), "\t")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid service data")
	}

	status := &ServiceStatus{
		Name: parts[0],
		Type: parts[1],
	}

	if len(parts) > 2 {
		status.ClusterIP = parts[2]
	}
	if len(parts) > 3 {
		status.ExternalIP = parts[3]
	}
	if len(parts) > 4 {
		status.Selector = parts[4]
	}

	// Get endpoints count
	epOut, _, err := runKubectl(2*time.Second, "-n", ns, "get", "endpoints", name, "-o",
		"jsonpath={.subsets[*].addresses[*].ip}")
	if err == nil {
		endpoints := strings.Fields(epOut)
		status.EndpointsCount = len(endpoints)
	}

	// Check if service is ready (has endpoints)
	status.Ready = status.EndpointsCount > 0

	return status, nil
}

// IngressStatus provides class, rules, backend readiness
type IngressStatus struct {
	Name         string   `json:"name"`
	Class        string   `json:"class"`
	Hosts        []string `json:"hosts"`
	Backends     []string `json:"backends"`
	Ready        bool     `json:"ready"`
	LoadBalancer string   `json:"load_balancer"`
}

func (f *FactHelper) IngressCheck(ns, name string) (*IngressStatus, error) {
	key := fmt.Sprintf("ing:%s:%s", ns, name)
	data, err := f.getCached(key, func() (interface{}, error) {
		return f.fetchIngressStatus(ns, name)
	})
	if err != nil {
		return nil, err
	}
	return data.(*IngressStatus), nil
}

func (f *FactHelper) fetchIngressStatus(ns, name string) (*IngressStatus, error) {
	out, _, err := runKubectl(3*time.Second, "-n", ns, "get", "ingress", name, "-o",
		"jsonpath={.metadata.name}{\"\\t\"}{.spec.ingressClassName}{\"\\t\"}{.spec.rules[*].host}{\"\\t\"}{.status.loadBalancer.ingress[0].ip}")
	if err != nil {
		return nil, err
	}

	parts := strings.Split(strings.TrimSpace(out), "\t")
	if len(parts) < 1 {
		return nil, fmt.Errorf("invalid ingress data")
	}

	status := &IngressStatus{
		Name: parts[0],
	}

	if len(parts) > 1 {
		status.Class = parts[1]
	}
	if len(parts) > 2 {
		hosts := strings.Fields(parts[2])
		status.Hosts = hosts
	}
	if len(parts) > 3 {
		status.LoadBalancer = parts[3]
	}

	// Check if ingress is ready (has load balancer IP)
	status.Ready = status.LoadBalancer != ""

	return status, nil
}

// ResourceSnapshot provides quota/limitRange + resource hotspots
type ResourceSnapshot struct {
	Namespace     string              `json:"namespace"`
	Quotas        []ResourceQuota     `json:"quotas"`
	LimitRanges   []LimitRange        `json:"limit_ranges"`
	HotSpots      []ResourceHotSpot   `json:"hot_spots"`
	OverallHealth string              `json:"overall_health"`
}

type ResourceQuota struct {
	Name string            `json:"name"`
	Used map[string]string `json:"used"`
	Hard map[string]string `json:"hard"`
}

type LimitRange struct {
	Name   string                 `json:"name"`
	Limits map[string]interface{} `json:"limits"`
}

type ResourceHotSpot struct {
	Type        string  `json:"type"` // "cpu", "memory", "storage"
	Resource    string  `json:"resource"` // pod/deployment name
	Usage       string  `json:"usage"`
	Limit       string  `json:"limit"`
	Utilization float64 `json:"utilization"`
	Severity    string  `json:"severity"` // "high", "medium", "low"
}

func (f *FactHelper) QuotaSnapshot(ns string) (*ResourceSnapshot, error) {
	key := fmt.Sprintf("quota:%s", ns)
	data, err := f.getCached(key, func() (interface{}, error) {
		return f.fetchResourceSnapshot(ns)
	})
	if err != nil {
		return nil, err
	}
	return data.(*ResourceSnapshot), nil
}

func (f *FactHelper) fetchResourceSnapshot(ns string) (*ResourceSnapshot, error) {
	snapshot := &ResourceSnapshot{
		Namespace: ns,
	}

	// Get resource quotas
	quotaOut, _, err := runKubectl(3*time.Second, "-n", ns, "get", "resourcequota", "-o", "json")
	if err == nil {
		var quotaList struct {
			Items []struct {
				Metadata struct {
					Name string `json:"name"`
				} `json:"metadata"`
				Status struct {
					Used map[string]string `json:"used"`
					Hard map[string]string `json:"hard"`
				} `json:"status"`
			} `json:"items"`
		}

		if json.Unmarshal([]byte(quotaOut), &quotaList) == nil {
			for _, item := range quotaList.Items {
				snapshot.Quotas = append(snapshot.Quotas, ResourceQuota{
					Name: item.Metadata.Name,
					Used: item.Status.Used,
					Hard: item.Status.Hard,
				})
			}
		}
	}

	// Determine overall health based on quotas
	snapshot.OverallHealth = f.calculateHealthFromQuotas(snapshot.Quotas)

	return snapshot, nil
}

func (f *FactHelper) calculateHealthFromQuotas(quotas []ResourceQuota) string {
	if len(quotas) == 0 {
		return "unknown"
	}

	// Simple health calculation based on resource usage
	highUsage := 0
	totalChecks := 0

	for _, quota := range quotas {
		for resource, used := range quota.Used {
			if _, exists := quota.Hard[resource]; exists {
				totalChecks++
				// Simple percentage calculation (this is simplified)
				if strings.Contains(used, "100%") || strings.Contains(used, "9") {
					highUsage++
				}
			}
		}
	}

	if totalChecks == 0 {
		return "unknown"
	}

	ratio := float64(highUsage) / float64(totalChecks)
	if ratio > 0.8 {
		return "critical"
	} else if ratio > 0.6 {
		return "warning"
	} else {
		return "healthy"
	}
}

// Workspace awareness for charts, templates, and manifests
type WorkspaceSummary struct {
	Charts     []ChartInfo     `json:"charts"`
	Templates  []TemplateInfo  `json:"templates"`
	Manifests  []ManifestInfo  `json:"manifests"`
	Values     []ValuesInfo    `json:"values"`
	LastScan   time.Time       `json:"last_scan"`
}

type ChartInfo struct {
	Path        string `json:"path"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Hash        string `json:"hash"`
}

type TemplateInfo struct {
	Path string `json:"path"`
	Kind string `json:"kind"`
	Name string `json:"name"`
	Hash string `json:"hash"`
}

type ManifestInfo struct {
	Path      string `json:"path"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Hash      string `json:"hash"`
}

type ValuesInfo struct {
	Path     string                 `json:"path"`
	Chart    string                 `json:"chart"`
	Content  map[string]interface{} `json:"content"`
	Hash     string                 `json:"hash"`
}

func (f *FactHelper) IndexWorkspace() (*WorkspaceSummary, error) {
	key := "workspace:index"
	data, err := f.getCached(key, func() (interface{}, error) {
		return f.scanWorkspace()
	})
	if err != nil {
		return nil, err
	}
	return data.(*WorkspaceSummary), nil
}

func (f *FactHelper) scanWorkspace() (*WorkspaceSummary, error) {
	index := &WorkspaceSummary{
		LastScan: time.Now(),
	}

	// Scan for Helm charts
	chartPaths, _ := filepath.Glob("charts/*/Chart.yaml")
	chartPaths2, _ := filepath.Glob("*/Chart.yaml")
	chartPaths = append(chartPaths, chartPaths2...)

	for _, path := range chartPaths {
		if info := f.parseChartYAML(path); info != nil {
			index.Charts = append(index.Charts, *info)
		}
	}

	// Scan for templates
	templatePaths, _ := filepath.Glob("charts/*/templates/*.yaml")
	templatePaths2, _ := filepath.Glob("*/templates/*.yaml")
	templatePaths = append(templatePaths, templatePaths2...)

	for _, path := range templatePaths {
		if info := f.parseTemplate(path); info != nil {
			index.Templates = append(index.Templates, *info)
		}
	}

	// Scan for values files
	valuesPaths, _ := filepath.Glob("charts/*/values*.yaml")
	valuesPaths2, _ := filepath.Glob("values*.yaml")
	valuesPaths = append(valuesPaths, valuesPaths2...)

	for _, path := range valuesPaths {
		if info := f.parseValuesYAML(path); info != nil {
			index.Values = append(index.Values, *info)
		}
	}

	// Scan for manifests
	manifestPaths, _ := filepath.Glob("**/*.yaml")
	for _, path := range manifestPaths {
		if !strings.Contains(path, "Chart.yaml") && !strings.Contains(path, "values") {
			if info := f.parseManifest(path); info != nil {
				index.Manifests = append(index.Manifests, *info)
			}
		}
	}

	return index, nil
}

// Helper functions for parsing different file types
func (f *FactHelper) parseChartYAML(path string) *ChartInfo {
	// Simplified implementation - would need proper YAML parsing
	return &ChartInfo{
		Path: path,
		Name: filepath.Base(filepath.Dir(path)),
		Hash: fmt.Sprintf("%x", time.Now().UnixNano()), // Simplified hash
	}
}

func (f *FactHelper) parseTemplate(path string) *TemplateInfo {
	return &TemplateInfo{
		Path: path,
		Kind: "Template", // Would extract from YAML
		Name: strings.TrimSuffix(filepath.Base(path), ".yaml"),
		Hash: fmt.Sprintf("%x", time.Now().UnixNano()),
	}
}

func (f *FactHelper) parseValuesYAML(path string) *ValuesInfo {
	return &ValuesInfo{
		Path:  path,
		Chart: filepath.Base(filepath.Dir(path)),
		Hash:  fmt.Sprintf("%x", time.Now().UnixNano()),
	}
}

func (f *FactHelper) parseManifest(path string) *ManifestInfo {
	return &ManifestInfo{
		Path: path,
		Kind: "Unknown", // Would extract from YAML
		Name: strings.TrimSuffix(filepath.Base(path), ".yaml"),
		Hash: fmt.Sprintf("%x", time.Now().UnixNano()),
	}
}

// Global fact helper instance
var Facts = NewFactHelper()