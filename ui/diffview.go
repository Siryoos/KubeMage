package ui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DiffLine represents a single line in a diff
type DiffLine struct {
	Type      string // "added", "removed", "context", "header"
	Content   string
	OldLineNo int
	NewLineNo int
	HunkIndex int
}

// DiffHunk represents a contiguous block of diff changes
type DiffHunk struct {
	OldStart   int
	OldCount   int
	NewStart   int
	NewCount   int
	Lines      []DiffLine
	StartIndex int // Index of first line in the overall diff
	EndIndex   int // Index of last line in the overall diff
}

// DiffView manages the display and navigation of unified diffs
type DiffView struct {
	theme        *Theme
	rawDiff      string
	lines        []DiffLine
	hunks        []DiffHunk
	currentHunk  int
	viewOffset   int
	viewHeight   int
	viewWidth    int
	showLineNums bool
	focused      bool
	staged       map[int]bool // For future hunk staging feature
}

// NewDiffView creates a new diff viewer
func NewDiffView(diffContent string, theme *Theme) *DiffView {
	dv := &DiffView{
		theme:        theme,
		rawDiff:      diffContent,
		currentHunk:  0,
		viewOffset:   0,
		viewHeight:   20,
		viewWidth:    80,
		showLineNums: true,
		focused:      false,
		staged:       make(map[int]bool),
	}

	dv.parseDiff()
	return dv
}

// parseDiff parses the unified diff format
func (dv *DiffView) parseDiff() {
	lines := strings.Split(dv.rawDiff, "\n")
	dv.lines = []DiffLine{}
	dv.hunks = []DiffHunk{}

	var currentHunk *DiffHunk
	hunkIndex := -1
	lineIndex := 0

	for _, line := range lines {
		if strings.HasPrefix(line, "@@") {
			// New hunk header
			if currentHunk != nil {
				currentHunk.EndIndex = lineIndex - 1
				dv.hunks = append(dv.hunks, *currentHunk)
			}

			hunkIndex++
			currentHunk = dv.parseHunkHeader(line, hunkIndex, lineIndex)
			if currentHunk == nil {
				continue
			}

			// Add hunk header line
			dv.lines = append(dv.lines, DiffLine{
				Type:      "header",
				Content:   line,
				OldLineNo: -1,
				NewLineNo: -1,
				HunkIndex: hunkIndex,
			})
			lineIndex++

		} else if currentHunk != nil {
			// Parse diff line
			diffLine := dv.parseDiffLine(line, currentHunk, hunkIndex)
			dv.lines = append(dv.lines, diffLine)
			currentHunk.Lines = append(currentHunk.Lines, diffLine)
			lineIndex++

		} else if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") {
			// File header lines
			dv.lines = append(dv.lines, DiffLine{
				Type:      "header",
				Content:   line,
				OldLineNo: -1,
				NewLineNo: -1,
				HunkIndex: -1,
			})
			lineIndex++
		}
	}

	// Finalize last hunk
	if currentHunk != nil {
		currentHunk.EndIndex = lineIndex - 1
		dv.hunks = append(dv.hunks, *currentHunk)
	}
}

// parseHunkHeader parses a hunk header line (@@...)
func (dv *DiffView) parseHunkHeader(line string, hunkIndex, lineIndex int) *DiffHunk {
	// Format: @@ -oldStart,oldCount +newStart,newCount @@
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return nil
	}

	oldPart := strings.TrimPrefix(parts[1], "-")
	newPart := strings.TrimPrefix(parts[2], "+")

	oldStart, oldCount := dv.parseRange(oldPart)
	newStart, newCount := dv.parseRange(newPart)

	return &DiffHunk{
		OldStart:   oldStart,
		OldCount:   oldCount,
		NewStart:   newStart,
		NewCount:   newCount,
		Lines:      []DiffLine{},
		StartIndex: lineIndex,
		EndIndex:   lineIndex, // Will be updated when hunk ends
	}
}

// parseRange parses a range like "10,5" or just "10"
func (dv *DiffView) parseRange(rangeStr string) (start, count int) {
	parts := strings.Split(rangeStr, ",")
	if len(parts) == 2 {
		start, _ = strconv.Atoi(parts[0])
		count, _ = strconv.Atoi(parts[1])
	} else if len(parts) == 1 {
		start, _ = strconv.Atoi(parts[0])
		count = 1
	}
	return
}

// parseDiffLine parses a single diff line
func (dv *DiffView) parseDiffLine(line string, hunk *DiffHunk, hunkIndex int) DiffLine {
	if len(line) == 0 {
		return DiffLine{
			Type:      "context",
			Content:   "",
			OldLineNo: -1,
			NewLineNo: -1,
			HunkIndex: hunkIndex,
		}
	}

	prefix := line[0:1]
	content := ""
	if len(line) > 1 {
		content = line[1:]
	}

	var lineType string
	var oldLineNo, newLineNo int = -1, -1

	switch prefix {
	case "+":
		lineType = "added"
		// Calculate line numbers (simplified)
		newLineNo = hunk.NewStart + len(hunk.Lines)

	case "-":
		lineType = "removed"
		// Calculate line numbers (simplified)
		oldLineNo = hunk.OldStart + len(hunk.Lines)

	case " ":
		lineType = "context"
		// Both old and new line numbers increment
		oldLineNo = hunk.OldStart + len(hunk.Lines)
		newLineNo = hunk.NewStart + len(hunk.Lines)

	default:
		lineType = "context"
		content = line // Keep the original line if prefix is unrecognized
	}

	return DiffLine{
		Type:      lineType,
		Content:   content,
		OldLineNo: oldLineNo,
		NewLineNo: newLineNo,
		HunkIndex: hunkIndex,
	}
}

// NextHunk moves to the next hunk
func (dv *DiffView) NextHunk() {
	if dv.currentHunk < len(dv.hunks)-1 {
		dv.currentHunk++
		dv.scrollToCurrentHunk()
	}
}

// PrevHunk moves to the previous hunk
func (dv *DiffView) PrevHunk() {
	if dv.currentHunk > 0 {
		dv.currentHunk--
		dv.scrollToCurrentHunk()
	}
}

// scrollToCurrentHunk scrolls the view to show the current hunk
func (dv *DiffView) scrollToCurrentHunk() {
	if len(dv.hunks) == 0 {
		return
	}

	hunk := dv.hunks[dv.currentHunk]
	// Position the hunk at the top of the view with some context
	dv.viewOffset = hunk.StartIndex
	if dv.viewOffset > 2 {
		dv.viewOffset -= 2 // Show 2 lines of context above
	}
}

// ScrollUp scrolls the view up
func (dv *DiffView) ScrollUp() {
	if dv.viewOffset > 0 {
		dv.viewOffset--
	}
}

// ScrollDown scrolls the view down
func (dv *DiffView) ScrollDown() {
	maxOffset := len(dv.lines) - dv.viewHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	if dv.viewOffset < maxOffset {
		dv.viewOffset++
	}
}

// SetSize updates the view dimensions
func (dv *DiffView) SetSize(width, height int) {
	dv.viewWidth = width
	dv.viewHeight = height
}

// SetFocused sets the focus state
func (dv *DiffView) SetFocused(focused bool) {
	dv.focused = focused
}

// ToggleLineNumbers toggles line number display
func (dv *DiffView) ToggleLineNumbers() {
	dv.showLineNums = !dv.showLineNums
}

// ToggleHunkStaging toggles staging for the current hunk
func (dv *DiffView) ToggleHunkStaging() {
	if len(dv.hunks) > 0 {
		hunkIndex := dv.currentHunk
		dv.staged[hunkIndex] = !dv.staged[hunkIndex]
	}
}

// GetStats returns diff statistics
func (dv *DiffView) GetStats() (added, removed int) {
	for _, line := range dv.lines {
		switch line.Type {
		case "added":
			added++
		case "removed":
			removed++
		}
	}
	return
}

// Render renders the diff view
func (dv *DiffView) Render() string {
	if len(dv.lines) == 0 {
		emptyStyle := dv.theme.HighlightStyle("comment").Italic(true)
		return emptyStyle.Render("No diff content")
	}

	// Calculate visible lines
	startLine := dv.viewOffset
	endLine := startLine + dv.viewHeight
	if endLine > len(dv.lines) {
		endLine = len(dv.lines)
	}

	var renderedLines []string

	// Render header with navigation info
	header := dv.renderHeader()
	if header != "" {
		renderedLines = append(renderedLines, header)
	}

	// Render visible diff lines
	for i := startLine; i < endLine; i++ {
		line := dv.lines[i]
		renderedLine := dv.renderDiffLine(line, i)
		renderedLines = append(renderedLines, renderedLine)
	}

	// Add footer with shortcuts
	footer := dv.renderFooter()
	if footer != "" {
		renderedLines = append(renderedLines, footer)
	}

	content := strings.Join(renderedLines, "\n")

	// Apply overall styling
	style := dv.theme.PaneStyle()
	if dv.focused {
		style = dv.theme.PaneFocusedStyle()
	}

	return style.Width(dv.viewWidth).Height(dv.viewHeight).Render(content)
}

// renderHeader renders the diff header with navigation info
func (dv *DiffView) renderHeader() string {
	if len(dv.hunks) == 0 {
		return ""
	}

	headerStyle := dv.theme.HighlightStyle("keyword").Bold(true)
	hunkInfo := fmt.Sprintf("Hunk %d/%d", dv.currentHunk+1, len(dv.hunks))

	added, removed := dv.GetStats()
	statsInfo := fmt.Sprintf("(+%d/-%d)", added, removed)

	return headerStyle.Render(fmt.Sprintf("%s %s", hunkInfo, statsInfo))
}

// renderFooter renders keyboard shortcuts
func (dv *DiffView) renderFooter() string {
	if !dv.focused {
		return ""
	}

	footerStyle := dv.theme.HighlightStyle("comment")
	shortcuts := "n/p: next/prev hunk • ↑/↓: scroll • space: stage hunk • l: toggle line numbers"
	return footerStyle.Render(shortcuts)
}

// renderDiffLine renders a single diff line
func (dv *DiffView) renderDiffLine(line DiffLine, lineIndex int) string {
	var parts []string

	// Line numbers
	if dv.showLineNums {
		lineNumStyle := dv.theme.HighlightStyle("comment").Width(8)

		var lineNumText string
		switch line.Type {
		case "added":
			lineNumText = fmt.Sprintf("   +%-4d", line.NewLineNo)
		case "removed":
			lineNumText = fmt.Sprintf("   -%-4d", line.OldLineNo)
		case "context":
			if line.OldLineNo >= 0 && line.NewLineNo >= 0 {
				lineNumText = fmt.Sprintf("%4d:%4d", line.OldLineNo, line.NewLineNo)
			} else {
				lineNumText = "        "
			}
		default:
			lineNumText = "        "
		}

		parts = append(parts, lineNumStyle.Render(lineNumText))
	}

	// Diff prefix and content
	var contentStyle lipgloss.Style
	var prefix string

	switch line.Type {
	case "added":
		contentStyle = dv.theme.DiffStyle("added")
		prefix = "+"
	case "removed":
		contentStyle = dv.theme.DiffStyle("removed")
		prefix = "-"
	case "header":
		contentStyle = dv.theme.HighlightStyle("keyword").Bold(true)
		prefix = ""
	default:
		contentStyle = dv.theme.DiffStyle("context")
		prefix = " "
	}

	// Highlight current hunk
	if line.HunkIndex == dv.currentHunk && dv.focused {
		contentStyle = contentStyle.Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(dv.theme.Colors.Accent)
	}

	// Staging indicator
	stagingIndicator := " "
	if dv.staged[line.HunkIndex] {
		stagingIndicator = "●"
	}

	content := fmt.Sprintf("%s%s%s", stagingIndicator, prefix, line.Content)
	parts = append(parts, contentStyle.Render(content))

	return lipgloss.JoinHorizontal(lipgloss.Left, parts...)
}

// Init initializes the DiffView component
func (dv *DiffView) Init() tea.Cmd {
	return nil
}

// View renders the DiffView component
func (dv *DiffView) View() string {
	return dv.Render()
}

// Update handles Bubble Tea messages
func (dv *DiffView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !dv.focused {
		return dv, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "n":
			dv.NextHunk()
		case "p":
			dv.PrevHunk()
		case "up", "k":
			dv.ScrollUp()
		case "down", "j":
			dv.ScrollDown()
		case " ":
			dv.ToggleHunkStaging()
		case "l":
			dv.ToggleLineNumbers()
		case "g":
			dv.viewOffset = 0 // Go to top
		case "G":
			dv.viewOffset = len(dv.lines) - dv.viewHeight // Go to bottom
			if dv.viewOffset < 0 {
				dv.viewOffset = 0
			}
		}
	}

	return dv, nil
}

// GetStagedHunks returns indices of staged hunks
func (dv *DiffView) GetStagedHunks() []int {
	var staged []int
	for hunkIndex, isStaged := range dv.staged {
		if isStaged {
			staged = append(staged, hunkIndex)
		}
	}
	return staged
}

// GetHunkCount returns the total number of hunks
func (dv *DiffView) GetHunkCount() int {
	return len(dv.hunks)
}

// GetCurrentHunk returns the index of the current hunk
func (dv *DiffView) GetCurrentHunk() int {
	return dv.currentHunk
}

// RenderCompact renders a compact summary of the diff
func (dv *DiffView) RenderCompact() string {
	added, removed := dv.GetStats()
	compactStyle := dv.theme.HighlightStyle("comment")

	summary := fmt.Sprintf("Diff: %d hunks, +%d/-%d lines",
		len(dv.hunks), added, removed)

	if len(dv.hunks) > 0 {
		summary += fmt.Sprintf(" (viewing %d/%d)",
			dv.currentHunk+1, len(dv.hunks))
	}

	return compactStyle.Render(summary)
}
