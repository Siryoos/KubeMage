package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ModalType represents different types of modals
type ModalType int

const (
	ModalConfirm ModalType = iota
	ModalDanger
	ModalInfo
	ModalInput
)

// ModalResult represents the result of a modal interaction
type ModalResult int

const (
	ModalResultNone ModalResult = iota
	ModalResultConfirm
	ModalResultCancel
	ModalResultTypedConfirm
)

// Modal represents a modal dialog
type Modal struct {
	Type          ModalType
	Title         string
	Message       string
	RequireTyped  string // For dangerous operations, require typing this exact string
	Buttons       []string
	DefaultButton int
	CurrentButton int
	UserInput     string
	CursorPos     int
	IsOpen        bool
	Width         int
	Height        int
	Result        ModalResult
}

// ModalManager manages modal dialogs
type ModalManager struct {
	theme      *Theme
	modal      *Modal
	overlayDim bool
}

// NewModalManager creates a new modal manager
func NewModalManager(theme *Theme) *ModalManager {
	return &ModalManager{
		theme:      theme,
		modal:      nil,
		overlayDim: true,
	}
}

// ShowConfirm shows a confirmation modal
func (mm *ModalManager) ShowConfirm(title, message string) {
	mm.modal = &Modal{
		Type:          ModalConfirm,
		Title:         title,
		Message:       message,
		Buttons:       []string{"OK", "Cancel"},
		DefaultButton: 0,
		CurrentButton: 0,
		IsOpen:        true,
		Width:         60,
		Height:        12,
		Result:        ModalResultNone,
	}
}

// ShowDanger shows a dangerous operation confirmation modal
func (mm *ModalManager) ShowDanger(title, message, requireTyped string) {
	mm.modal = &Modal{
		Type:          ModalDanger,
		Title:         title,
		Message:       message,
		RequireTyped:  requireTyped,
		Buttons:       []string{"Confirm", "Cancel"},
		DefaultButton: 1, // Default to Cancel for safety
		CurrentButton: 1,
		IsOpen:        true,
		Width:         70,
		Height:        15,
		Result:        ModalResultNone,
		UserInput:     "",
	}
}

// ShowInfo shows an informational modal
func (mm *ModalManager) ShowInfo(title, message string) {
	mm.modal = &Modal{
		Type:          ModalInfo,
		Title:         title,
		Message:       message,
		Buttons:       []string{"OK"},
		DefaultButton: 0,
		CurrentButton: 0,
		IsOpen:        true,
		Width:         50,
		Height:        10,
		Result:        ModalResultNone,
	}
}

// ShowInput shows an input modal
func (mm *ModalManager) ShowInput(title, message, placeholder string) {
	mm.modal = &Modal{
		Type:          ModalInput,
		Title:         title,
		Message:       message,
		Buttons:       []string{"OK", "Cancel"},
		DefaultButton: 0,
		CurrentButton: 0,
		IsOpen:        true,
		Width:         60,
		Height:        12,
		Result:        ModalResultNone,
		UserInput:     placeholder,
		CursorPos:     len(placeholder),
	}
}

// Close closes the current modal
func (mm *ModalManager) Close() {
	mm.modal = nil
}

// IsOpen returns whether a modal is currently open
func (mm *ModalManager) IsOpen() bool {
	return mm.modal != nil && mm.modal.IsOpen
}

// GetResult returns the modal result
func (mm *ModalManager) GetResult() ModalResult {
	if mm.modal == nil {
		return ModalResultNone
	}
	return mm.modal.Result
}

// GetUserInput returns the user input for input modals
func (mm *ModalManager) GetUserInput() string {
	if mm.modal == nil {
		return ""
	}
	return mm.modal.UserInput
}

// Update handles modal input
func (mm *ModalManager) Update(msg tea.Msg) tea.Cmd {
	if !mm.IsOpen() {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			mm.modal.Result = ModalResultCancel
			mm.Close()
			return nil

		case tea.KeyEnter:
			return mm.handleEnter()

		case tea.KeyTab, tea.KeyRight:
			mm.moveToNextButton()
			return nil

		case tea.KeyShiftTab, tea.KeyLeft:
			mm.moveToPrevButton()
			return nil

		default:
			// Handle typing for danger modals and input modals
			if mm.modal.Type == ModalDanger || mm.modal.Type == ModalInput {
				return mm.handleTextInput(msg)
			}
		}
	}

	return nil
}

// handleEnter handles Enter key press
func (mm *ModalManager) handleEnter() tea.Cmd {
	switch mm.modal.Type {
	case ModalDanger:
		if mm.modal.CurrentButton == 0 { // Confirm button
			if mm.modal.UserInput == mm.modal.RequireTyped {
				mm.modal.Result = ModalResultTypedConfirm
			} else {
				// Invalid input, don't close
				return nil
			}
		} else {
			mm.modal.Result = ModalResultCancel
		}

	case ModalInput:
		if mm.modal.CurrentButton == 0 { // OK button
			mm.modal.Result = ModalResultConfirm
		} else {
			mm.modal.Result = ModalResultCancel
		}

	default:
		if mm.modal.CurrentButton == 0 {
			mm.modal.Result = ModalResultConfirm
		} else {
			mm.modal.Result = ModalResultCancel
		}
	}

	mm.Close()
	return nil
}

// handleTextInput handles text input for danger and input modals
func (mm *ModalManager) handleTextInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyBackspace:
		if len(mm.modal.UserInput) > 0 && mm.modal.CursorPos > 0 {
			mm.modal.UserInput = mm.modal.UserInput[:mm.modal.CursorPos-1] +
				mm.modal.UserInput[mm.modal.CursorPos:]
			mm.modal.CursorPos--
		}

	case tea.KeyDelete:
		if mm.modal.CursorPos < len(mm.modal.UserInput) {
			mm.modal.UserInput = mm.modal.UserInput[:mm.modal.CursorPos] +
				mm.modal.UserInput[mm.modal.CursorPos+1:]
		}

	case tea.KeyCtrlA:
		mm.modal.CursorPos = 0

	case tea.KeyCtrlE:
		mm.modal.CursorPos = len(mm.modal.UserInput)

	case tea.KeyRunes:
		// Insert characters at cursor position
		runes := string(msg.Runes)
		mm.modal.UserInput = mm.modal.UserInput[:mm.modal.CursorPos] +
			runes + mm.modal.UserInput[mm.modal.CursorPos:]
		mm.modal.CursorPos += len(runes)
	}

	return nil
}

// moveToNextButton moves to the next button
func (mm *ModalManager) moveToNextButton() {
	if mm.modal.CurrentButton < len(mm.modal.Buttons)-1 {
		mm.modal.CurrentButton++
	} else {
		mm.modal.CurrentButton = 0
	}
}

// moveToPrevButton moves to the previous button
func (mm *ModalManager) moveToPrevButton() {
	if mm.modal.CurrentButton > 0 {
		mm.modal.CurrentButton--
	} else {
		mm.modal.CurrentButton = len(mm.modal.Buttons) - 1
	}
}

// Render renders the modal dialog
func (mm *ModalManager) Render(screenWidth, screenHeight int) string {
	if !mm.IsOpen() {
		return ""
	}

	// Create overlay background
	overlay := ""
	if mm.overlayDim {
		overlay = mm.renderOverlay(screenWidth, screenHeight)
	}

	// Render modal content
	modalContent := mm.renderModal()

	// Center the modal
	centeredModal := lipgloss.Place(
		screenWidth, screenHeight,
		lipgloss.Center, lipgloss.Center,
		modalContent,
	)

	if overlay != "" {
		return overlay + "\n" + centeredModal
	}

	return centeredModal
}

// renderOverlay renders the dimmed background overlay
func (mm *ModalManager) renderOverlay(width, height int) string {
	overlayStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("0")).
		Foreground(lipgloss.Color("0")).
		Width(width).
		Height(height)

	return overlayStyle.Render(strings.Repeat(" ", width*height))
}

// renderModal renders the modal content
func (mm *ModalManager) renderModal() string {
	var sections []string

	// Title
	titleStyle := mm.getTitleStyle()
	title := titleStyle.Render(mm.modal.Title)
	sections = append(sections, title)

	// Empty line
	sections = append(sections, "")

	// Message
	messageStyle := mm.theme.HighlightStyle("string")
	message := messageStyle.Render(mm.wrapText(mm.modal.Message, mm.modal.Width-4))
	sections = append(sections, message)

	// Type-specific content
	switch mm.modal.Type {
	case ModalDanger:
		sections = append(sections, mm.renderDangerInput())
	case ModalInput:
		sections = append(sections, mm.renderTextInput())
	}

	// Empty line before buttons
	sections = append(sections, "")

	// Buttons
	buttons := mm.renderButtons()
	sections = append(sections, buttons)

	// Combine all sections
	content := strings.Join(sections, "\n")

	// Apply modal styling
	modalStyle := mm.theme.ModalStyle().
		Width(mm.modal.Width).
		Align(lipgloss.Center)

	return modalStyle.Render(content)
}

// renderDangerInput renders the danger confirmation input
func (mm *ModalManager) renderDangerInput() string {
	var sections []string

	// Warning
	warningStyle := mm.theme.HighlightStyle("string").
		Foreground(mm.theme.Colors.Warning).
		Bold(true)
	warning := warningStyle.Render(fmt.Sprintf("Type '%s' to confirm:", mm.modal.RequireTyped))
	sections = append(sections, warning)

	// Input field
	inputContent := mm.modal.UserInput
	if mm.modal.CursorPos < len(inputContent) {
		// Insert cursor
		inputContent = inputContent[:mm.modal.CursorPos] + "█" + inputContent[mm.modal.CursorPos:]
	} else {
		inputContent += "█"
	}

	inputStyle := mm.theme.InputStyle().Width(mm.modal.Width - 8)

	// Color input based on correctness
	if mm.modal.UserInput == mm.modal.RequireTyped {
		inputStyle = inputStyle.BorderForeground(mm.theme.Colors.Success)
	} else if len(mm.modal.UserInput) > 0 {
		inputStyle = inputStyle.BorderForeground(mm.theme.Colors.Error)
	}

	input := inputStyle.Render(inputContent)
	sections = append(sections, input)

	return strings.Join(sections, "\n")
}

// renderTextInput renders a regular text input
func (mm *ModalManager) renderTextInput() string {
	inputContent := mm.modal.UserInput
	if mm.modal.CursorPos < len(inputContent) {
		// Insert cursor
		inputContent = inputContent[:mm.modal.CursorPos] + "█" + inputContent[mm.modal.CursorPos:]
	} else {
		inputContent += "█"
	}

	inputStyle := mm.theme.InputStyle().Width(mm.modal.Width - 8)
	return inputStyle.Render(inputContent)
}

// renderButtons renders the modal buttons
func (mm *ModalManager) renderButtons() string {
	var buttons []string

	for i, buttonText := range mm.modal.Buttons {
		var buttonStyle lipgloss.Style

		if i == mm.modal.CurrentButton {
			buttonStyle = mm.theme.ChipStyle(true).Bold(true)
		} else {
			buttonStyle = mm.theme.ChipStyle(false)
		}

		// Special styling for danger operations
		if mm.modal.Type == ModalDanger {
			if i == 0 { // Confirm button
				if mm.modal.UserInput == mm.modal.RequireTyped {
					buttonStyle = buttonStyle.Foreground(mm.theme.Colors.Success)
				} else {
					buttonStyle = buttonStyle.Foreground(mm.theme.Colors.Muted)
				}
			} else { // Cancel button
				buttonStyle = buttonStyle.Foreground(mm.theme.Colors.Warning)
			}
		}

		button := buttonStyle.Render(buttonText)
		buttons = append(buttons, button)
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, buttons...)
}

// getTitleStyle returns the appropriate title style
func (mm *ModalManager) getTitleStyle() lipgloss.Style {
	switch mm.modal.Type {
	case ModalDanger:
		return mm.theme.HeaderStyle().
			Background(mm.theme.Colors.Error).
			Foreground(mm.theme.Colors.Primary)
	case ModalInfo:
		return mm.theme.HeaderStyle().
			Background(mm.theme.Colors.Info).
			Foreground(mm.theme.Colors.Primary)
	default:
		return mm.theme.HeaderStyle()
	}
}

// wrapText wraps text to fit within the specified width
func (mm *ModalManager) wrapText(text string, width int) string {
	if len(text) <= width {
		return text
	}

	var wrapped []string
	words := strings.Fields(text)
	var currentLine strings.Builder

	for _, word := range words {
		if currentLine.Len()+len(word)+1 > width {
			if currentLine.Len() > 0 {
				wrapped = append(wrapped, currentLine.String())
				currentLine.Reset()
			}
		}

		if currentLine.Len() > 0 {
			currentLine.WriteString(" ")
		}
		currentLine.WriteString(word)
	}

	if currentLine.Len() > 0 {
		wrapped = append(wrapped, currentLine.String())
	}

	return strings.Join(wrapped, "\n")
}

// SetOverlayDim sets whether to show a dimmed overlay
func (mm *ModalManager) SetOverlayDim(dim bool) {
	mm.overlayDim = dim
}