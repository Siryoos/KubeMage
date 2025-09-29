package engine

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// DiffLine represents a single line in a diff
type DiffLine struct {
	Type    string // "context", "add", "remove", "header"
	Content string
	LineNum int
}

// UnifiedDiff represents a unified diff
type UnifiedDiff struct {
	OriginalFile string
	ModifiedFile string
	Lines        []DiffLine
}

var (
	// Diff parsing regexes
	reDiffHeader  = regexp.MustCompile(`^diff --git a/(.*) b/(.*)`)
	reFileHeader  = regexp.MustCompile(`^(\+\+\+|---) (.+)`)
	reHunkHeader  = regexp.MustCompile(`^@@ -(\d+),?(\d*) \+(\d+),?(\d*) @@(.*)`)
	reAddLine     = regexp.MustCompile(`^\+(.*)`)
	reRemoveLine  = regexp.MustCompile(`^-(.*)`)
	reContextLine = regexp.MustCompile(`^ (.*)`)
)

// Diff styles for colored rendering
type DiffStyles struct {
	addStyle     lipgloss.Style
	removeStyle  lipgloss.Style
	contextStyle lipgloss.Style
	headerStyle  lipgloss.Style
	hunkStyle    lipgloss.Style
}

func NewDiffStyles() DiffStyles {
	return DiffStyles{
		addStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("82")),             // Green
		removeStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("196")),            // Red
		contextStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),            // Gray
		headerStyle:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")), // White
		hunkStyle:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33")),  // Blue
	}
}

// ParseUnifiedDiff parses a unified diff string
func ParseUnifiedDiff(diffText string) (*UnifiedDiff, error) {
	diff := &UnifiedDiff{}
	scanner := bufio.NewScanner(strings.NewReader(diffText))

	lineNum := 1
	for scanner.Scan() {
		line := scanner.Text()

		if matches := reDiffHeader.FindStringSubmatch(line); len(matches) > 2 {
			diff.OriginalFile = matches[1]
			diff.ModifiedFile = matches[2]
			diff.Lines = append(diff.Lines, DiffLine{
				Type:    "header",
				Content: line,
				LineNum: lineNum,
			})
		} else if reFileHeader.MatchString(line) {
			diff.Lines = append(diff.Lines, DiffLine{
				Type:    "header",
				Content: line,
				LineNum: lineNum,
			})
		} else if reHunkHeader.MatchString(line) {
			diff.Lines = append(diff.Lines, DiffLine{
				Type:    "hunk",
				Content: line,
				LineNum: lineNum,
			})
		} else if matches := reAddLine.FindStringSubmatch(line); len(matches) > 1 {
			diff.Lines = append(diff.Lines, DiffLine{
				Type:    "add",
				Content: matches[1],
				LineNum: lineNum,
			})
		} else if matches := reRemoveLine.FindStringSubmatch(line); len(matches) > 1 {
			diff.Lines = append(diff.Lines, DiffLine{
				Type:    "remove",
				Content: matches[1],
				LineNum: lineNum,
			})
		} else if matches := reContextLine.FindStringSubmatch(line); len(matches) > 1 {
			diff.Lines = append(diff.Lines, DiffLine{
				Type:    "context",
				Content: matches[1],
				LineNum: lineNum,
			})
		} else {
			// Unrecognized line, treat as context
			diff.Lines = append(diff.Lines, DiffLine{
				Type:    "context",
				Content: line,
				LineNum: lineNum,
			})
		}
		lineNum++
	}

	return diff, scanner.Err()
}

// RenderColoredDiff renders a diff with colors for terminal display
func (d *UnifiedDiff) RenderColoredDiff() string {
	styles := NewDiffStyles()
	var output strings.Builder

	output.WriteString(styles.headerStyle.Render(fmt.Sprintf("📝 Diff: %s → %s", d.OriginalFile, d.ModifiedFile)) + "\n")
	output.WriteString(strings.Repeat("─", 60) + "\n")

	for _, line := range d.Lines {
		var rendered string
		switch line.Type {
		case "add":
			rendered = styles.addStyle.Render("+ " + line.Content)
		case "remove":
			rendered = styles.removeStyle.Render("- " + line.Content)
		case "context":
			rendered = styles.contextStyle.Render("  " + line.Content)
		case "header":
			rendered = styles.headerStyle.Render(line.Content)
		case "hunk":
			rendered = styles.hunkStyle.Render(line.Content)
		default:
			rendered = line.Content
		}
		output.WriteString(rendered + "\n")
	}

	return output.String()
}

// GetDiffStats returns statistics about the diff
func (d *UnifiedDiff) GetDiffStats() (int, int) {
	adds := 0
	removes := 0

	for _, line := range d.Lines {
		switch line.Type {
		case "add":
			adds++
		case "remove":
			removes++
		}
	}

	return adds, removes
}

// GenerateUnifiedDiff creates a unified diff between two text contents
func GenerateUnifiedDiff(original, modified, filename string) (string, error) {
	// Simple unified diff generation
	// In a production environment, you might want to use a more sophisticated diff library

	originalLines := strings.Split(original, "\n")
	modifiedLines := strings.Split(modified, "\n")

	var diff strings.Builder

	// Write diff header
	diff.WriteString(fmt.Sprintf("diff --git a/%s b/%s\n", filename, filename))
	diff.WriteString(fmt.Sprintf("--- a/%s\n", filename))
	diff.WriteString(fmt.Sprintf("+++ b/%s\n", filename))

	// Simple line-by-line comparison
	maxLines := len(originalLines)
	if len(modifiedLines) > maxLines {
		maxLines = len(modifiedLines)
	}

	// Find changed sections
	changeStart := -1
	changes := []diffSection{}

	for i := 0; i < maxLines; i++ {
		origLine := ""
		modLine := ""

		if i < len(originalLines) {
			origLine = originalLines[i]
		}
		if i < len(modifiedLines) {
			modLine = modifiedLines[i]
		}

		if origLine != modLine {
			if changeStart == -1 {
				changeStart = i
			}
		} else {
			if changeStart != -1 {
				// End of change section
				changes = append(changes, diffSection{
					start: changeStart,
					end:   i,
				})
				changeStart = -1
			}
		}
	}

	// Handle case where file ends with changes
	if changeStart != -1 {
		changes = append(changes, diffSection{
			start: changeStart,
			end:   maxLines,
		})
	}

	// Generate hunks for each change section
	for _, section := range changes {
		contextBefore := 3
		contextAfter := 3

		hunkStart := section.start - contextBefore
		if hunkStart < 0 {
			hunkStart = 0
		}

		hunkEnd := section.end + contextAfter
		if hunkEnd > maxLines {
			hunkEnd = maxLines
		}

		// Calculate hunk header
		origStart := hunkStart + 1
		origLines := 0
		newStart := hunkStart + 1
		newLines := 0

		for i := hunkStart; i < hunkEnd; i++ {
			if i < len(originalLines) {
				origLines++
			}
			if i < len(modifiedLines) {
				newLines++
			}
		}

		diff.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n", origStart, origLines, newStart, newLines))

		// Write hunk content
		for i := hunkStart; i < hunkEnd; i++ {
			origLine := ""
			modLine := ""

			if i < len(originalLines) {
				origLine = originalLines[i]
			}
			if i < len(modifiedLines) {
				modLine = modifiedLines[i]
			}

			if i >= section.start && i < section.end {
				// This is in the changed section
				if i < len(originalLines) && origLine != "" {
					diff.WriteString("-" + origLine + "\n")
				}
				if i < len(modifiedLines) && modLine != "" {
					diff.WriteString("+" + modLine + "\n")
				}
			} else {
				// Context line
				if i < len(originalLines) {
					diff.WriteString(" " + origLine + "\n")
				} else if i < len(modifiedLines) {
					diff.WriteString(" " + modLine + "\n")
				}
			}
		}
	}

	return diff.String(), nil
}

type diffSection struct {
	start int
	end   int
}

// ApplyUnifiedDiff applies a unified diff to original content
func ApplyUnifiedDiff(original, diffText string) (string, error) {
	diff, err := ParseUnifiedDiff(diffText)
	if err != nil {
		return "", err
	}

	originalLines := strings.Split(original, "\n")
	var result []string

	// Simple application - this is a basic implementation
	// In production, you'd want a more robust patch application algorithm

	originalIndex := 0

	for _, line := range diff.Lines {
		switch line.Type {
		case "context":
			if originalIndex < len(originalLines) {
				result = append(result, originalLines[originalIndex])
				originalIndex++
			}
		case "remove":
			// Skip the original line
			if originalIndex < len(originalLines) {
				originalIndex++
			}
		case "add":
			// Add the new line
			result = append(result, line.Content)
		}
	}

	// Add any remaining original lines
	for originalIndex < len(originalLines) {
		result = append(result, originalLines[originalIndex])
		originalIndex++
	}

	return strings.Join(result, "\n"), nil
}

// ValidateDiff checks if a diff is valid and safe to apply
func ValidateDiff(diffText string) error {
	diff, err := ParseUnifiedDiff(diffText)
	if err != nil {
		return fmt.Errorf("invalid diff format: %v", err)
	}

	if len(diff.Lines) == 0 {
		return fmt.Errorf("empty diff")
	}

	// Basic validation - ensure we have both additions and removals or context
	hasContent := false
	for _, line := range diff.Lines {
		if line.Type == "add" || line.Type == "remove" || line.Type == "context" {
			hasContent = true
			break
		}
	}

	if !hasContent {
		return fmt.Errorf("diff contains no content changes")
	}

	return nil
}

// CreateBackup creates a backup of a file before applying changes and returns the backup path
func CreateBackup(filepath string) (string, error) {
	content, err := GetFileContent(filepath)
	if err != nil {
		return "", err
	}

	backupPath := fmt.Sprintf("%s.backup.%d", filepath, time.Now().Unix())
	if err := WriteFileContent(backupPath, content); err != nil {
		return "", err
	}
	return backupPath, nil
}
