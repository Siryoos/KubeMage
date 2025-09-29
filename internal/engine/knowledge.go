// knowledge.go - Intelligence engine with playbooks and root-cause detection
package engine

import (
	"sort"
	"strings"
)

// PlaybookLibrary manages curated diagnostic and resolution playbooks
type PlaybookLibrary struct {
	playbooks map[string]*Playbook
	patterns  map[string]*RootCausePattern
}

// Playbook defines systematic approaches to common Kubernetes issues
type Playbook struct {
	Name        string           `json:"name"`
	Category    string           `json:"category"`
	Description string           `json:"description"`
	Triggers    []string         `json:"triggers"`   // Error patterns to match
	Steps       []PlaybookStep   `json:"steps"`      // Diagnostic commands
	Heuristics  []string         `json:"heuristics"` // What to look for
	NextMoves   []ActionableStep `json:"next_moves"` // Suggested actions
	Confidence  float64          `json:"confidence"` // Pattern match confidence
}

type PlaybookStep struct {
	Name        string `json:"name"`
	Command     string `json:"command"`
	Description string `json:"description"`
	Expected    string `json:"expected"`
	Timeout     int    `json:"timeout"` // seconds
}

type ActionableStep struct {
	Action      string `json:"action"`
	Command     string `json:"command"`
	Risk        string `json:"risk"`     // "low", "medium", "high"
	Category    string `json:"category"` // "fix", "investigate", "scale", "rollback"
	Description string `json:"description"`
}

// RootCausePattern defines intelligent pattern recognition
type RootCausePattern struct {
	Name       string           `json:"name"`
	Category   string           `json:"category"`
	Indicators []string         `json:"indicators"` // Text patterns to match
	Conditions []ConditionCheck `json:"conditions"` // Structural checks
	Solutions  []Solution       `json:"solutions"`  // Ranked solutions
	Confidence float64          `json:"confidence"` // Detection confidence
	Severity   string           `json:"severity"`   // "low", "medium", "high", "critical"
}

type ConditionCheck struct {
	Field    string      `json:"field"`    // JSON path or field name
	Operator string      `json:"operator"` // "equals", "contains", "greater_than", etc.
	Value    interface{} `json:"value"`    // Expected value
}

type Solution struct {
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	Command       string  `json:"command"`
	Risk          string  `json:"risk"`
	Effectiveness float64 `json:"effectiveness"` // 0.0-1.0
}

// RootCauseAnalysis represents the result of intelligent analysis
type RootCauseAnalysis struct {
	RootCause      string           `json:"root_cause"`
	Confidence     float64          `json:"confidence"`
	Category       string           `json:"category"`
	Severity       string           `json:"severity"`
	Indicators     []string         `json:"indicators"`
	NextSteps      []ActionableStep `json:"next_steps"`
	RelatedIssues  []string         `json:"related_issues"`
	Timeline       string           `json:"timeline"` // Expected fix time
	PreventionTips []string         `json:"prevention_tips"`
}

// NewPlaybookLibrary creates and initializes the playbook library
func NewPlaybookLibrary() *PlaybookLibrary {
	library := &PlaybookLibrary{
		playbooks: make(map[string]*Playbook),
		patterns:  make(map[string]*RootCausePattern),
	}
	library.loadBuiltinPlaybooks()
	library.loadRootCausePatterns()
	return library
}

// loadBuiltinPlaybooks loads curated playbooks for common issues
func (pl *PlaybookLibrary) loadBuiltinPlaybooks() {
	// Pod Not Ready Playbook
	pl.playbooks["pod-not-ready"] = &Playbook{
		Name:        "Pod Not Ready Investigation",
		Category:    "pod-issues",
		Description: "Systematic diagnosis of pods stuck in non-Ready states",
		Triggers:    []string{"not ready", "pending", "crashloop", "imagepull"},
		Steps: []PlaybookStep{
			{
				Name:        "Get Pod Details",
				Command:     "kubectl describe pod {pod} -n {namespace}",
				Description: "Examine pod configuration and current state",
				Expected:    "Pod details with events and conditions",
				Timeout:     10,
			},
			{
				Name:        "Check Recent Events",
				Command:     "kubectl get events -n {namespace} --field-selector involvedObject.name={pod} --sort-by=.lastTimestamp",
				Description: "Review chronological events for this pod",
				Expected:    "Time-ordered list of events",
				Timeout:     5,
			},
			{
				Name:        "Examine Container Logs",
				Command:     "kubectl logs {pod} -n {namespace} --all-containers --tail=100",
				Description: "Check application logs for errors",
				Expected:    "Recent log entries from all containers",
				Timeout:     15,
			},
		},
		Heuristics: []string{
			"Look for 'Failed to pull image' in events → ImagePullBackOff",
			"Check for 'CrashLoopBackOff' in pod status → Application startup failure",
			"Examine 'FailedScheduling' events → Resource constraints or node issues",
			"Review container exit codes → Application-specific errors",
		},
		NextMoves: []ActionableStep{
			{
				Action:      "Fix image pull issues",
				Command:     "kubectl get pods {pod} -n {namespace} -o jsonpath='{.spec.containers[0].image}'",
				Risk:        "low",
				Category:    "investigate",
				Description: "Verify image name, tag, and registry accessibility",
			},
			{
				Action:      "Check resource requests",
				Command:     "kubectl describe nodes | grep -A 10 'Allocated resources'",
				Risk:        "low",
				Category:    "investigate",
				Description: "Ensure sufficient cluster resources for scheduling",
			},
		},
	}

	// CrashLoopBackOff Playbook
	pl.playbooks["crashloop-backoff"] = &Playbook{
		Name:        "CrashLoopBackOff Resolution",
		Category:    "pod-issues",
		Description: "Diagnose and resolve container crash loops",
		Triggers:    []string{"crashloop", "restart", "exit", "crash"},
		Steps: []PlaybookStep{
			{
				Name:        "Check Container Status",
				Command:     "kubectl describe pod {pod} -n {namespace}",
				Description: "Examine container restart count and exit codes",
				Expected:    "Container status with restart reasons",
				Timeout:     10,
			},
			{
				Name:        "Get Application Logs",
				Command:     "kubectl logs {pod} -n {namespace} --previous",
				Description: "Check logs from the crashed container",
				Expected:    "Error logs from failed container",
				Timeout:     10,
			},
			{
				Name:        "Review Liveness Probes",
				Command:     "kubectl get pod {pod} -n {namespace} -o jsonpath='{.spec.containers[*].livenessProbe}'",
				Description: "Check if liveness probe is causing restarts",
				Expected:    "Probe configuration details",
				Timeout:     5,
			},
		},
		Heuristics: []string{
			"High restart count with quick failures → Configuration error",
			"Exit code 125 → Container runtime error",
			"Exit code 126 → Container not executable",
			"Exit code 127 → Command not found",
			"Exit code 1 → Application-specific error",
		},
		NextMoves: []ActionableStep{
			{
				Action:      "Disable liveness probe temporarily",
				Command:     "kubectl patch pod {pod} -n {namespace} --type='json' -p='[{\"op\": \"remove\", \"path\": \"/spec/containers/0/livenessProbe\"}]'",
				Risk:        "medium",
				Category:    "fix",
				Description: "Remove liveness probe to isolate application issues",
			},
			{
				Action:      "Check resource limits",
				Command:     "kubectl describe pod {pod} -n {namespace} | grep -A 5 'Limits'",
				Risk:        "low",
				Category:    "investigate",
				Description: "Verify memory/CPU limits aren't too restrictive",
			},
		},
	}

	// Service No Endpoints Playbook
	pl.playbooks["service-no-endpoints"] = &Playbook{
		Name:        "Service No Endpoints Resolution",
		Category:    "networking",
		Description: "Diagnose services with no backing endpoints",
		Triggers:    []string{"no endpoints", "service unavailable", "503", "connection refused"},
		Steps: []PlaybookStep{
			{
				Name:        "Check Service Configuration",
				Command:     "kubectl describe service {service} -n {namespace}",
				Description: "Examine service selector and port configuration",
				Expected:    "Service details with selector labels",
				Timeout:     5,
			},
			{
				Name:        "List Matching Pods",
				Command:     "kubectl get pods -n {namespace} --show-labels",
				Description: "Find pods that should match the service selector",
				Expected:    "Pods with their labels",
				Timeout:     10,
			},
			{
				Name:        "Check Endpoints Status",
				Command:     "kubectl get endpoints {service} -n {namespace} -o yaml",
				Description: "Examine current endpoint configuration",
				Expected:    "Endpoint addresses and ports",
				Timeout:     5,
			},
		},
		Heuristics: []string{
			"No endpoints → Selector mismatch or no ready pods",
			"Endpoints exist but connection fails → Port mismatch",
			"Pods exist but not ready → Health check failures",
		},
		NextMoves: []ActionableStep{
			{
				Action:      "Fix service selector",
				Command:     "kubectl patch service {service} -n {namespace} --type='json' -p='[{\"op\": \"replace\", \"path\": \"/spec/selector\", \"value\": {\"app\": \"correct-label\"}}]'",
				Risk:        "medium",
				Category:    "fix",
				Description: "Update service selector to match pod labels",
			},
			{
				Action:      "Check pod readiness",
				Command:     "kubectl get pods -n {namespace} -l app={app} -o wide",
				Risk:        "low",
				Category:    "investigate",
				Description: "Verify pods are ready and healthy",
			},
		},
	}

	// Storage Issues Playbook
	pl.playbooks["pvc-pending"] = &Playbook{
		Name:        "PVC Pending Resolution",
		Category:    "storage",
		Description: "Resolve PersistentVolumeClaim stuck in Pending state",
		Triggers:    []string{"pvc pending", "storage", "volume", "mount"},
		Steps: []PlaybookStep{
			{
				Name:        "Check PVC Status",
				Command:     "kubectl describe pvc {pvc} -n {namespace}",
				Description: "Examine PVC configuration and events",
				Expected:    "PVC details with binding status",
				Timeout:     10,
			},
			{
				Name:        "List Available PVs",
				Command:     "kubectl get pv",
				Description: "Check for available PersistentVolumes",
				Expected:    "List of PVs with their status",
				Timeout:     5,
			},
			{
				Name:        "Check Storage Classes",
				Command:     "kubectl get storageclass",
				Description: "Verify available storage classes",
				Expected:    "Available storage classes and provisioners",
				Timeout:     5,
			},
		},
		Heuristics: []string{
			"No matching PV → Need dynamic provisioning or manual PV creation",
			"Insufficient space → Storage class capacity issues",
			"Wrong access mode → ReadWriteOnce vs ReadWriteMany mismatch",
		},
		NextMoves: []ActionableStep{
			{
				Action:      "Create matching PV",
				Command:     "kubectl apply -f - <<EOF\napiVersion: v1\nkind: PersistentVolume\nmetadata:\n  name: {pv-name}\nspec:\n  capacity:\n    storage: {size}\n  accessModes:\n    - ReadWriteOnce\nEOF",
				Risk:        "high",
				Category:    "fix",
				Description: "Create a PersistentVolume that matches the PVC requirements",
			},
		},
	}
}

// loadRootCausePatterns loads intelligent pattern recognition rules
func (pl *PlaybookLibrary) loadRootCausePatterns() {
	// ImagePullBackOff Pattern
	pl.patterns["image-pull-backoff"] = &RootCausePattern{
		Name:     "ImagePullBackOff",
		Category: "image-issues",
		Indicators: []string{
			"imagepullbackoff", "errimagepull", "failed to pull image",
			"image not found", "unauthorized", "pull access denied",
		},
		Solutions: []Solution{
			{
				Title:         "Verify Image Name and Tag",
				Description:   "Check if the image name and tag are correct",
				Command:       "docker pull {image}",
				Risk:          "low",
				Effectiveness: 0.8,
			},
			{
				Title:         "Check Registry Authentication",
				Description:   "Ensure imagePullSecrets are configured correctly",
				Command:       "kubectl get secret {secret} -o yaml",
				Risk:          "low",
				Effectiveness: 0.9,
			},
		},
		Confidence: 0.95,
		Severity:   "medium",
	}

	// OOMKilled Pattern
	pl.patterns["oom-killed"] = &RootCausePattern{
		Name:     "OOMKilled",
		Category: "resource-limits",
		Indicators: []string{
			"oomkilled", "out of memory", "killed", "exit code 137",
			"memory limit exceeded",
		},
		Solutions: []Solution{
			{
				Title:         "Increase Memory Limits",
				Description:   "Raise memory limits for the container",
				Command:       "kubectl patch deployment {deployment} -p '{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"{container}\",\"resources\":{\"limits\":{\"memory\":\"{new-limit}\"}}}]}}}}'",
				Risk:          "medium",
				Effectiveness: 0.9,
			},
			{
				Title:         "Optimize Application Memory Usage",
				Description:   "Review application for memory leaks",
				Command:       "kubectl top pod {pod}",
				Risk:          "low",
				Effectiveness: 0.7,
			},
		},
		Confidence: 0.9,
		Severity:   "high",
	}

	// Failed Scheduling Pattern
	pl.patterns["failed-scheduling"] = &RootCausePattern{
		Name:     "FailedScheduling",
		Category: "scheduling",
		Indicators: []string{
			"failedscheduling", "insufficient", "no nodes available",
			"unschedulable", "taints", "affinity",
		},
		Solutions: []Solution{
			{
				Title:         "Check Node Resources",
				Description:   "Verify available CPU/memory on nodes",
				Command:       "kubectl describe nodes | grep -A 10 'Allocated resources'",
				Risk:          "low",
				Effectiveness: 0.8,
			},
			{
				Title:         "Review Node Affinity Rules",
				Description:   "Check if pod affinity/anti-affinity rules are too restrictive",
				Command:       "kubectl describe pod {pod} | grep -A 20 'Node-Selectors'",
				Risk:          "low",
				Effectiveness: 0.7,
			},
		},
		Confidence: 0.85,
		Severity:   "medium",
	}
}

// DetectRootCause performs intelligent analysis of observations
func (pl *PlaybookLibrary) DetectRootCause(observations []string) *RootCauseAnalysis {
	candidates := make(map[string]float64)
	indicators := make(map[string][]string)

	// Score each pattern against observations
	for patternName, pattern := range pl.patterns {
		score := 0.0
		matched := []string{}

		for _, observation := range observations {
			obsLower := strings.ToLower(observation)
			for _, indicator := range pattern.Indicators {
				if strings.Contains(obsLower, strings.ToLower(indicator)) {
					score += pattern.Confidence
					matched = append(matched, indicator)
				}
			}
		}

		if score > 0 {
			candidates[patternName] = score
			indicators[patternName] = matched
		}
	}

	// Find best match
	var bestPattern string
	var bestScore float64
	for pattern, score := range candidates {
		if score > bestScore {
			bestScore = score
			bestPattern = pattern
		}
	}

	if bestPattern == "" {
		return &RootCauseAnalysis{
			RootCause:  "Unknown",
			Confidence: 0.0,
			Category:   "unknown",
			Severity:   "unknown",
			NextSteps:  []ActionableStep{},
		}
	}

	// Build analysis from best match
	pattern := pl.patterns[bestPattern]
	analysis := &RootCauseAnalysis{
		RootCause:      pattern.Name,
		Confidence:     min(bestScore/2.0, 1.0), // Normalize confidence
		Category:       pattern.Category,
		Severity:       pattern.Severity,
		Indicators:     indicators[bestPattern],
		Timeline:       pl.estimateFixTime(pattern),
		PreventionTips: pl.getPreventionTips(pattern),
	}

	// Convert solutions to actionable steps
	for _, solution := range pattern.Solutions {
		step := ActionableStep{
			Action:      solution.Title,
			Command:     solution.Command,
			Risk:        solution.Risk,
			Category:    "fix",
			Description: solution.Description,
		}
		analysis.NextSteps = append(analysis.NextSteps, step)
	}

	// Sort next steps by effectiveness
	sort.Slice(analysis.NextSteps, func(i, j int) bool {
		// Priority: low risk first, then by solution effectiveness
		if analysis.NextSteps[i].Risk != analysis.NextSteps[j].Risk {
			riskOrder := map[string]int{"low": 0, "medium": 1, "high": 2}
			return riskOrder[analysis.NextSteps[i].Risk] < riskOrder[analysis.NextSteps[j].Risk]
		}
		return true // Maintain original order for same risk level
	})

	return analysis
}

// GetPlaybook retrieves appropriate playbook for the given issue
func (pl *PlaybookLibrary) GetPlaybook(issue string) *Playbook {
	issueLower := strings.ToLower(issue)

	// Find best matching playbook
	var bestMatch *Playbook
	var bestScore float64

	for _, playbook := range pl.playbooks {
		score := 0.0
		for _, trigger := range playbook.Triggers {
			if strings.Contains(issueLower, strings.ToLower(trigger)) {
				score += 1.0
			}
		}
		if score > bestScore {
			bestScore = score
			bestMatch = playbook
		}
	}

	return bestMatch
}

// SuggestNextMoves provides intelligent suggestions based on analysis
func (pl *PlaybookLibrary) SuggestNextMoves(analysis *RootCauseAnalysis) []ActionableStep {
	// Enhance next steps with additional context-aware suggestions
	enhanced := make([]ActionableStep, len(analysis.NextSteps))
	copy(enhanced, analysis.NextSteps)

	// Add category-specific suggestions
	switch analysis.Category {
	case "image-issues":
		enhanced = append(enhanced, ActionableStep{
			Action:      "Test image locally",
			Command:     "docker run --rm {image} echo 'Image works'",
			Risk:        "low",
			Category:    "investigate",
			Description: "Verify the image works in a local Docker environment",
		})

	case "resource-limits":
		enhanced = append(enhanced, ActionableStep{
			Action:      "Monitor resource usage",
			Command:     "kubectl top pod {pod} --containers",
			Risk:        "low",
			Category:    "investigate",
			Description: "Monitor actual resource consumption patterns",
		})

	case "networking":
		enhanced = append(enhanced, ActionableStep{
			Action:      "Test service connectivity",
			Command:     "kubectl run test-pod --image=busybox --rm -it -- wget -O- {service}:{port}",
			Risk:        "low",
			Category:    "investigate",
			Description: "Test service connectivity from within cluster",
		})
	}

	return enhanced
}

// Helper functions
func (pl *PlaybookLibrary) estimateFixTime(pattern *RootCausePattern) string {
	switch pattern.Severity {
	case "low":
		return "5-15 minutes"
	case "medium":
		return "15-30 minutes"
	case "high":
		return "30-60 minutes"
	case "critical":
		return "1-2 hours"
	default:
		return "varies"
	}
}

func (pl *PlaybookLibrary) getPreventionTips(pattern *RootCausePattern) []string {
	tips := map[string][]string{
		"image-issues": {
			"Use specific image tags instead of 'latest'",
			"Set up imagePullSecrets for private registries",
			"Test images in staging before production deployment",
		},
		"resource-limits": {
			"Set appropriate resource requests and limits",
			"Monitor application memory usage patterns",
			"Use horizontal pod autoscaling for variable workloads",
		},
		"scheduling": {
			"Ensure cluster has sufficient capacity",
			"Review node affinity rules for conflicts",
			"Use pod disruption budgets for high availability",
		},
		"networking": {
			"Verify service selectors match pod labels",
			"Test service connectivity during deployment",
			"Use readiness probes to ensure pod health",
		},
	}

	if tipList, exists := tips[pattern.Category]; exists {
		return tipList
	}
	return []string{"Follow Kubernetes best practices", "Test changes in non-production first"}
}

// Removed global instance - now created via dependency injection

// math helper function
// min function is now defined in types.go
