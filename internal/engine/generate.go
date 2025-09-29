package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GenerationResult represents the result of a generation operation
type GenerationResult struct {
	Type       string // "k8s" or "helm"
	Name       string // Resource/chart name
	Files      []GeneratedFile
	OutputDir  string   // Where files were saved
	Validation []string // Validation command results
}

type GeneratedFile struct {
	Path    string // Relative path
	Content string // File content
	Saved   bool   // Whether successfully saved
}

// SaveGeneratedContent saves generated content to files and runs validation
func SaveGeneratedContent(content, genType, name string) (*GenerationResult, error) {
	result := &GenerationResult{
		Type:      genType,
		Name:      name,
		OutputDir: "./out",
	}

	// Create output directory
	outputDir := fmt.Sprintf("./out/%s-%s-%d", genType, name, time.Now().Unix())
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}
	result.OutputDir = outputDir

	if genType == "k8s" {
		// Single YAML file for Kubernetes manifests
		filePath := filepath.Join(outputDir, fmt.Sprintf("%s.yaml", name))
		file := GeneratedFile{
			Path:    fmt.Sprintf("%s.yaml", name),
			Content: content,
		}

		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return nil, fmt.Errorf("failed to write file %s: %w", filePath, err)
		}
		file.Saved = true
		result.Files = append(result.Files, file)

		// Validate Kubernetes manifest
		validationCmd := fmt.Sprintf("kubectl apply --dry-run=client -f %s", filePath)
		validationResult, err := runShell(10*time.Second, validationCmd, 4096)
		if err != nil {
			result.Validation = append(result.Validation, fmt.Sprintf("❌ Validation failed: %v", err))
			result.Validation = append(result.Validation, validationResult)
		} else {
			result.Validation = append(result.Validation, "✅ Kubernetes manifest validation passed")
		}

	} else if genType == "helm" {
		// Parse multiple files for Helm charts
		files := parseHelmFiles(content, name)
		for _, file := range files {
			filePath := filepath.Join(outputDir, file.Path)

			// Create subdirectories if needed
			if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
				return nil, fmt.Errorf("failed to create directory for %s: %w", file.Path, err)
			}

			if err := os.WriteFile(filePath, []byte(file.Content), 0644); err != nil {
				return nil, fmt.Errorf("failed to write file %s: %w", filePath, err)
			}
			file.Saved = true
			result.Files = append(result.Files, file)
		}

		// Validate Helm chart
		lintCmd := fmt.Sprintf("helm lint %s", outputDir)
		lintResult, err := runShell(10*time.Second, lintCmd, 4096)
		if err != nil {
			result.Validation = append(result.Validation, fmt.Sprintf("❌ Helm lint failed: %v", err))
			result.Validation = append(result.Validation, lintResult)
		} else {
			result.Validation = append(result.Validation, "✅ Helm lint passed")
		}

		templateCmd := fmt.Sprintf("helm template %s %s --debug --dry-run", name, outputDir)
		templateResult, err := runShell(10*time.Second, templateCmd, 4096)
		if err != nil {
			result.Validation = append(result.Validation, fmt.Sprintf("❌ Helm template failed: %v", err))
			result.Validation = append(result.Validation, templateResult)
		} else {
			result.Validation = append(result.Validation, "✅ Helm template generation passed")
		}
	}

	return result, nil
}

// parseHelmFiles extracts individual files from LLM-generated Helm chart content
func parseHelmFiles(content, chartName string) []GeneratedFile {
	var files []GeneratedFile

	// Split by file separators (common patterns)
	sections := strings.Split(content, "---")
	currentFile := ""
	currentContent := ""

	for _, section := range sections {
		section = strings.TrimSpace(section)
		if section == "" {
			continue
		}

		// Check if this looks like a file header
		lines := strings.Split(section, "\n")
		firstLine := strings.TrimSpace(lines[0])

		if strings.Contains(strings.ToLower(firstLine), "chart.yaml") {
			if currentFile != "" {
				files = append(files, GeneratedFile{Path: currentFile, Content: currentContent})
			}
			currentFile = "Chart.yaml"
			currentContent = strings.Join(lines[1:], "\n")
		} else if strings.Contains(strings.ToLower(firstLine), "values.yaml") {
			if currentFile != "" {
				files = append(files, GeneratedFile{Path: currentFile, Content: currentContent})
			}
			currentFile = "values.yaml"
			currentContent = strings.Join(lines[1:], "\n")
		} else if strings.Contains(strings.ToLower(firstLine), "deployment.yaml") {
			if currentFile != "" {
				files = append(files, GeneratedFile{Path: currentFile, Content: currentContent})
			}
			currentFile = "templates/deployment.yaml"
			currentContent = strings.Join(lines[1:], "\n")
		} else if strings.Contains(strings.ToLower(firstLine), "service.yaml") {
			if currentFile != "" {
				files = append(files, GeneratedFile{Path: currentFile, Content: currentContent})
			}
			currentFile = "templates/service.yaml"
			currentContent = strings.Join(lines[1:], "\n")
		} else {
			// Continuation of current file
			if currentContent != "" {
				currentContent += "\n"
			}
			currentContent += section
		}
	}

	// Add the last file
	if currentFile != "" {
		files = append(files, GeneratedFile{Path: currentFile, Content: currentContent})
	}

	// If no files were parsed, treat the whole content as a single YAML
	if len(files) == 0 {
		files = append(files, GeneratedFile{
			Path:    "templates/deployment.yaml",
			Content: content,
		})
	}

	return files
}

// EnhancedManifestPrompt creates a more detailed prompt for K8s generation
func EnhancedManifestPrompt(kind, name, namespace string) string {
	prompt := fmt.Sprintf(`Generate a complete Kubernetes %s manifest for '%s'`, kind, name)
	if namespace != "" {
		prompt += fmt.Sprintf(` in namespace '%s'`, namespace)
	}
	prompt += `. Requirements:

1. Return ONLY valid YAML without markdown formatting or explanations
2. Include appropriate metadata (name, labels, annotations)
3. Add reasonable resource limits and requests
4. Include security context and best practices
5. Use industry-standard labels (app.kubernetes.io/name, etc.)
6. Add health checks where applicable

Generate the manifest:`

	return prompt
}

// EnhancedHelmPrompt creates a detailed prompt for Helm chart generation
func EnhancedHelmPrompt(name, description string) string {
	prompt := fmt.Sprintf(`Generate a complete Helm chart for application '%s'`, name)
	if description != "" {
		prompt += fmt.Sprintf(` (%s)`, description)
	}
	prompt += `. Requirements:

1. Include Chart.yaml with proper metadata
2. Include values.yaml with configurable parameters
3. Include templates/deployment.yaml with deployment manifest
4. Include templates/service.yaml with service manifest
5. Use Go templating for configurable values
6. Follow Helm best practices and conventions
7. Separate each file with clear headers like "--- Chart.yaml ---"

Generate the chart files:`

	return prompt
}

// Backward compatibility functions
func GetManifestGenerationPrompt(kind string, name string) string {
	return EnhancedManifestPrompt(kind, name, "")
}

func GetHelmGenerationPrompt(name string) string {
	return EnhancedHelmPrompt(name, "")
}
