package ui

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LogLevel represents different log levels
type LogLevel int

const (
	LogLevelAll LogLevel = iota
	LogLevelError
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
)

// LogLine represents a single log line with metadata
type LogLine struct {
	Timestamp   time.Time
	Level       LogLevel
	Content     string
	Source      string
	Highlighted bool
}

// LogFilter holds filtering criteria
type LogFilter struct {
	Level      LogLevel
	SearchTerm string
	Source     string
	Since      time.Time
}

// LogViewer manages log display with filtering and search
type LogViewer struct {
	theme           *Theme
	lines           []LogLine
	filteredLines   []LogLine
	filter          LogFilter
	viewOffset      int
	viewHeight      int
	viewWidth       int
	defaultLimit    int
	expanded        bool
	searchActive    bool
	searchQuery     string
	focused         bool
	truncatedCount  int
	totalLines      int
}

// Log level patterns for parsing
var logLevelPatterns = map[LogLevel]*regexp.Regexp{
	LogLevelError: regexp.MustCompile(`(?i)\b(error|err|fail|fatal|exception|panic)\b`),
	LogLevelWarn:  regexp.MustCompile(`(?i)\b(warn|warning|deprecated)\b`),
	LogLevelInfo:  regexp.MustCompile(`(?i)\b(info|information)\b`),
	LogLevelDebug: regexp.MustCompile(`(?i)\b(debug|trace|verbose)\b`),
}

// Timestamp patterns for parsing
var timestampPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`),                      // ISO format
	regexp.MustCompile(`\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}`),                      // Standard format
	regexp.MustCompile(`\w{3} \d{2} \d{2}:\d{2}:\d{2}`),                            // Syslog format
	regexp.MustCompile(`\d{2}:\d{2}:\d{2}`),                                         // Time only
}

// NewLogViewer creates a new log viewer
func NewLogViewer(theme *Theme) *LogViewer {
	return &LogViewer{
		theme:        theme,
		lines:        []LogLine{},
		filteredLines: []LogLine{},
		filter: LogFilter{
			Level: LogLevelAll,
		},
		viewOffset:   0,
		viewHeight:   20,
		viewWidth:    80,
		defaultLimit: 200,
		expanded:     false,
		searchActive: false,
		focused:      false,
	}
}

// SetContent sets the log content
func (lv *LogViewer) SetContent(content string) {
	lv.parseContent(content)
	lv.applyFilters()
}

// AppendContent appends new content to the log
func (lv *LogViewer) AppendContent(content string) {
	newLines := lv.parseLines(content)
	lv.lines = append(lv.lines, newLines...)
	lv.totalLines = len(lv.lines)
	lv.applyFilters()
}

// parseContent parses log content into structured lines
func (lv *LogViewer) parseContent(content string) {
	lv.lines = lv.parseLines(content)
	lv.totalLines = len(lv.lines)
}

// parseLines parses text content into LogLines
func (lv *LogViewer) parseLines(content string) []LogLine {
	lines := strings.Split(content, "\n")
	var logLines []LogLine

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		logLine := LogLine{
			Content: line,
			Level:   lv.detectLogLevel(line),
			Source:  lv.detectSource(line),
		}

		// Try to parse timestamp
		if ts := lv.parseTimestamp(line); !ts.IsZero() {
			logLine.Timestamp = ts
		} else {
			logLine.Timestamp = time.Now()
		}

		logLines = append(logLines, logLine)
	}

	return logLines
}

// detectLogLevel detects the log level from line content
func (lv *LogViewer) detectLogLevel(line string) LogLevel {
	for level, pattern := range logLevelPatterns {
		if pattern.MatchString(line) {
			return level
		}
	}
	return LogLevelInfo
}

// detectSource attempts to detect the log source
func (lv *LogViewer) detectSource(line string) string {
	// Simple heuristics for common sources
	if strings.Contains(line, "kubectl") {
		return "kubectl"
	}
	if strings.Contains(line, "kubelet") {
		return "kubelet"
	}
	if strings.Contains(line, "helm") {
		return "helm"
	}
	return "unknown"
}

// parseTimestamp attempts to parse timestamp from line
func (lv *LogViewer) parseTimestamp(line string) time.Time {
	for _, pattern := range timestampPatterns {
		if match := pattern.FindString(line); match != "" {
			// Try common time formats
			formats := []string{
				time.RFC3339,
				"2006-01-02T15:04:05",
				"2006/01/02 15:04:05",
				"Jan 02 15:04:05",
				"15:04:05",
			}

			for _, format := range formats {
				if t, err := time.Parse(format, match); err == nil {
					// For time-only formats, use today's date
					if format == "15:04:05" {
						now := time.Now()
						t = time.Date(now.Year(), now.Month(), now.Day(),
							t.Hour(), t.Minute(), t.Second(), 0, time.Local)
					}
					return t
				}
			}
		}
	}
	return time.Time{}
}

// SetFilter updates the log filter
func (lv *LogViewer) SetFilter(filter LogFilter) {
	lv.filter = filter
	lv.applyFilters()
}

// SetLevelFilter sets the log level filter
func (lv *LogViewer) SetLevelFilter(level LogLevel) {
	lv.filter.Level = level
	lv.applyFilters()
}

// SetSearchTerm sets the search term
func (lv *LogViewer) SetSearchTerm(term string) {
	lv.filter.SearchTerm = term
	lv.applyFilters()
}

// ToggleErrorFilter toggles error-only filtering
func (lv *LogViewer) ToggleErrorFilter() {
	if lv.filter.Level == LogLevelError {
		lv.filter.Level = LogLevelAll
	} else {
		lv.filter.Level = LogLevelError
	}
	lv.applyFilters()
}

// ToggleWarnFilter toggles warning-only filtering
func (lv *LogViewer) ToggleWarnFilter() {
	if lv.filter.Level == LogLevelWarn {
		lv.filter.Level = LogLevelAll
	} else {
		lv.filter.Level = LogLevelWarn
	}
	lv.applyFilters()
}

// ToggleExpanded toggles between truncated and full view
func (lv *LogViewer) ToggleExpanded() {
	lv.expanded = !lv.expanded
	lv.applyFilters()
}

// applyFilters applies current filters to generate filtered view
func (lv *LogViewer) applyFilters() {
	lv.filteredLines = []LogLine{}

	for _, line := range lv.lines {
		if lv.passesFilter(line) {
			// Highlight search terms
			if lv.filter.SearchTerm != "" && strings.Contains(
				strings.ToLower(line.Content),
				strings.ToLower(lv.filter.SearchTerm)) {
				line.Highlighted = true
			}
			lv.filteredLines = append(lv.filteredLines, line)
		}
	}

	// Apply truncation if not expanded
	if !lv.expanded && len(lv.filteredLines) > lv.defaultLimit {
		lv.truncatedCount = len(lv.filteredLines) - lv.defaultLimit
		lv.filteredLines = lv.filteredLines[len(lv.filteredLines)-lv.defaultLimit:]
	} else {
		lv.truncatedCount = 0
	}

	// Reset view offset if needed
	if lv.viewOffset >= len(lv.filteredLines) {
		lv.viewOffset = 0
	}
}

// passesFilter checks if a line passes current filters
func (lv *LogViewer) passesFilter(line LogLine) bool {
	// Level filter
	if lv.filter.Level != LogLevelAll && line.Level != lv.filter.Level {
		return false
	}

	// Search term filter
	if lv.filter.SearchTerm != "" {
		if !strings.Contains(
			strings.ToLower(line.Content),
			strings.ToLower(lv.filter.SearchTerm)) {
			return false
		}
	}

	// Source filter
	if lv.filter.Source != "" && line.Source != lv.filter.Source {
		return false
	}

	// Time filter
	if !lv.filter.Since.IsZero() && line.Timestamp.Before(lv.filter.Since) {
		return false
	}

	return true
}

// ScrollUp scrolls the view up
func (lv *LogViewer) ScrollUp() {
	if lv.viewOffset > 0 {
		lv.viewOffset--
	}
}

// ScrollDown scrolls the view down
func (lv *LogViewer) ScrollDown() {
	maxOffset := len(lv.filteredLines) - lv.viewHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	if lv.viewOffset < maxOffset {
		lv.viewOffset++
	}
}

// ScrollToBottom scrolls to the bottom
func (lv *LogViewer) ScrollToBottom() {
	maxOffset := len(lv.filteredLines) - lv.viewHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	lv.viewOffset = maxOffset
}

// SetSize updates the view dimensions
func (lv *LogViewer) SetSize(width, height int) {
	lv.viewWidth = width
	lv.viewHeight = height
}

// SetFocused sets the focus state
func (lv *LogViewer) SetFocused(focused bool) {
	lv.focused = focused
}

// StartSearch activates search mode
func (lv *LogViewer) StartSearch() {
	lv.searchActive = true
	lv.searchQuery = ""
}

// EndSearch deactivates search mode
func (lv *LogViewer) EndSearch() {
	lv.searchActive = false
	lv.SetSearchTerm(lv.searchQuery)
}

// UpdateSearchQuery updates the search query
func (lv *LogViewer) UpdateSearchQuery(query string) {
	lv.searchQuery = query
}

// Render renders the log view
func (lv *LogViewer) Render() string {
	var sections []string

	// Header with filter info
	header := lv.renderHeader()
	if header != "" {
		sections = append(sections, header)
	}

	// Truncation notice
	if lv.truncatedCount > 0 {
		truncationNotice := lv.renderTruncationNotice()
		sections = append(sections, truncationNotice)
	}

	// Log lines
	logContent := lv.renderLogLines()
	sections = append(sections, logContent)

	// Search bar (if active)
	if lv.searchActive {
		searchBar := lv.renderSearchBar()
		sections = append(sections, searchBar)
	}

	// Footer with shortcuts
	if lv.focused {
		footer := lv.renderFooter()
		sections = append(sections, footer)
	}

	content := strings.Join(sections, "\n")

	// Apply styling
	style := lv.theme.PaneStyle()
	if lv.focused {
		style = lv.theme.PaneFocusedStyle()
	}

	return style.Width(lv.viewWidth).Height(lv.viewHeight).Render(content)
}

// renderHeader renders the log viewer header
func (lv *LogViewer) renderHeader() string {
	headerStyle := lv.theme.HighlightStyle("keyword").Bold(true)

	var parts []string

	// Total lines info
	parts = append(parts, fmt.Sprintf("Lines: %d", len(lv.filteredLines)))

	// Filter info
	if lv.filter.Level != LogLevelAll {
		level := lv.getLevelName(lv.filter.Level)
		parts = append(parts, fmt.Sprintf("Filter: %s", level))
	}

	// Search info
	if lv.filter.SearchTerm != "" {
		parts = append(parts, fmt.Sprintf("Search: \"%s\"", lv.filter.SearchTerm))
	}

	return headerStyle.Render(strings.Join(parts, " • "))
}

// renderTruncationNotice renders the truncation notice
func (lv *LogViewer) renderTruncationNotice() string {
	noticeStyle := lv.theme.HighlightStyle("warning").
		Background(lv.theme.Colors.Warning).
		Foreground(lv.theme.Colors.Primary).
		Padding(0, 1)

	notice := fmt.Sprintf("(...%d lines truncated... press ] to expand)", lv.truncatedCount)
	return noticeStyle.Render(notice)
}

// renderLogLines renders the visible log lines
func (lv *LogViewer) renderLogLines() string {
	if len(lv.filteredLines) == 0 {
		emptyStyle := lv.theme.HighlightStyle("comment").Italic(true)
		return emptyStyle.Render("No log entries match the current filter")
	}

	// Calculate visible range
	start := lv.viewOffset
	end := start + lv.viewHeight - 3 // Reserve space for header/footer
	if end > len(lv.filteredLines) {
		end = len(lv.filteredLines)
	}

	var lines []string
	for i := start; i < end; i++ {
		line := lv.filteredLines[i]
		renderedLine := lv.renderLogLine(line)
		lines = append(lines, renderedLine)
	}

	return strings.Join(lines, "\n")
}

// renderLogLine renders a single log line
func (lv *LogViewer) renderLogLine(line LogLine) string {
	var parts []string

	// Timestamp
	if !line.Timestamp.IsZero() {
		timestampStyle := lv.theme.HighlightStyle("comment")
		timestamp := line.Timestamp.Format("15:04:05")
		parts = append(parts, timestampStyle.Render(timestamp))
	}

	// Level indicator
	levelStyle := lv.getLevelStyle(line.Level)
	levelIndicator := lv.getLevelIndicator(line.Level)
	parts = append(parts, levelStyle.Render(levelIndicator))

	// Content
	content := line.Content
	contentStyle := lv.theme.HighlightStyle("string")

	// Apply level-specific styling
	switch line.Level {
	case LogLevelError:
		contentStyle = lv.theme.HighlightStyle("string").Foreground(lv.theme.Colors.Error)
	case LogLevelWarn:
		contentStyle = lv.theme.HighlightStyle("string").Foreground(lv.theme.Colors.Warning)
	}

	// Highlight search terms
	if line.Highlighted && lv.filter.SearchTerm != "" {
		content = lv.highlightSearchTerm(content, lv.filter.SearchTerm)
	}

	parts = append(parts, contentStyle.Render(content))

	return strings.Join(parts, " ")
}

// highlightSearchTerm highlights search terms in content
func (lv *LogViewer) highlightSearchTerm(content, term string) string {
	if term == "" {
		return content
	}

	highlightStyle := lv.theme.HighlightStyle("keyword").
		Background(lv.theme.Colors.Accent).
		Foreground(lv.theme.Colors.Primary)

	// Case-insensitive replacement
	re := regexp.MustCompile("(?i)" + regexp.QuoteMeta(term))
	return re.ReplaceAllStringFunc(content, func(match string) string {
		return highlightStyle.Render(match)
	})
}

// renderSearchBar renders the search input bar
func (lv *LogViewer) renderSearchBar() string {
	searchStyle := lv.theme.InputStyle()
	prompt := lv.theme.HighlightStyle("keyword").Render("Search: ")
	query := lv.searchQuery + "█" // Add cursor

	return searchStyle.Render(prompt + query)
}

// renderFooter renders keyboard shortcuts
func (lv *LogViewer) renderFooter() string {
	footerStyle := lv.theme.HighlightStyle("comment")
	shortcuts := "e: errors • w: warnings • /: search • ]: expand • ↑↓: scroll"
	return footerStyle.Render(shortcuts)
}

// Helper methods for styling
func (lv *LogViewer) getLevelStyle(level LogLevel) lipgloss.Style {
	switch level {
	case LogLevelError:
		return lv.theme.HighlightStyle("string").Foreground(lv.theme.Colors.Error)
	case LogLevelWarn:
		return lv.theme.HighlightStyle("string").Foreground(lv.theme.Colors.Warning)
	case LogLevelInfo:
		return lv.theme.HighlightStyle("string").Foreground(lv.theme.Colors.Info)
	case LogLevelDebug:
		return lv.theme.HighlightStyle("string").Foreground(lv.theme.Colors.Muted)
	default:
		return lv.theme.HighlightStyle("string")
	}
}

func (lv *LogViewer) getLevelIndicator(level LogLevel) string {
	switch level {
	case LogLevelError:
		return "ERR"
	case LogLevelWarn:
		return "WRN"
	case LogLevelInfo:
		return "INF"
	case LogLevelDebug:
		return "DBG"
	default:
		return "LOG"
	}
}

func (lv *LogViewer) getLevelName(level LogLevel) string {
	switch level {
	case LogLevelError:
		return "Error"
	case LogLevelWarn:
		return "Warning"
	case LogLevelInfo:
		return "Info"
	case LogLevelDebug:
		return "Debug"
	default:
		return "All"
	}
}

// Init initializes the LogViewer component
func (lv *LogViewer) Init() tea.Cmd {
	return nil
}

// View renders the LogViewer component
func (lv *LogViewer) View() string {
	return lv.Render()
}

// Update handles Bubble Tea messages
func (lv *LogViewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !lv.focused {
		return lv, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if lv.searchActive {
			return lv.handleSearchInput(msg)
		}

		switch msg.String() {
		case "e":
			lv.ToggleErrorFilter()
		case "w":
			lv.ToggleWarnFilter()
		case "/":
			lv.StartSearch()
		case "]":
			lv.ToggleExpanded()
		case "up", "k":
			lv.ScrollUp()
		case "down", "j":
			lv.ScrollDown()
		case "g":
			lv.viewOffset = 0
		case "G":
			lv.ScrollToBottom()
		}
	}

	return lv, nil
}

// handleSearchInput handles input during search mode
func (lv *LogViewer) handleSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		lv.EndSearch()
	case tea.KeyEsc:
		lv.searchActive = false
		lv.searchQuery = ""
	case tea.KeyBackspace:
		if len(lv.searchQuery) > 0 {
			lv.searchQuery = lv.searchQuery[:len(lv.searchQuery)-1]
		}
	default:
		if msg.Type == tea.KeyRunes {
			lv.searchQuery += string(msg.Runes)
		}
	}

	return lv, nil
}

// GetFilteredLineCount returns the number of filtered lines
func (lv *LogViewer) GetFilteredLineCount() int {
	return len(lv.filteredLines)
}

// GetTotalLineCount returns the total number of lines
func (lv *LogViewer) GetTotalLineCount() int {
	return lv.totalLines
}

// IsExpanded returns whether the view is expanded
func (lv *LogViewer) IsExpanded() bool {
	return lv.expanded
}

// GetCurrentFilter returns the current filter
func (lv *LogViewer) GetCurrentFilter() LogFilter {
	return lv.filter
}