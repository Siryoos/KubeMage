// validator.go
package main

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	reKubectlApplyCreatePatch = regexp.MustCompile(`^\s*kubectl\s+(apply|create|patch)\b`)
	reKubectlDelete           = regexp.MustCompile(`^\s*kubectl\s+delete\b`)
	reHelmInstallUpgrade      = regexp.MustCompile(`^\s*helm\s+(install|upgrade)\b`)
	reReadOnlyKubectl         = regexp.MustCompile(`^\s*kubectl\s+(get|describe|logs|top|kustomize|api-resources|version|explain)\b`)

	// Enhanced danger detection patterns
	reDangerousDelete        = regexp.MustCompile(`delete\s+(all|--all-namespaces|--all)`)
	reWildcardSelector       = regexp.MustCompile(`-l\s*["']?\*["']?|--selector\s*["']?\*["']?`)
	reForceDelete            = regexp.MustCompile(`--force|--grace-period=0`)
	reNamespaceWide          = regexp.MustCompile(`--all-namespaces|-A\b`)
	reDangerousResourceTypes = regexp.MustCompile(`\b(node|namespace|persistentvolume|storageclass|clusterrole|clusterrolebinding)\b`)
	reRecursiveDelete        = regexp.MustCompile(`--cascade\s*=\s*orphan|--cascade\s*=\s*background|--cascade\s*=\s*foreground`)
	reRootContext            = regexp.MustCompile(`--context\s*=\s*(admin|root|cluster-admin)`)
	reDeleteWithoutNamespace = regexp.MustCompile(`^\s*kubectl\s+delete\s+[^-]`)
	reWildcardResources      = regexp.MustCompile(`\s+\*\s*$|\s+\*\.|\s+all\s*$`)
)

// PreviewCheck is a command we run BEFORE any mutating action.
type PreviewCheck struct {
	Name string
	Cmd  string
}

type PreExecPlan struct {
	Original             string
	Checks               []PreviewCheck // e.g., helm lint/template, extra sanity
	FirstRunCommand      string         // typically a --dry-run variant
	RequireSecondConfirm bool           // after dry-run/lint succeeded, ask again to apply for real
	RequireTypedConfirm  bool           // for very dangerous commands, require typing "yes"
	DangerLevel          string         // "low", "medium", "high", "critical"
	Notes                []string
	SafetyChecks         []string // Additional safety validation messages
}

// BuildPreExecPlan decides the safe preview flow for a command.
func BuildPreExecPlan(cmd string) PreExecPlan {
	c := strings.TrimSpace(cmd)
	plan := PreExecPlan{
		Original:    c,
		DangerLevel: "low",
	}

	// Read-only kubectl → execute directly.
	if reReadOnlyKubectl.MatchString(c) {
		plan.FirstRunCommand = c
		plan.SafetyChecks = append(plan.SafetyChecks, "✅ Read-only operation - safe to execute")
		return plan
	}

	// Enhanced danger detection with severity levels
	plan = analyzeDangerLevel(c, plan)

	// Specific safety checks
	if reDangerousDelete.MatchString(c) {
		plan.Notes = append(plan.Notes, "🚨 CRITICAL: Bulk delete operation detected!")
		plan.RequireSecondConfirm = true
		plan.RequireTypedConfirm = true
		plan.DangerLevel = "critical"
		plan.SafetyChecks = append(plan.SafetyChecks, "⛔ This command may delete multiple resources across namespaces")
	}

	if reWildcardSelector.MatchString(c) {
		plan.Notes = append(plan.Notes, "🚨 CRITICAL: Wildcard selector detected!")
		plan.RequireSecondConfirm = true
		plan.RequireTypedConfirm = true
		plan.DangerLevel = "critical"
		plan.SafetyChecks = append(plan.SafetyChecks, "⛔ Wildcard selectors can affect ALL resources")
	}

	if reForceDelete.MatchString(c) {
		plan.Notes = append(plan.Notes, "⚠️  HIGH RISK: Force delete detected!")
		plan.RequireSecondConfirm = true
		plan.RequireTypedConfirm = true
		plan.DangerLevel = "high"
		plan.SafetyChecks = append(plan.SafetyChecks, "⚠️  Force delete bypasses graceful termination")
	}

	if reNamespaceWide.MatchString(c) && reKubectlDelete.MatchString(c) {
		plan.Notes = append(plan.Notes, "🚨 CRITICAL: Cross-namespace delete detected!")
		plan.RequireSecondConfirm = true
		plan.RequireTypedConfirm = true
		plan.DangerLevel = "critical"
		plan.SafetyChecks = append(plan.SafetyChecks, "⛔ This affects ALL namespaces in the cluster")
	}

	if reDangerousResourceTypes.MatchString(c) && reKubectlDelete.MatchString(c) {
		plan.Notes = append(plan.Notes, "🚨 CRITICAL: Cluster-level resource deletion!")
		plan.RequireSecondConfirm = true
		plan.RequireTypedConfirm = true
		plan.DangerLevel = "critical"
		plan.SafetyChecks = append(plan.SafetyChecks, "⛔ Deleting cluster-level resources can break the entire cluster")
	}

	if reDeleteWithoutNamespace.MatchString(c) && !strings.Contains(c, "get") {
		plan.Notes = append(plan.Notes, "⚠️  MEDIUM RISK: Delete without explicit namespace")
		plan.RequireSecondConfirm = true
		if plan.DangerLevel == "low" {
			plan.DangerLevel = "medium"
		}
		plan.SafetyChecks = append(plan.SafetyChecks, "⚠️  Consider specifying -n <namespace> for safety")
	}

	if reWildcardResources.MatchString(c) {
		plan.Notes = append(plan.Notes, "🚨 CRITICAL: Wildcard resource pattern detected!")
		plan.RequireSecondConfirm = true
		plan.RequireTypedConfirm = true
		plan.DangerLevel = "critical"
		plan.SafetyChecks = append(plan.SafetyChecks, "⛔ Wildcard patterns can affect unintended resources")
	}

	// kubectl delete/apply/create/patch → enforce --dry-run=client first.
	if reKubectlDelete.MatchString(c) || reKubectlApplyCreatePatch.MatchString(c) {
		if !strings.Contains(c, "--dry-run") {
			dr := c + " --dry-run=client"
			plan.FirstRunCommand = dr
			plan.RequireSecondConfirm = true
			plan.Notes = append(plan.Notes, "🔍 Mutating kubectl detected → added --dry-run=client for first run.")
			plan.SafetyChecks = append(plan.SafetyChecks, "✅ Dry-run will validate changes without applying them")
			if plan.DangerLevel == "low" {
				plan.DangerLevel = "medium"
			}
		} else {
			plan.FirstRunCommand = c
			plan.RequireSecondConfirm = true
			plan.Notes = append(plan.Notes, "🔍 Dry-run detected; will require second confirm to apply for real.")
		}

		// Add specific validation checks for kubectl operations
		if reKubectlDelete.MatchString(c) {
			plan.Checks = append(plan.Checks, PreviewCheck{
				Name: "Resource validation",
				Cmd:  strings.Replace(c, "delete", "get", 1) + " --dry-run=client",
			})
		}

		return plan
	}

	// helm install/upgrade → comprehensive validation pipeline
	if reHelmInstallUpgrade.MatchString(c) {
		rel := extractHelmReleaseName(c)
		chartPath := extractHelmChartPath(c)

		// Enhanced Helm validation pipeline
		plan.Checks = append(plan.Checks,
			PreviewCheck{Name: "helm dependency check", Cmd: fmt.Sprintf("helm dependency list %s", chartPath)},
			PreviewCheck{Name: "helm lint", Cmd: fmt.Sprintf("helm lint %s", chartPath)},
			PreviewCheck{Name: "helm template (dry)", Cmd: fmt.Sprintf("helm template %s %s --dry-run --debug", rel, chartPath)},
			PreviewCheck{Name: "kubectl dry-run validation", Cmd: fmt.Sprintf("helm template %s %s | kubectl apply --dry-run=client -f -", rel, chartPath)},
		)

		plan.FirstRunCommand = c + " --dry-run"
		plan.RequireSecondConfirm = true
		plan.DangerLevel = "medium"

		if !strings.Contains(c, "--install") && strings.HasPrefix(strings.TrimSpace(c), "helm upgrade") {
			plan.Notes = append(plan.Notes, "💡 Consider adding --install for first-time upgrades.")
		}

		plan.Notes = append(plan.Notes, "📦 Helm operation → running comprehensive validation pipeline.")
		plan.SafetyChecks = append(plan.SafetyChecks, "✅ Dependencies, linting, templating, and kubectl validation will be performed")

		// Check for production namespace indicators
		if strings.Contains(c, "production") || strings.Contains(c, "prod") || strings.Contains(c, "live") {
			plan.Notes = append(plan.Notes, "🚨 PRODUCTION DEPLOYMENT DETECTED!")
			plan.RequireTypedConfirm = true
			plan.DangerLevel = "high"
			plan.SafetyChecks = append(plan.SafetyChecks, "⚠️  This appears to target a production environment")
		}

		return plan
	}

	// Unknown tool → treat as potentially dangerous, require validation
	plan.FirstRunCommand = c
	plan.RequireSecondConfirm = true
	plan.DangerLevel = "medium"
	plan.Notes = append(plan.Notes, "❓ Unknown command - please verify before execution.")
	plan.SafetyChecks = append(plan.SafetyChecks, "⚠️  Unrecognized command type - exercise caution")

	// Check for potentially dangerous patterns in unknown commands
	dangerousPatterns := []string{"rm ", "delete", "drop", "truncate", "destroy", "purge"}
	for _, pattern := range dangerousPatterns {
		if strings.Contains(strings.ToLower(c), pattern) {
			plan.Notes = append(plan.Notes, "🚨 POTENTIAL DATA LOSS: Destructive operation detected!")
			plan.RequireTypedConfirm = true
			plan.DangerLevel = "high"
			plan.SafetyChecks = append(plan.SafetyChecks, "⛔ Command appears to perform destructive operations")
			break
		}
	}

	return plan
}

// HumanPreview renders a short explanation for the UI preview panel.
func (p PreExecPlan) HumanPreview() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Pre-exec plan:\n• Original: %s\n", p.Original)
	if len(p.Checks) > 0 {
		fmt.Fprintf(&b, "• Checks:\n")
		for _, ch := range p.Checks {
			fmt.Fprintf(&b, "   - %s: %s\n", ch.Name, ch.Cmd)
		}
	}
	if p.FirstRunCommand != "" && p.FirstRunCommand != p.Original {
		fmt.Fprintf(&b, "• First run (safe): %s\n", p.FirstRunCommand)
	} else {
		fmt.Fprintf(&b, "• First run: %s\n", p.FirstRunCommand)
	}
	if p.RequireSecondConfirm {
		fmt.Fprintf(&b, "• Second confirm required to execute for real.\n")
	}
	for _, n := range p.Notes {
		fmt.Fprintf(&b, "• Note: %s\n", n)
	}
	return b.String()
}

// analyzeDangerLevel evaluates the overall risk level of a command
func analyzeDangerLevel(cmd string, plan PreExecPlan) PreExecPlan {
	// Start with base danger assessment
	if reKubectlDelete.MatchString(cmd) {
		plan.DangerLevel = "medium"
	} else if reKubectlApplyCreatePatch.MatchString(cmd) {
		plan.DangerLevel = "medium"
	} else if reHelmInstallUpgrade.MatchString(cmd) {
		plan.DangerLevel = "medium"
	}

	// Escalate based on specific patterns
	dangerPatterns := []struct {
		pattern *regexp.Regexp
		level   string
		message string
	}{
		{reDangerousDelete, "critical", "Bulk deletion pattern"},
		{reWildcardSelector, "critical", "Wildcard selector usage"},
		{reForceDelete, "high", "Forced operation"},
		{reNamespaceWide, "high", "Cross-namespace operation"},
		{reDangerousResourceTypes, "high", "Cluster-level resource"},
		{reWildcardResources, "critical", "Wildcard resource pattern"},
	}

	for _, dp := range dangerPatterns {
		if dp.pattern.MatchString(cmd) {
			if shouldEscalateDanger(plan.DangerLevel, dp.level) {
				plan.DangerLevel = dp.level
			}
		}
	}

	return plan
}

// shouldEscalateDanger determines if danger level should be escalated
func shouldEscalateDanger(current, new string) bool {
	levels := map[string]int{
		"low":      1,
		"medium":   2,
		"high":     3,
		"critical": 4,
	}
	return levels[new] > levels[current]
}

// extractHelmChartPath extracts the chart path from a helm command
func extractHelmChartPath(cmd string) string {
	parts := strings.Fields(cmd)

	// Find the position of install/upgrade
	var releaseNamePos = -1
	for i, part := range parts {
		if part == "install" || part == "upgrade" {
			if i+1 < len(parts) {
				releaseNamePos = i + 1
			}
			break
		}
	}

	// Chart path should be the next positional argument after release name
	if releaseNamePos != -1 {
		for i := releaseNamePos + 1; i < len(parts); i++ {
			part := parts[i]
			// Skip flags
			if strings.HasPrefix(part, "-") {
				// Skip flag and its value if it's a short flag
				if !strings.HasPrefix(part, "--") && i+1 < len(parts) && !strings.HasPrefix(parts[i+1], "-") {
					i++ // Skip the flag value
				}
				continue
			}
			// This should be the chart path
			return part
		}
	}

	return "." // Default fallback
}

// GetDangerLevelEmoji returns appropriate emoji for danger level
func (p PreExecPlan) GetDangerLevelEmoji() string {
	switch p.DangerLevel {
	case "critical":
		return "🚨"
	case "high":
		return "⚠️"
	case "medium":
		return "🔶"
	default:
		return "✅"
	}
}

// GetSafetyReport generates a comprehensive safety report
func (p PreExecPlan) GetSafetyReport() string {
	var b strings.Builder

	fmt.Fprintf(&b, "%s Danger Level: %s\n", p.GetDangerLevelEmoji(), strings.ToUpper(p.DangerLevel))

	if len(p.SafetyChecks) > 0 {
		fmt.Fprintf(&b, "\n🔍 Safety Analysis:\n")
		for _, check := range p.SafetyChecks {
			fmt.Fprintf(&b, "  %s\n", check)
		}
	}

	if len(p.Checks) > 0 {
		fmt.Fprintf(&b, "\n🧪 Pre-execution Validation:\n")
		for _, check := range p.Checks {
			fmt.Fprintf(&b, "  • %s\n", check.Name)
		}
	}

	if p.RequireTypedConfirm {
		fmt.Fprintf(&b, "\n⚠️ This command requires typing 'yes' to confirm execution.\n")
	} else if p.RequireSecondConfirm {
		fmt.Fprintf(&b, "\n🔄 This command requires confirmation after dry-run.\n")
	}

	return b.String()
}
