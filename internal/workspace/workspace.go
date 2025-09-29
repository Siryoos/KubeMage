package workspace

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// WorkspaceFile represents a file in the workspace with metadata
type WorkspaceFile struct {
	Path       string    `json:"path"`
	Kind       string    `json:"kind"`        // kubernetes kind (if applicable) or file type
	FirstLines []string  `json:"first_lines"` // first few lines for preview
	Size       int64     `json:"size"`
	Hash       string    `json:"hash"`
	ModTime    time.Time `json:"mod_time"`
	IsChart    bool      `json:"is_chart"`    // true if under a helm chart
	IsTemplate bool      `json:"is_template"` // true if helm template
	IsValues   bool      `json:"is_values"`   // true if values file
}

// WorkspaceIndex maintains an index of workspace files
type WorkspaceIndex struct {
	Files    map[string]*WorkspaceFile `json:"files"`
	LastScan time.Time                 `json:"last_scan"`
	Root     string                    `json:"root"`
}

var (
	// Patterns for file matching
	reChartPattern    = regexp.MustCompile(`charts/.*/.*`)
	reTemplatePattern = regexp.MustCompile(`templates/.*\.ya?ml$`)
	reValuesPattern   = regexp.MustCompile(`values.*\.ya?ml$`)
	reManifestPattern = regexp.MustCompile(`.*\.ya?ml$`)
	reKubernetesKind  = regexp.MustCompile(`(?m)^kind:\s*([^\s]+)`)
	reHelmChart       = regexp.MustCompile(`Chart\.ya?ml$`)
)

// NewWorkspaceIndex creates a new workspace index
func NewWorkspaceIndex() *WorkspaceIndex {
	return &WorkspaceIndex{
		Files: make(map[string]*WorkspaceFile),
		Root:  ".",
	}
}

// ScanWorkspace scans the current directory for charts, templates, values, and manifests
func (wi *WorkspaceIndex) ScanWorkspace(rootPath string) error {
	rootAbs, err := filepath.Abs(rootPath)
	if err != nil {
		return err
	}

	wi.LastScan = time.Now()
	wi.Root = filepath.Clean(rootAbs)
	wi.Files = make(map[string]*WorkspaceFile) // Reset the index

	return filepath.Walk(rootAbs, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and hidden files
		if info.IsDir() || strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		// Skip non-YAML files for most checks, but include Chart.yaml
		rel, relErr := filepath.Rel(wi.Root, path)
		if relErr != nil {
			rel = path
		}
		cleanRel := filepath.ToSlash(rel)

		if !reManifestPattern.MatchString(cleanRel) && !reHelmChart.MatchString(cleanRel) {
			return nil
		}

		// Get relative path from root
		relPath := cleanRel

		file := &WorkspaceFile{
			Path:    relPath,
			Size:    info.Size(),
			ModTime: info.ModTime(),
		}

		// Determine file type and characteristics
		file.IsChart = reChartPattern.MatchString(relPath) || reHelmChart.MatchString(relPath)
		file.IsTemplate = reTemplatePattern.MatchString(relPath)
		file.IsValues = reValuesPattern.MatchString(relPath)

		// Read file content for analysis
		if err := wi.analyzeFile(path, file); err != nil {
			// Log error but continue scanning
			file.Kind = "error"
			file.FirstLines = []string{fmt.Sprintf("Error reading file: %v", err)}
		}

		wi.Files[relPath] = file
		return nil
	})
}

// analyzeFile reads and analyzes file content
func (wi *WorkspaceIndex) analyzeFile(path string, file *WorkspaceFile) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Calculate hash
	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return err
	}
	file.Hash = fmt.Sprintf("%x", hasher.Sum(nil))[:16] // Shortened hash

	// Reset file pointer for content reading
	f.Seek(0, 0)

	scanner := bufio.NewScanner(f)
	var lines []string
	var content strings.Builder

	// Read first 10 lines and all content for analysis
	lineCount := 0
	for scanner.Scan() {
		line := scanner.Text()
		content.WriteString(line + "\n")

		if lineCount < 10 {
			lines = append(lines, line)
			lineCount++
		}
	}

	file.FirstLines = lines
	fullContent := content.String()

	// Determine file kind
	if file.IsValues {
		file.Kind = "values"
	} else if file.IsTemplate {
		file.Kind = "helm-template"
	} else if reHelmChart.MatchString(file.Path) {
		file.Kind = "chart"
	} else if matches := reKubernetesKind.FindStringSubmatch(fullContent); len(matches) > 1 {
		file.Kind = strings.ToLower(matches[1])
	} else if strings.Contains(fullContent, "apiVersion:") {
		file.Kind = "kubernetes"
	} else {
		file.Kind = "yaml"
	}

	return scanner.Err()
}

// NormalizePath converts the provided path to a workspace-relative, slash-separated path.
func (wi *WorkspaceIndex) NormalizePath(path string) string {
	if wi == nil {
		return filepath.ToSlash(filepath.Clean(path))
	}

	if path == "" {
		return ""
	}

	cleanPath := filepath.Clean(path)
	if filepath.IsAbs(cleanPath) {
		rel, err := filepath.Rel(wi.Root, cleanPath)
		if err == nil {
			return filepath.ToSlash(rel)
		}
	}

	return filepath.ToSlash(cleanPath)
}

// AbsPath resolves a workspace-relative path to an absolute path.
func (wi *WorkspaceIndex) AbsPath(path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}

	root := wi.Root
	if root == "" {
		root = "."
	}

	return filepath.Join(root, filepath.FromSlash(path))
}

// GetFilesByType returns files of a specific type
func (wi *WorkspaceIndex) GetFilesByType(fileType string) []*WorkspaceFile {
	var files []*WorkspaceFile
	for _, file := range wi.Files {
		switch fileType {
		case "values":
			if file.IsValues {
				files = append(files, file)
			}
		case "templates":
			if file.IsTemplate {
				files = append(files, file)
			}
		case "charts":
			if file.IsChart {
				files = append(files, file)
			}
		case "manifests":
			if file.Kind != "values" && file.Kind != "helm-template" && file.Kind != "chart" {
				files = append(files, file)
			}
		default:
			if file.Kind == fileType {
				files = append(files, file)
			}
		}
	}
	return files
}

// FindFile searches for a file by path or partial path
func (wi *WorkspaceIndex) FindFile(searchPath string) *WorkspaceFile {
	// Direct match first
	normalized := wi.NormalizePath(searchPath)
	if file, exists := wi.Files[normalized]; exists {
		return file
	}

	// Partial match - find files containing the search string
	for path, file := range wi.Files {
		if strings.Contains(path, normalized) {
			return file
		}
	}

	return nil
}

// GetHelmCharts returns all helm charts found in the workspace
func (wi *WorkspaceIndex) GetHelmCharts() map[string]*WorkspaceFile {
	charts := make(map[string]*WorkspaceFile)
	for path, file := range wi.Files {
		if file.Kind == "chart" {
			// Extract chart directory
			chartDir := filepath.ToSlash(filepath.Dir(path))
			charts[chartDir] = file
		}
	}
	return charts
}

// IsUnderHelmChart checks if a file is under a helm chart directory
func (wi *WorkspaceIndex) IsUnderHelmChart(filePath string) (bool, string) {
	charts := wi.GetHelmCharts()
	normalized := wi.NormalizePath(filePath)
	for chartDir := range charts {
		if normalized == chartDir || strings.HasPrefix(normalized, chartDir+"/") {
			return true, chartDir
		}
	}
	return false, ""
}

// GetWorkspaceSummary returns a summary of the workspace
func (wi *WorkspaceIndex) GetWorkspaceSummary() string {
	var summary strings.Builder

	summary.WriteString(fmt.Sprintf("📁 Workspace Index (scanned %s)\n", wi.LastScan.Format("15:04:05")))
	summary.WriteString(fmt.Sprintf("Total files: %d\n\n", len(wi.Files)))

	// Count by type
	typeCounts := make(map[string]int)
	for _, file := range wi.Files {
		typeCounts[file.Kind]++
	}

	summary.WriteString("📊 File Types:\n")
	for kind, count := range typeCounts {
		summary.WriteString(fmt.Sprintf("  %s: %d\n", kind, count))
	}

	// Helm charts
	charts := wi.GetHelmCharts()
	if len(charts) > 0 {
		summary.WriteString(fmt.Sprintf("\n⎈ Helm Charts (%d):\n", len(charts)))
		for chartDir := range charts {
			summary.WriteString(fmt.Sprintf("  %s\n", chartDir))
		}
	}

	// Values files
	valuesFiles := wi.GetFilesByType("values")
	if len(valuesFiles) > 0 {
		summary.WriteString(fmt.Sprintf("\n⚙️  Values Files (%d):\n", len(valuesFiles)))
		for _, file := range valuesFiles {
			summary.WriteString(fmt.Sprintf("  %s\n", file.Path))
		}
	}

	return summary.String()
}

// ValidateFile checks if a file needs reindexing based on modification time
func (wi *WorkspaceIndex) ValidateFile(path string) bool {
	file := wi.FindFile(path)
	if file == nil {
		return false
	}

	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.ModTime().Equal(file.ModTime)
}

// GetFileContent reads the current content of a file
func GetFileContent(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// WriteFileContent writes content to a file, creating directories if needed
func WriteFileContent(path, content string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(content), 0644)
}

// Global workspace index instance
var WorkspaceIdx *WorkspaceIndex

// InitializeWorkspace initializes the global workspace index
func InitializeWorkspace() error {
	WorkspaceIdx = NewWorkspaceIndex()
	return WorkspaceIdx.ScanWorkspace(".")
}

// RefreshWorkspace rescans the workspace
func RefreshWorkspace() error {
	if WorkspaceIdx == nil {
		return InitializeWorkspace()
	}
	return WorkspaceIdx.ScanWorkspace(".")
}
