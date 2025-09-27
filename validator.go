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
	reDangerousDelete         = regexp.MustCompile(`delete\s+(all|--all-namespaces|--all)`)
	reWildcardSelector        = regexp.MustCompile(`-l\s*["']?\*["']?|--selector\s*["']?\*["']?`)
	reForceDelete             = regexp.MustCompile(`--force|--grace-period=0`)
	reNamespaceWide           = regexp.MustCompile(`--all-namespaces|-A\b`)
)

// PreviewCheck is a command we run BEFORE any mutating action.
type PreviewCheck struct {
Name string
Cmd  string
}

type PreExecPlan struct {
Original            string
Checks              []PreviewCheck // e.g., helm lint/template, extra sanity
FirstRunCommand     string         // typically a --dry-run variant
RequireSecondConfirm bool          // after dry-run/lint succeeded, ask again to apply for real
Notes               []string
}

// BuildPreExecPlan decides the safe preview flow for a command.
func BuildPreExecPlan(cmd string) PreExecPlan {
c := strings.TrimSpace(cmd)
plan := PreExecPlan{Original: c}

// Read-only kubectl → execute directly.
if reReadOnlyKubectl.MatchString(c) {
plan.FirstRunCommand = c
return plan
}

	// Enhanced danger detection
	if reDangerousDelete.MatchString(c) {
		plan.Notes = append(plan.Notes, "DANGEROUS COMMAND! This may delete multiple resources.")
		plan.RequireSecondConfirm = true
	}
	if reWildcardSelector.MatchString(c) {
		plan.Notes = append(plan.Notes, "⚠️  DANGEROUS: Wildcard selector detected!")
		plan.RequireSecondConfirm = true
	}
	if reForceDelete.MatchString(c) {
		plan.Notes = append(plan.Notes, "⚠️  DANGEROUS: Force delete detected!")
		plan.RequireSecondConfirm = true
	}
	if reNamespaceWide.MatchString(c) && reKubectlDelete.MatchString(c) {
		plan.Notes = append(plan.Notes, "⚠️  DANGEROUS: Cross-namespace delete detected!")
		plan.RequireSecondConfirm = true
	}

	// kubectl delete/apply/create/patch → enforce --dry-run=client first.
	if reKubectlDelete.MatchString(c) || reKubectlApplyCreatePatch.MatchString(c) {
		if !strings.Contains(c, "--dry-run") {
			dr := c + " --dry-run=client"
			plan.FirstRunCommand = dr
			plan.RequireSecondConfirm = true
			plan.Notes = append(plan.Notes, "🔍 Mutating kubectl detected → added --dry-run=client for first run.")
		} else {
			plan.FirstRunCommand = c
			plan.RequireSecondConfirm = true
			plan.Notes = append(plan.Notes, "🔍 Dry-run detected; will require second confirm to apply for real.")
		}
		return plan
	}

	// helm install/upgrade → helm lint + template --dry-run first.
	if reHelmInstallUpgrade.MatchString(c) {
		rel := extractHelmReleaseName(c)
		plan.Checks = append(plan.Checks,
			PreviewCheck{Name: "helm lint", Cmd: "helm lint ."},
			PreviewCheck{Name: "helm template (dry)", Cmd: fmt.Sprintf("helm template %s . --dry-run --debug", rel)},
		)
		plan.FirstRunCommand = c + " --dry-run"
		if !strings.Contains(c, "--install") && strings.HasPrefix(strings.TrimSpace(c), "helm upgrade") {
			plan.Notes = append(plan.Notes, "💡 Consider adding --install for first-time upgrades.")
		}
		plan.RequireSecondConfirm = true
		plan.Notes = append(plan.Notes, "📦 Helm operation → running lint/template and first as --dry-run.")
		return plan
	}

	// Unknown tool → treat as potentially dangerous, ask to run as-is (user decides).
	plan.FirstRunCommand = c
	plan.Notes = append(plan.Notes, "❓ Unknown command - please verify before execution.")
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

