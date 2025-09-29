package validator

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	
	"github.com/siryoos/kubemage/internal/workspace"
)

// ValidationResult represents the result of a validation operation
type ValidationResult struct {
	Type     string        `json:"type"` // "kubectl", "helm-lint", "helm-template"
	Success  bool          `json:"success"`
	Output   string        `json:"output"`
	Errors   []string      `json:"errors"`
	Warnings []string      `json:"warnings"`
	Duration time.Duration `json:"duration"`
}

// ValidationPipeline manages validation of files and directories
type ValidationPipeline struct {
	KubectlPath string
	HelmPath    string
	Timeout     time.Duration
}

const manifestProbeLimit = 32 * 1024

func isYAMLFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yaml" || ext == ".yml"
}

func isLikelyManifest(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	buf := make([]byte, manifestProbeLimit)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return false
	}

	lower := strings.ToLower(string(buf[:n]))
	return strings.Contains(lower, "apiversion:") && strings.Contains(lower, "kind:")
}

func findChartRoot(startPath string) string {
	info, err := os.Stat(startPath)
	if err != nil {
		return ""
	}

	dir := startPath
	if !info.IsDir() {
		dir = filepath.Dir(startPath)
	}

	for {
		chartYaml := filepath.Join(dir, "Chart.yaml")
		if stat, err := os.Stat(chartYaml); err == nil && !stat.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return ""
}

func ensureWorkspace() {
	if workspace.WorkspaceIdx == nil {
		workspace.InitializeWorkspace()
	}
}

func resolveWorkspacePaths(path string) (abs string, rel string) {
	ensureWorkspace()

	if workspace.WorkspaceIdx != nil {
		rel = workspace.WorkspaceIdx.NormalizePath(path)
		abs = workspace.WorkspaceIdx.AbsPath(rel)
		return
	}

	clean := filepath.Clean(path)
	if filepath.IsAbs(clean) {
		abs = clean
	} else {
		if a, err := filepath.Abs(clean); err == nil {
			abs = a
		} else {
			abs = clean
		}
	}

	rel = filepath.ToSlash(clean)
	return
}

// NewValidationPipeline creates a new validation pipeline
func NewValidationPipeline() *ValidationPipeline {
	return &ValidationPipeline{
		KubectlPath: "kubectl",
		HelmPath:    "helm",
		Timeout:     10 * time.Second,
	}
}

// ValidateKubernetesManifest validates a Kubernetes YAML file using kubectl dry-run
func (vp *ValidationPipeline) ValidateKubernetesManifest(filePath string) *ValidationResult {
	start := time.Now()
	result := &ValidationResult{
		Type: "kubectl",
	}

	// kubectl apply --dry-run=client -f <file>
	cmd := exec.Command(vp.KubectlPath, "apply", "--dry-run=client", "-f", filePath)
	output, err := cmd.CombinedOutput()
	result.Duration = time.Since(start)
	result.Output = string(output)

	if err != nil {
		result.Success = false
		result.Errors = append(result.Errors, fmt.Sprintf("kubectl validation failed: %v", err))

		// Parse kubectl error output for specific issues
		outputStr := strings.ToLower(string(output))
		if strings.Contains(outputStr, "unable to recognize") {
			result.Errors = append(result.Errors, "Invalid Kubernetes resource format")
		}
		if strings.Contains(outputStr, "missing required field") {
			result.Errors = append(result.Errors, "Missing required fields in manifest")
		}
		if strings.Contains(outputStr, "forbidden") {
			result.Errors = append(result.Errors, "Insufficient permissions or policy violation")
		}
	} else {
		result.Success = true

		// Parse warnings from successful output
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(strings.ToLower(line), "warning") {
				result.Warnings = append(result.Warnings, strings.TrimSpace(line))
			}
		}
	}

	return result
}

// ValidateHelmChart runs helm lint on a chart directory
func (vp *ValidationPipeline) ValidateHelmChart(chartPath string) *ValidationResult {
	start := time.Now()
	result := &ValidationResult{
		Type: "helm-lint",
	}

	// helm lint <chart-path>
	cmd := exec.Command(vp.HelmPath, "lint", chartPath)
	output, err := cmd.CombinedOutput()
	result.Duration = time.Since(start)
	result.Output = string(output)

	if err != nil {
		result.Success = false
		result.Errors = append(result.Errors, fmt.Sprintf("helm lint failed: %v", err))
	} else {
		result.Success = true
	}

	// Parse helm lint output for errors and warnings
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		lowerLine := strings.ToLower(line)
		if strings.Contains(lowerLine, "[error]") {
			result.Errors = append(result.Errors, strings.TrimSpace(line))
			result.Success = false
		} else if strings.Contains(lowerLine, "[warning]") {
			result.Warnings = append(result.Warnings, strings.TrimSpace(line))
		}
	}

	return result
}

// ValidateHelmTemplate runs helm template to validate chart rendering
func (vp *ValidationPipeline) ValidateHelmTemplate(chartPath, releaseName string) *ValidationResult {
	start := time.Now()
	result := &ValidationResult{
		Type: "helm-template",
	}

	if releaseName == "" {
		releaseName = "test-release"
	}

	// helm template <release-name> <chart-path> --dry-run
	cmd := exec.Command(vp.HelmPath, "template", releaseName, chartPath, "--dry-run")
	output, err := cmd.CombinedOutput()
	result.Duration = time.Since(start)
	result.Output = string(output)

	if err != nil {
		result.Success = false
		result.Errors = append(result.Errors, fmt.Sprintf("helm template failed: %v", err))

		// Parse template error output
		outputStr := strings.ToLower(string(output))
		if strings.Contains(outputStr, "template") && strings.Contains(outputStr, "error") {
			result.Errors = append(result.Errors, "Template rendering error")
		}
		if strings.Contains(outputStr, "values") && strings.Contains(outputStr, "error") {
			result.Errors = append(result.Errors, "Values file error")
		}
	} else {
		result.Success = true
	}

	return result
}

// ValidateFile runs appropriate validation based on file type and location
func (vp *ValidationPipeline) ValidateFile(filePath string) []*ValidationResult {
	absPath, relPath := resolveWorkspacePaths(filePath)
	info, err := os.Stat(absPath)
	if err != nil {
		return []*ValidationResult{{
			Type:    "error",
			Success: false,
			Errors:  []string{fmt.Sprintf("File not found: %v", err)},
		}}
	}

	if info.IsDir() {
		return vp.ValidateDirectory(absPath)
	}

	chartRoot := findChartRoot(absPath)
	if chartRoot == "" && workspace.WorkspaceIdx != nil {
		if ok, chartDir := workspace.WorkspaceIdx.IsUnderHelmChart(relPath); ok {
			chartRoot = workspace.WorkspaceIdx.AbsPath(chartDir)
		}
	}

	var results []*ValidationResult
	if chartRoot != "" {
		results = append(results, vp.ValidateHelmChart(chartRoot))
		results = append(results, vp.ValidateHelmTemplate(chartRoot, ""))
		return results
	}

	if !isYAMLFile(absPath) {
		return []*ValidationResult{{
			Type:    "skip",
			Success: true,
			Output:  "Validation skipped (non-YAML file)",
		}}
	}

	if !isLikelyManifest(absPath) {
		return []*ValidationResult{{
			Type:    "skip",
			Success: true,
			Output:  "Validation skipped (no Kubernetes manifest detected)",
		}}
	}

	results = append(results, vp.ValidateKubernetesManifest(absPath))
	return results
}

// ValidatePaths runs validation across multiple paths and groups the results per path
func (vp *ValidationPipeline) ValidatePaths(paths []string) map[string][]*ValidationResult {
	resultMap := make(map[string][]*ValidationResult)
	for _, p := range paths {
		if strings.TrimSpace(p) == "" {
			continue
		}
		resultMap[p] = vp.ValidateFile(p)
	}
	return resultMap
}

// ValidateDirectory runs validation on all relevant files in a directory
func (vp *ValidationPipeline) ValidateDirectory(dirPath string) []*ValidationResult {
	var results []*ValidationResult

	// Check if it's a helm chart directory
	absDir, _ := resolveWorkspacePaths(dirPath)
	chartYaml := filepath.Join(absDir, "Chart.yaml")
	if _, err := os.Stat(chartYaml); err == nil {
		// It's a helm chart
		results = append(results, vp.ValidateHelmChart(absDir))
		results = append(results, vp.ValidateHelmTemplate(absDir, ""))
	} else {
		// Look for kubernetes manifests
		err := filepath.Walk(absDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && (strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")) {
				if isLikelyManifest(path) {
					result := vp.ValidateKubernetesManifest(path)
					results = append(results, result)
				}
			}
			return nil
		})

		if err != nil {
			results = append(results, &ValidationResult{
				Type:    "error",
				Success: false,
				Errors:  []string{fmt.Sprintf("Error walking directory: %v", err)},
			})
		}
	}

	return results
}

// RenderValidationResults renders validation results for display
func RenderValidationResults(results []*ValidationResult) string {
	if len(results) == 0 {
		return "✅ No validation needed"
	}

	var output strings.Builder

	output.WriteString("🔍 Validation Results\n")
	output.WriteString(strings.Repeat("─", 40) + "\n")

	allPassed := true
	totalTime := time.Duration(0)

	for i, result := range results {
		totalTime += result.Duration

		// Status icon
		icon := "✅"
		if !result.Success {
			icon = "❌"
			allPassed = false
		} else if len(result.Warnings) > 0 {
			icon = "⚠️"
		} else if result.Type == "skip" {
			icon = "ℹ️"
		}

		output.WriteString(fmt.Sprintf("%s %s (%s)\n", icon, result.Type, result.Duration.Round(time.Millisecond)))

		// Show errors
		for _, err := range result.Errors {
			output.WriteString(fmt.Sprintf("  ❌ %s\n", err))
		}

		// Show warnings
		for _, warning := range result.Warnings {
			output.WriteString(fmt.Sprintf("  ⚠️  %s\n", warning))
		}

		// Show success message if no errors/warnings
		if strings.TrimSpace(result.Output) != "" && (len(result.Errors) > 0 || len(result.Warnings) > 0) {
			trimmed := strings.TrimSpace(result.Output)
			if len(trimmed) > 0 {
				output.WriteString(fmt.Sprintf("  📄 %s\n", trimmed))
			}
		}

		if result.Success && len(result.Errors) == 0 && len(result.Warnings) == 0 {
			output.WriteString("  ✅ Validation passed\n")
		}

		if result.Type == "skip" && strings.TrimSpace(result.Output) != "" {
			output.WriteString(fmt.Sprintf("  ℹ️  %s\n", strings.TrimSpace(result.Output)))
		}

		if i < len(results)-1 {
			output.WriteString("\n")
		}
	}

	output.WriteString(strings.Repeat("─", 40) + "\n")

	if allPassed {
		output.WriteString(fmt.Sprintf("✅ All validations passed (total: %s)\n", totalTime.Round(time.Millisecond)))
	} else {
		output.WriteString(fmt.Sprintf("❌ Some validations failed (total: %s)\n", totalTime.Round(time.Millisecond)))
	}

	return output.String()
}

// CheckToolAvailability checks if required tools are available
func (vp *ValidationPipeline) CheckToolAvailability() map[string]bool {
	tools := make(map[string]bool)

	// Check kubectl
	if _, err := exec.LookPath(vp.KubectlPath); err == nil {
		tools["kubectl"] = true
	} else {
		tools["kubectl"] = false
	}

	// Check helm
	if _, err := exec.LookPath(vp.HelmPath); err == nil {
		tools["helm"] = true
	} else {
		tools["helm"] = false
	}

	return tools
}

// Removed global instance - now created via dependency injection

// InitializeValidation initializes the validation pipeline (deprecated - use NewValidationPipeline directly)
func InitializeValidation() *ValidationPipeline {
	return NewValidationPipeline()
}
