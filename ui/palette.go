package ui

import (
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PaletteItem represents a single command in the palette
type PaletteItem struct {
	Command     string
	Description string
	Category    string
	Template    string // The template to insert into input
	Priority    int    // Higher priority items appear first
}

// PaletteMatch represents a matched item with score
type PaletteMatch struct {
	Item  PaletteItem
	Score int
}

// CommandPalette manages the command palette UI
type CommandPalette struct {
	theme       *Theme
	isOpen      bool
	query       string
	items       []PaletteItem
	matches     []PaletteMatch
	selected    int
	maxItems    int
	width       int
	height      int
}

// Default palette items
var defaultPaletteItems = []PaletteItem{
	// Diagnostics
	{
		Command:     "Diagnose Pod",
		Description: "Run comprehensive pod diagnostics",
		Category:    "Diagnostics",
		Template:    "/diag-pod ",
		Priority:    10,
	},

	// Generation
	{
		Command:     "Generate Deployment",
		Description: "Generate deployment manifest",
		Category:    "Generation",
		Template:    "/gen-deploy ",
		Priority:    9,
	},
	{
		Command:     "Generate Helm Chart",
		Description: "Generate Helm chart skeleton",
		Category:    "Generation",
		Template:    "/gen-helm ",
		Priority:    9,
	},

	// Editing
	{
		Command:     "Edit YAML",
		Description: "Generate diff for manifest file",
		Category:    "Editing",
		Template:    "/edit-yaml ",
		Priority:    8,
	},
	{
		Command:     "Edit Values",
		Description: "Generate diff for Helm values",
		Category:    "Editing",
		Template:    "/edit-values ",
		Priority:    8,
	},

	// Context & Environment
	{
		Command:     "Switch Namespace",
		Description: "Change active namespace",
		Category:    "Context",
		Template:    "/ns set ",
		Priority:    7,
	},
	{
		Command:     "Show Context",
		Description: "Display current context info",
		Category:    "Context",
		Template:    "/ctx",
		Priority:    6,
	},

	// Models
	{
		Command:     "List Models",
		Description: "Show available Ollama models",
		Category:    "Models",
		Template:    "/model list",
		Priority:    5,
	},
	{
		Command:     "Set Chat Model",
		Description: "Change chat model",
		Category:    "Models",
		Template:    "/model set chat ",
		Priority:    5,
	},
	{
		Command:     "Set Generation Model",
		Description: "Change generation/diff model",
		Category:    "Models",
		Template:    "/model set generation ",
		Priority:    5,
	},

	// Session Management
	{
		Command:     "Show Metrics",
		Description: "Display session metrics",
		Category:    "Session",
		Template:    "/metrics",
		Priority:    4,
	},
	{
		Command:     "Resolve Task",
		Description: "Mark current task as resolved",
		Category:    "Session",
		Template:    "/resolve ",
		Priority:    4,
	},
	{
		Command:     "Cancel Operation",
		Description: "Cancel pending diff/generation",
		Category:    "Session",
		Template:    "/cancel",
		Priority:    3,
	},

	// Help
	{
		Command:     "Help",
		Description: "Toggle help display",
		Category:    "Help",
		Template:    "/help",
		Priority:    2,
	},
}

// NewCommandPalette creates a new command palette
func NewCommandPalette(theme *Theme) *CommandPalette {
	palette := &CommandPalette{
		theme:    theme,
		isOpen:   false,
		query:    "",
		items:    defaultPaletteItems,
		matches:  []PaletteMatch{},
		selected: 0,
		maxItems: 8,
		width:    60,
		height:   12,
	}

	// Initialize with all items
	palette.updateMatches()
	return palette
}

// Open opens the command palette
func (cp *CommandPalette) Open() {
	cp.isOpen = true
	cp.query = ""
	cp.selected = 0
	cp.updateMatches()
}

// Close closes the command palette
func (cp *CommandPalette) Close() {
	cp.isOpen = false
	cp.query = ""
	cp.selected = 0
}

// IsOpen returns whether the palette is open
func (cp *CommandPalette) IsOpen() bool {
	return cp.isOpen
}

// SetQuery updates the search query and recalculates matches
func (cp *CommandPalette) SetQuery(query string) {
	cp.query = query
	cp.selected = 0
	cp.updateMatches()
}

// GetQuery returns the current search query
func (cp *CommandPalette) GetQuery() string {
	return cp.query
}

// MoveUp moves selection up
func (cp *CommandPalette) MoveUp() {
	if cp.selected > 0 {
		cp.selected--
	}
}

// MoveDown moves selection down
func (cp *CommandPalette) MoveDown() {
	if cp.selected < len(cp.matches)-1 {
		cp.selected++
	}
}

// GetSelectedItem returns the currently selected item
func (cp *CommandPalette) GetSelectedItem() *PaletteItem {
	if cp.selected >= 0 && cp.selected < len(cp.matches) {
		return &cp.matches[cp.selected].Item
	}
	return nil
}

// updateMatches calculates fuzzy matches for the current query
func (cp *CommandPalette) updateMatches() {
	cp.matches = []PaletteMatch{}

	if cp.query == "" {
		// Show all items sorted by priority
		for _, item := range cp.items {
			cp.matches = append(cp.matches, PaletteMatch{
				Item:  item,
				Score: item.Priority * 10,
			})
		}
	} else {
		// Calculate fuzzy match scores
		query := strings.ToLower(cp.query)
		for _, item := range cp.items {
			score := cp.calculateMatchScore(item, query)
			if score > 0 {
				cp.matches = append(cp.matches, PaletteMatch{
					Item:  item,
					Score: score,
				})
			}
		}
	}

	// Sort by score (descending)
	sort.Slice(cp.matches, func(i, j int) bool {
		return cp.matches[i].Score > cp.matches[j].Score
	})

	// Limit to maxItems
	if len(cp.matches) > cp.maxItems {
		cp.matches = cp.matches[:cp.maxItems]
	}

	// Ensure selected index is valid
	if cp.selected >= len(cp.matches) {
		cp.selected = len(cp.matches) - 1
	}
	if cp.selected < 0 && len(cp.matches) > 0 {
		cp.selected = 0
	}
}

// calculateMatchScore calculates a fuzzy match score for an item
func (cp *CommandPalette) calculateMatchScore(item PaletteItem, query string) int {
	command := strings.ToLower(item.Command)
	description := strings.ToLower(item.Description)
	category := strings.ToLower(item.Category)

	score := 0
	basePriority := item.Priority

	// Exact matches get highest score
	if strings.Contains(command, query) {
		score += 100 + basePriority
	}
	if strings.Contains(description, query) {
		score += 80 + basePriority
	}
	if strings.Contains(category, query) {
		score += 60 + basePriority
	}

	// Partial matches for individual words
	queryWords := strings.Fields(query)
	for _, word := range queryWords {
		if len(word) < 2 {
			continue
		}

		if strings.Contains(command, word) {
			score += 50 + basePriority
		}
		if strings.Contains(description, word) {
			score += 30 + basePriority
		}
		if strings.Contains(category, word) {
			score += 20 + basePriority
		}
	}

	// Character-by-character fuzzy matching
	if cp.fuzzyMatch(command, query) {
		score += 40 + basePriority
	}
	if cp.fuzzyMatch(description, query) {
		score += 20 + basePriority
	}

	return score
}

// fuzzyMatch performs character-by-character fuzzy matching
func (cp *CommandPalette) fuzzyMatch(text, pattern string) bool {
	if len(pattern) == 0 {
		return true
	}
	if len(text) == 0 {
		return false
	}

	patternIdx := 0
	for _, char := range text {
		if patternIdx < len(pattern) && char == rune(pattern[patternIdx]) {
			patternIdx++
		}
	}

	return patternIdx == len(pattern)
}

// Render renders the command palette
func (cp *CommandPalette) Render() string {
	if !cp.isOpen {
		return ""
	}

	theme := CurrentTheme()

	// Header
	headerStyle := theme.HeaderStyle().Width(cp.width)
	header := headerStyle.Render("Command Palette")

	// Search input
	inputStyle := theme.InputStyle().Width(cp.width - 4)
	searchPrompt := theme.HighlightStyle("keyword").Render("❯ ")
	searchInput := inputStyle.Render(searchPrompt + cp.query)

	// Results
	var resultLines []string
	if len(cp.matches) == 0 {
		noResultsStyle := theme.HighlightStyle("comment").Italic(true)
		resultLines = append(resultLines, noResultsStyle.Render("No matching commands"))
	} else {
		for i, match := range cp.matches {
			resultLines = append(resultLines, cp.renderItem(match.Item, i == cp.selected, theme))
		}
	}

	// Instructions
	instructionStyle := theme.HighlightStyle("comment")
	instructions := instructionStyle.Render("↑/↓ navigate • Enter select • Esc close")

	// Combine all parts
	content := strings.Join([]string{
		header,
		"",
		searchInput,
		"",
		strings.Join(resultLines, "\n"),
		"",
		instructions,
	}, "\n")

	// Wrap in modal style
	modalStyle := theme.ModalStyle().Width(cp.width).Align(lipgloss.Center)
	return modalStyle.Render(content)
}

// renderItem renders a single palette item
func (cp *CommandPalette) renderItem(item PaletteItem, selected bool, theme *Theme) string {
	// Category badge
	categoryStyle := theme.HighlightStyle("keyword").
		Background(theme.Colors.Surface).
		Padding(0, 1)
	categoryBadge := categoryStyle.Render(item.Category)

	// Command name
	commandStyle := theme.HighlightStyle("string")
	if selected {
		commandStyle = commandStyle.Bold(true).Foreground(theme.Colors.Accent)
	}
	commandText := commandStyle.Render(item.Command)

	// Description
	descStyle := theme.HighlightStyle("comment")
	descText := descStyle.Render(item.Description)

	// Selection indicator
	indicator := " "
	if selected {
		indicator = theme.HighlightStyle("keyword").Render("❯")
	}

	// Layout
	line := lipgloss.JoinHorizontal(
		lipgloss.Left,
		indicator+" ",
		categoryBadge,
		" ",
		commandText,
		" - ",
		descText,
	)

	if selected {
		selectionStyle := theme.PaneFocusedStyle().
			Background(theme.Colors.Surface).
			Width(cp.width - 4)
		return selectionStyle.Render(line)
	}

	return line
}

// Update handles Bubble Tea messages for the palette
func (cp *CommandPalette) Update(msg tea.Msg) tea.Cmd {
	if !cp.isOpen {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			cp.Close()
			return nil

		case tea.KeyUp:
			cp.MoveUp()
			return nil

		case tea.KeyDown:
			cp.MoveDown()
			return nil

		case tea.KeyEnter:
			// Selection will be handled by the parent component
			return nil

		case tea.KeyBackspace:
			if len(cp.query) > 0 {
				cp.SetQuery(cp.query[:len(cp.query)-1])
			}
			return nil

		default:
			// Add character to query
			if msg.Type == tea.KeyRunes {
				cp.SetQuery(cp.query + string(msg.Runes))
			}
			return nil
		}
	}

	return nil
}

// AddCustomItem adds a custom item to the palette
func (cp *CommandPalette) AddCustomItem(item PaletteItem) {
	cp.items = append(cp.items, item)
	cp.updateMatches()
}

// SetSize updates the palette dimensions
func (cp *CommandPalette) SetSize(width, height int) {
	if width > 30 {
		cp.width = width
	}
	if height > 8 {
		cp.height = height
	}
}