package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ChipAction represents an action that can be performed from a chip
type ChipAction struct {
	ID       string
	Label    string
	Command  string
	Tooltip  string
	Category string
	Icon     string
}

// QuickActionChips manages a row of actionable chips
type QuickActionChips struct {
	theme      *Theme
	chips      []ChipAction
	focused    int
	messageID  string
	enabled    bool
	maxVisible int
}

// ChipSet contains predefined chip configurations for different contexts
type ChipSet struct {
	Default    []ChipAction
	Command    []ChipAction
	Diff       []ChipAction
	Generation []ChipAction
	Error      []ChipAction
}

// Predefined chip sets
var defaultChipSets = ChipSet{
	Default: []ChipAction{
		{
			ID:       "dry-run",
			Label:    "Dry-run",
			Command:  "dry-run",
			Tooltip:  "Execute command with --dry-run flag",
			Category: "execution",
			Icon:     "🧪",
		},
		{
			ID:       "save-yaml",
			Label:    "Save YAML",
			Command:  "save",
			Tooltip:  "Save response to YAML file",
			Category: "file",
			Icon:     "💾",
		},
		{
			ID:       "copy",
			Label:    "Copy",
			Command:  "copy",
			Tooltip:  "Copy content to clipboard",
			Category: "clipboard",
			Icon:     "📋",
		},
	},

	Command: []ChipAction{
		{
			ID:       "execute",
			Label:    "Execute",
			Command:  "execute",
			Tooltip:  "Execute the suggested command",
			Category: "execution",
			Icon:     "▶️",
		},
		{
			ID:       "dry-run",
			Label:    "Dry-run",
			Command:  "dry-run",
			Tooltip:  "Test command with --dry-run",
			Category: "execution",
			Icon:     "🧪",
		},
		{
			ID:       "explain",
			Label:    "Explain",
			Command:  "explain",
			Tooltip:  "Explain what this command does",
			Category: "help",
			Icon:     "❓",
		},
		{
			ID:       "save-yaml",
			Label:    "Save YAML",
			Command:  "save",
			Tooltip:  "Save to file",
			Category: "file",
			Icon:     "💾",
		},
	},

	Diff: []ChipAction{
		{
			ID:       "apply",
			Label:    "Apply",
			Command:  "apply-diff",
			Tooltip:  "Apply the diff to the file",
			Category: "execution",
			Icon:     "✅",
		},
		{
			ID:       "preview",
			Label:    "Preview",
			Command:  "preview-diff",
			Tooltip:  "Preview diff in detailed view",
			Category: "view",
			Icon:     "👁️",
		},
		{
			ID:       "edit",
			Label:    "Edit",
			Command:  "edit-diff",
			Tooltip:  "Modify the diff before applying",
			Category: "edit",
			Icon:     "✏️",
		},
		{
			ID:       "cancel",
			Label:    "Cancel",
			Command:  "cancel-diff",
			Tooltip:  "Cancel diff operation",
			Category: "cancel",
			Icon:     "❌",
		},
	},

	Generation: []ChipAction{
		{
			ID:       "save",
			Label:    "Save",
			Command:  "save-generated",
			Tooltip:  "Save generated content",
			Category: "file",
			Icon:     "💾",
		},
		{
			ID:       "preview",
			Label:    "Preview",
			Command:  "preview-generated",
			Tooltip:  "Preview generated content",
			Category: "view",
			Icon:     "👁️",
		},
		{
			ID:       "regenerate",
			Label:    "Regenerate",
			Command:  "regenerate",
			Tooltip:  "Generate again with different parameters",
			Category: "generate",
			Icon:     "🔄",
		},
	},

	Error: []ChipAction{
		{
			ID:       "retry",
			Label:    "Retry",
			Command:  "retry",
			Tooltip:  "Retry the failed operation",
			Category: "execution",
			Icon:     "🔄",
		},
		{
			ID:       "diagnose",
			Label:    "Diagnose",
			Command:  "diagnose",
			Tooltip:  "Run diagnostics for the error",
			Category: "debug",
			Icon:     "🔍",
		},
		{
			ID:       "fix",
			Label:    "Fix",
			Command:  "fix",
			Tooltip:  "Suggest a fix for the error",
			Category: "fix",
			Icon:     "🔧",
		},
	},
}

// NewQuickActionChips creates a new quick action chips component
func NewQuickActionChips(theme *Theme, messageID string) *QuickActionChips {
	return &QuickActionChips{
		theme:      theme,
		chips:      defaultChipSets.Default,
		focused:    -1,
		messageID:  messageID,
		enabled:    true,
		maxVisible: 6,
	}
}

// SetChips sets the available chips
func (qac *QuickActionChips) SetChips(chips []ChipAction) {
	qac.chips = chips
	qac.focused = -1
}

// SetChipSet sets chips from a predefined set
func (qac *QuickActionChips) SetChipSet(setType string) {
	switch strings.ToLower(setType) {
	case "command":
		qac.SetChips(defaultChipSets.Command)
	case "diff":
		qac.SetChips(defaultChipSets.Diff)
	case "generation":
		qac.SetChips(defaultChipSets.Generation)
	case "error":
		qac.SetChips(defaultChipSets.Error)
	default:
		qac.SetChips(defaultChipSets.Default)
	}
}

// AddCustomChip adds a custom chip to the current set
func (qac *QuickActionChips) AddCustomChip(chip ChipAction) {
	qac.chips = append(qac.chips, chip)
}

// SetEnabled enables or disables the chips
func (qac *QuickActionChips) SetEnabled(enabled bool) {
	qac.enabled = enabled
	if !enabled {
		qac.focused = -1
	}
}

// IsEnabled returns whether chips are enabled
func (qac *QuickActionChips) IsEnabled() bool {
	return qac.enabled
}

// MoveFocusNext moves focus to the next chip
func (qac *QuickActionChips) MoveFocusNext() {
	if !qac.enabled || len(qac.chips) == 0 {
		return
	}

	qac.focused++
	if qac.focused >= len(qac.chips) {
		qac.focused = 0
	}
}

// MoveFocusPrev moves focus to the previous chip
func (qac *QuickActionChips) MoveFocusPrev() {
	if !qac.enabled || len(qac.chips) == 0 {
		return
	}

	qac.focused--
	if qac.focused < 0 {
		qac.focused = len(qac.chips) - 1
	}
}

// ClearFocus removes focus from all chips
func (qac *QuickActionChips) ClearFocus() {
	qac.focused = -1
}

// GetFocusedChip returns the currently focused chip
func (qac *QuickActionChips) GetFocusedChip() *ChipAction {
	if qac.focused >= 0 && qac.focused < len(qac.chips) {
		return &qac.chips[qac.focused]
	}
	return nil
}

// HasFocus returns whether any chip has focus
func (qac *QuickActionChips) HasFocus() bool {
	return qac.focused >= 0
}

// Render renders the quick action chips
func (qac *QuickActionChips) Render() string {
	if !qac.enabled || len(qac.chips) == 0 {
		return ""
	}

	theme := CurrentTheme()

	var renderedChips []string
	visibleChips := qac.chips
	if len(visibleChips) > qac.maxVisible {
		visibleChips = visibleChips[:qac.maxVisible]
	}

	for i, chip := range visibleChips {
		isFocused := i == qac.focused
		renderedChips = append(renderedChips, qac.renderChip(chip, isFocused, theme))
	}

	// Add overflow indicator if there are more chips
	if len(qac.chips) > qac.maxVisible {
		overflowStyle := theme.HighlightStyle("comment")
		overflowIndicator := overflowStyle.Render(fmt.Sprintf("+%d", len(qac.chips)-qac.maxVisible))
		renderedChips = append(renderedChips, overflowIndicator)
	}

	// Join chips with spacing
	chipsRow := lipgloss.JoinHorizontal(lipgloss.Left, renderedChips...)

	// Add context label
	labelStyle := theme.HighlightStyle("comment").MarginRight(1)
	label := labelStyle.Render("Actions:")

	return lipgloss.JoinHorizontal(lipgloss.Left, label, chipsRow)
}

// renderChip renders a single chip
func (qac *QuickActionChips) renderChip(chip ChipAction, focused bool, theme *Theme) string {
	chipStyle := theme.ChipStyle(focused)

	// Build chip content
	var content strings.Builder

	// Add icon if present
	if chip.Icon != "" {
		content.WriteString(chip.Icon)
		content.WriteString(" ")
	}

	content.WriteString(chip.Label)

	// Add keyboard hint for focused chip
	if focused {
		content.WriteString(" ⏎")
	}

	renderedChip := chipStyle.Render(content.String())

	// Add margin between chips
	return lipgloss.NewStyle().MarginRight(1).Render(renderedChip)
}

// RenderWithTooltip renders chips with tooltip for focused chip
func (qac *QuickActionChips) RenderWithTooltip() string {
	chipsRow := qac.Render()

	if qac.focused >= 0 && qac.focused < len(qac.chips) {
		focusedChip := qac.chips[qac.focused]
		if focusedChip.Tooltip != "" {
			theme := CurrentTheme()
			tooltipStyle := theme.HighlightStyle("comment").
				Italic(true).
				MarginTop(1)
			tooltip := tooltipStyle.Render(fmt.Sprintf("💡 %s", focusedChip.Tooltip))

			return lipgloss.JoinVertical(lipgloss.Left, chipsRow, tooltip)
		}
	}

	return chipsRow
}

// Init initializes the QuickActionChips component
func (qac *QuickActionChips) Init() tea.Cmd {
	return nil
}

// View renders the QuickActionChips component
func (qac *QuickActionChips) View() string {
	return qac.Render()
}

// Update handles Bubble Tea messages for the chips
func (qac *QuickActionChips) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !qac.enabled {
		return qac, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab:
			qac.MoveFocusNext()
			return qac, nil

		case tea.KeyShiftTab:
			qac.MoveFocusPrev()
			return qac, nil

		case tea.KeyEnter:
			if focusedChip := qac.GetFocusedChip(); focusedChip != nil {
				return qac, qac.executeChipAction(*focusedChip)
			}
			return qac, nil

		case tea.KeyEsc:
			qac.ClearFocus()
			return qac, nil
		}
	}

	return qac, nil
}

// executeChipAction creates a command to execute the chip action
func (qac *QuickActionChips) executeChipAction(chip ChipAction) tea.Cmd {
	return func() tea.Msg {
		return ChipActionMsg{
			MessageID: qac.messageID,
			Action:    chip,
		}
	}
}

// ChipActionMsg is sent when a chip action is executed
type ChipActionMsg struct {
	MessageID string
	Action    ChipAction
}

// GetAvailableChipSets returns all available chip set names
func GetAvailableChipSets() []string {
	return []string{"default", "command", "diff", "generation", "error"}
}

// QuickActionRow is a convenience function to render chips for a message
func QuickActionRow(messageID, chipSet string, focused bool, theme *Theme) string {
	chips := NewQuickActionChips(theme, messageID)
	chips.SetChipSet(chipSet)
	chips.SetEnabled(focused)

	if focused {
		return chips.RenderWithTooltip()
	}
	return chips.Render()
}

// GetChipForCommand returns appropriate chip set based on command content
func GetChipForCommand(content string) string {
	content = strings.ToLower(content)

	if strings.Contains(content, "error") || strings.Contains(content, "failed") {
		return "error"
	}

	if strings.Contains(content, "diff") || strings.Contains(content, "patch") {
		return "diff"
	}

	if strings.Contains(content, "generate") || strings.Contains(content, "create") {
		return "generation"
	}

	if strings.Contains(content, "kubectl") || strings.Contains(content, "helm") {
		return "command"
	}

	return "default"
}
