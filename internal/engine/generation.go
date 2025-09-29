package engine

import (
	"fmt"
	"path/filepath"
	"strings"
	
	"github.com/siryoos/kubemage/internal/engine/validator"
)

// GenerationTemplate represents a template for generating files
type GenerationTemplate struct {
	Name        string
	Description string
	Template    string
	FileExt     string
}

// DeploymentOptions represents options for deployment generation
type DeploymentOptions struct {
	Name     string
	Image    string
	Replicas int
	Port     int
	Labels   map[string]string
	Env      map[string]string
}

// HelmChartOptions represents options for helm chart generation
type HelmChartOptions struct {
	Name         string
	Description  string
	Version      string
	AppVersion   string
	Dependencies []string
}

// GetDeploymentGenerationPrompt creates a prompt for LLM to generate deployment YAML
func GetDeploymentGenerationPrompt(options DeploymentOptions) string {
	var prompt strings.Builder

	prompt.WriteString("Generate a Kubernetes Deployment YAML manifest with the following specifications:\n\n")
	prompt.WriteString(fmt.Sprintf("- Name: %s\n", options.Name))
	prompt.WriteString(fmt.Sprintf("- Container Image: %s\n", options.Image))

	if options.Replicas > 0 {
		prompt.WriteString(fmt.Sprintf("- Replicas: %d\n", options.Replicas))
	} else {
		prompt.WriteString("- Replicas: 3 (default)\n")
	}

	if options.Port > 0 {
		prompt.WriteString(fmt.Sprintf("- Container Port: %d\n", options.Port))
	}

	if len(options.Labels) > 0 {
		prompt.WriteString("- Additional Labels:\n")
		for key, value := range options.Labels {
			prompt.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	if len(options.Env) > 0 {
		prompt.WriteString("- Environment Variables:\n")
		for key, value := range options.Env {
			prompt.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	prompt.WriteString("\nRequirements:\n")
	prompt.WriteString("- Use current Kubernetes API versions\n")
	prompt.WriteString("- Include proper metadata labels (app.kubernetes.io/name, app.kubernetes.io/instance)\n")
	prompt.WriteString("- Add resource requests and limits\n")
	prompt.WriteString("- Include readiness and liveness probes if a port is specified\n")
	prompt.WriteString("- Follow Kubernetes best practices\n")
	prompt.WriteString("- Output only the YAML manifest, no explanations\n")

	return prompt.String()
}

// GetServiceGenerationPrompt creates a prompt for LLM to generate service YAML
func GetServiceGenerationPrompt(name string, port int, targetPort int) string {
	var prompt strings.Builder

	prompt.WriteString("Generate a Kubernetes Service YAML manifest with the following specifications:\n\n")
	prompt.WriteString(fmt.Sprintf("- Name: %s-service\n", name))
	prompt.WriteString(fmt.Sprintf("- Selector: app.kubernetes.io/name: %s\n", name))
	prompt.WriteString(fmt.Sprintf("- Port: %d\n", port))

	if targetPort > 0 {
		prompt.WriteString(fmt.Sprintf("- Target Port: %d\n", targetPort))
	} else {
		prompt.WriteString(fmt.Sprintf("- Target Port: %d (same as port)\n", port))
	}

	prompt.WriteString("\nRequirements:\n")
	prompt.WriteString("- Use ClusterIP service type\n")
	prompt.WriteString("- Include proper metadata labels\n")
	prompt.WriteString("- Follow Kubernetes best practices\n")
	prompt.WriteString("- Output only the YAML manifest, no explanations\n")

	return prompt.String()
}

// GetHelmChartGenerationPrompt creates a prompt for LLM to generate helm chart structure
func GetHelmChartGenerationPrompt(options HelmChartOptions) string {
	var prompt strings.Builder

	prompt.WriteString("Generate a Helm chart structure with the following specifications:\n\n")
	prompt.WriteString(fmt.Sprintf("- Chart Name: %s\n", options.Name))

	if options.Description != "" {
		prompt.WriteString(fmt.Sprintf("- Description: %s\n", options.Description))
	}

	if options.Version != "" {
		prompt.WriteString(fmt.Sprintf("- Chart Version: %s\n", options.Version))
	} else {
		prompt.WriteString("- Chart Version: 0.1.0\n")
	}

	if options.AppVersion != "" {
		prompt.WriteString(fmt.Sprintf("- App Version: %s\n", options.AppVersion))
	} else {
		prompt.WriteString("- App Version: 1.0.0\n")
	}

	prompt.WriteString("\nGenerate the following files:\n\n")

	prompt.WriteString("1. **Chart.yaml** - Chart metadata\n")
	prompt.WriteString("2. **values.yaml** - Default configuration values\n")
	prompt.WriteString("3. **templates/deployment.yaml** - Deployment template\n")
	prompt.WriteString("4. **templates/service.yaml** - Service template\n")
	prompt.WriteString("5. **templates/_helpers.tpl** - Template helpers\n")

	prompt.WriteString("\nRequirements:\n")
	prompt.WriteString("- Use Helm v3 syntax\n")
	prompt.WriteString("- Include proper templating with values\n")
	prompt.WriteString("- Add common labels and annotations helpers\n")
	prompt.WriteString("- Include resource requests/limits templating\n")
	prompt.WriteString("- Add health check templating\n")
	prompt.WriteString("- Follow Helm best practices\n")
	prompt.WriteString("- Output each file clearly separated with file path headers\n")

	return prompt.String()
}

// GetEditValuesPrompt creates a prompt for LLM to edit values files
func GetEditValuesPrompt(currentContent, instruction string) string {
	var prompt strings.Builder

	prompt.WriteString("Edit the following Helm values.yaml file according to the instruction.\n\n")
	prompt.WriteString("**Current content:**\n")
	prompt.WriteString("```yaml\n")
	prompt.WriteString(currentContent)
	prompt.WriteString("\n```\n\n")

	prompt.WriteString("**Instruction:**\n")
	prompt.WriteString(instruction)
	prompt.WriteString("\n\n")

	prompt.WriteString("**Requirements:**\n")
	prompt.WriteString("- Provide the complete modified file content\n")
	prompt.WriteString("- Maintain YAML formatting and structure\n")
	prompt.WriteString("- Preserve comments where appropriate\n")
	prompt.WriteString("- Follow Helm values best practices\n")
	prompt.WriteString("- Output only the YAML content, no explanations\n")

	return prompt.String()
}

// GetEditYamlPrompt creates a prompt for LLM to edit YAML manifests
func GetEditYamlPrompt(currentContent, instruction string) string {
	var prompt strings.Builder

	prompt.WriteString("Edit the following Kubernetes YAML manifest according to the instruction.\n\n")
	prompt.WriteString("**Current content:**\n")
	prompt.WriteString("```yaml\n")
	prompt.WriteString(currentContent)
	prompt.WriteString("\n```\n\n")

	prompt.WriteString("**Instruction:**\n")
	prompt.WriteString(instruction)
	prompt.WriteString("\n\n")

	prompt.WriteString("**Requirements:**\n")
	prompt.WriteString("- Provide the complete modified file content\n")
	prompt.WriteString("- Maintain valid Kubernetes YAML format\n")
	prompt.WriteString("- Preserve existing structure unless changes are required\n")
	prompt.WriteString("- Use current Kubernetes API versions\n")
	prompt.WriteString("- Follow Kubernetes best practices\n")
	prompt.WriteString("- Output only the YAML content, no explanations\n")

	return prompt.String()
}

// ParseGeneratedContent extracts YAML content from LLM response
func ParseGeneratedContent(response string) string {
	// Remove markdown code fences if present
	content := response

	// Look for yaml code blocks
	yamlStart := strings.Index(content, "```yaml")
	if yamlStart != -1 {
		yamlStart += 7 // Skip "```yaml"
		yamlEnd := strings.Index(content[yamlStart:], "```")
		if yamlEnd != -1 {
			content = content[yamlStart : yamlStart+yamlEnd]
		}
	} else {
		// Look for generic code blocks
		codeStart := strings.Index(content, "```")
		if codeStart != -1 {
			codeStart += 3
			codeEnd := strings.Index(content[codeStart:], "```")
			if codeEnd != -1 {
				content = content[codeStart : codeStart+codeEnd]
			}
		}
	}

	return strings.TrimSpace(content)
}

// SaveGeneratedFile saves generated content to a file in the out directory
func SaveGeneratedFile(filename, content string) (string, error) {
	outPath := filepath.Join("out", filename)

	if err := WriteFileContent(outPath, content); err != nil {
		return "", err
	}

	return outPath, nil
}

// ParseHelmChartFiles parses multiple files from helm chart generation response
func ParseHelmChartFiles(response string) map[string]string {
	files := make(map[string]string)

	// Split by file separators (look for file path headers)
	lines := strings.Split(response, "\n")
	currentFile := ""
	var currentContent strings.Builder

	for _, line := range lines {
		// Look for file path patterns
		if strings.Contains(line, ".yaml") || strings.Contains(line, ".tpl") {
			// Save previous file if exists
			if currentFile != "" && currentContent.Len() > 0 {
				files[currentFile] = strings.TrimSpace(currentContent.String())
			}

			// Start new file
			if strings.HasPrefix(line, "**") && strings.HasSuffix(line, "**") {
				currentFile = strings.Trim(line, "* ")
			} else if strings.Contains(line, "/") {
				currentFile = strings.TrimSpace(line)
			}
			currentContent.Reset()
		} else if currentFile != "" {
			// Add line to current file content
			if !strings.HasPrefix(line, "```") && line != "" {
				currentContent.WriteString(line + "\n")
			}
		}
	}

	// Save last file
	if currentFile != "" && currentContent.Len() > 0 {
		files[currentFile] = strings.TrimSpace(currentContent.String())
	}

	return files
}

// SaveHelmChart saves a complete helm chart to the out directory
func SaveHelmChart(chartName string, files map[string]string) (string, error) {
	chartPath := filepath.Join("out", chartName)

	for filePath, content := range files {
		fullPath := filepath.Join(chartPath, filePath)
		if err := WriteFileContent(fullPath, content); err != nil {
			return "", fmt.Errorf("failed to save %s: %v", filePath, err)
		}
	}

	return chartPath, nil
}

// GenerationWorkflow represents a complete generation workflow
type GenerationWorkflow struct {
	Type              string // "deployment", "service", "helm-chart"
	OutputPath        string
	ValidationResults []*ValidationResult
	Success           bool
	Error             error
}

// RunGenerationWorkflow executes a complete generation workflow with validation
func RunGenerationWorkflow(workflowType, content, outputPath string) *GenerationWorkflow {
	workflow := &GenerationWorkflow{
		Type:       workflowType,
		OutputPath: outputPath,
	}

	// Save the generated content
	if err := WriteFileContent(outputPath, content); err != nil {
		workflow.Error = err
		return workflow
	}

	// Initialize validation if needed
	if ValidationPipe == nil {
		InitializeValidation()
	}

	// Run validation based on type
	switch workflowType {
	case "deployment", "service", "manifest":
		workflow.ValidationResults = ValidationPipe.ValidateFile(outputPath)
	case "helm-chart":
		workflow.ValidationResults = ValidationPipe.ValidateDirectory(filepath.Dir(outputPath))
	}

	// Check if all validations passed
	workflow.Success = true
	for _, result := range workflow.ValidationResults {
		if !result.Success {
			workflow.Success = false
			break
		}
	}

	return workflow
}
