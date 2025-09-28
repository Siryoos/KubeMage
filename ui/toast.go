package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ToastLevel represents different toast severity levels
type ToastLevel int

const (
	ToastInfo ToastLevel = iota
	ToastSuccess
	ToastWarning
	ToastError
)

// Toast represents a single toast notification
type Toast struct {
	ID        string
	Level     ToastLevel
	Message   string
	Duration  time.Duration
	CreatedAt time.Time
	Visible   bool
}

// ToastManager manages a queue of toast notifications
type ToastManager struct {
	theme           *Theme
	toasts          []Toast
	maxToasts       int
	defaultDuration time.Duration
	position        ToastPosition
	width           int
}

// ToastPosition represents where toasts are displayed
type ToastPosition int

const (
	ToastTopRight ToastPosition = iota
	ToastTopLeft
	ToastBottomRight
	ToastBottomLeft
	ToastCenter
)

// NewToastManager creates a new toast manager
func NewToastManager(theme *Theme) *ToastManager {
	return &ToastManager{
		theme:           theme,
		toasts:          []Toast{},
		maxToasts:       5,
		defaultDuration: 3 * time.Second,
		position:        ToastTopRight,
		width:           40,
	}
}

// AddToast adds a new toast notification
func (tm *ToastManager) AddToast(level ToastLevel, message string) string {
	toastID := tm.generateID()
	toast := Toast{
		ID:        toastID,
		Level:     level,
		Message:   message,
		Duration:  tm.defaultDuration,
		CreatedAt: time.Now(),
		Visible:   true,
	}

	tm.toasts = append(tm.toasts, toast)

	// Remove old toasts if we exceed max
	if len(tm.toasts) > tm.maxToasts {
		tm.toasts = tm.toasts[len(tm.toasts)-tm.maxToasts:]
	}

	return toastID
}

// Info adds an info toast
func (tm *ToastManager) Info(message string) string {
	return tm.AddToast(ToastInfo, message)
}

// Success adds a success toast
func (tm *ToastManager) Success(message string) string {
	return tm.AddToast(ToastSuccess, message)
}

// Warning adds a warning toast
func (tm *ToastManager) Warning(message string) string {
	return tm.AddToast(ToastWarning, message)
}

// Error adds an error toast
func (tm *ToastManager) Error(message string) string {
	return tm.AddToast(ToastError, message)
}

// RemoveToast removes a toast by ID
func (tm *ToastManager) RemoveToast(id string) {
	for i, toast := range tm.toasts {
		if toast.ID == id {
			tm.toasts = append(tm.toasts[:i], tm.toasts[i+1:]...)
			break
		}
	}
}

// DismissAll removes all toasts
func (tm *ToastManager) DismissAll() {
	tm.toasts = []Toast{}
}

// Update handles automatic toast expiration
func (tm *ToastManager) Update() tea.Cmd {
	now := time.Now()
	var expiredToasts []string

	for _, toast := range tm.toasts {
		if now.Sub(toast.CreatedAt) > toast.Duration {
			expiredToasts = append(expiredToasts, toast.ID)
		}
	}

	// Remove expired toasts
	for _, id := range expiredToasts {
		tm.RemoveToast(id)
	}

	// Return a command to check again later
	if len(tm.toasts) > 0 {
		return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
			return ToastUpdateMsg{}
		})
	}

	return nil
}

// Render renders all visible toasts
func (tm *ToastManager) Render() string {
	if len(tm.toasts) == 0 {
		return ""
	}

	var renderedToasts []string
	for _, toast := range tm.toasts {
		if toast.Visible {
			renderedToast := tm.renderToast(toast)
			renderedToasts = append(renderedToasts, renderedToast)
		}
	}

	if len(renderedToasts) == 0 {
		return ""
	}

	// Stack toasts vertically
	content := lipgloss.JoinVertical(lipgloss.Left, renderedToasts...)

	// Position based on settings
	switch tm.position {
	case ToastTopRight:
		return lipgloss.NewStyle().
			Align(lipgloss.Right).
			MarginTop(1).
			MarginRight(2).
			Render(content)
	case ToastTopLeft:
		return lipgloss.NewStyle().
			Align(lipgloss.Left).
			MarginTop(1).
			MarginLeft(2).
			Render(content)
	case ToastBottomRight:
		return lipgloss.NewStyle().
			Align(lipgloss.Right).
			MarginBottom(1).
			MarginRight(2).
			Render(content)
	case ToastBottomLeft:
		return lipgloss.NewStyle().
			Align(lipgloss.Left).
			MarginBottom(1).
			MarginLeft(2).
			Render(content)
	default:
		return lipgloss.NewStyle().
			Align(lipgloss.Center).
			Render(content)
	}
}

// renderToast renders a single toast
func (tm *ToastManager) renderToast(toast Toast) string {
	var levelIcon string
	var levelStyle lipgloss.Style

	switch toast.Level {
	case ToastSuccess:
		levelIcon = "✅"
		levelStyle = tm.theme.ToastStyle("success")
	case ToastWarning:
		levelIcon = "⚠️"
		levelStyle = tm.theme.ToastStyle("warning")
	case ToastError:
		levelIcon = "❌"
		levelStyle = tm.theme.ToastStyle("error")
	default:
		levelIcon = "ℹ️"
		levelStyle = tm.theme.ToastStyle("info")
	}

	// Calculate remaining time for visual feedback
	elapsed := time.Since(toast.CreatedAt)
	remaining := toast.Duration - elapsed

	// Add progress indicator if toast is about to expire
	var progressIndicator string
	if remaining < time.Second {
		progressIndicator = " ⏰"
	}

	content := levelIcon + " " + toast.Message + progressIndicator

	return levelStyle.
		Width(tm.width).
		Render(content)
}

// generateID generates a unique ID for toasts
func (tm *ToastManager) generateID() string {
	return time.Now().Format("20060102150405.000000")
}

// SetPosition sets the toast display position
func (tm *ToastManager) SetPosition(position ToastPosition) {
	tm.position = position
}

// SetMaxToasts sets the maximum number of concurrent toasts
func (tm *ToastManager) SetMaxToasts(max int) {
	tm.maxToasts = max
}

// SetDefaultDuration sets the default toast duration
func (tm *ToastManager) SetDefaultDuration(duration time.Duration) {
	tm.defaultDuration = duration
}

// SetWidth sets the toast width
func (tm *ToastManager) SetWidth(width int) {
	tm.width = width
}

// GetActiveToasts returns the number of active toasts
func (tm *ToastManager) GetActiveToasts() int {
	count := 0
	for _, toast := range tm.toasts {
		if toast.Visible {
			count++
		}
	}
	return count
}

// HasToasts returns whether there are any active toasts
func (tm *ToastManager) HasToasts() bool {
	return len(tm.toasts) > 0
}

// ToastUpdateMsg is sent when toasts need to be updated
type ToastUpdateMsg struct{}

// ToastExpiredMsg is sent when a toast expires
type ToastExpiredMsg struct {
	ID string
}

// ToastDismissMsg is sent when a toast is manually dismissed
type ToastDismissMsg struct {
	ID string
}
