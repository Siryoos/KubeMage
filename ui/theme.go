package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// ThemeType represents the available theme variants
type ThemeType int

const (
	ThemeDark ThemeType = iota
	ThemeLight
)

var currentTheme ThemeType = ThemeDark

// ColorPalette defines the colors for a theme
type ColorPalette struct {
	// Text colors
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Muted     lipgloss.Color
	Accent    lipgloss.Color
	Success   lipgloss.Color
	Warning   lipgloss.Color
	Error     lipgloss.Color
	Info      lipgloss.Color

	// Background colors
	Background lipgloss.Color
	Surface    lipgloss.Color
	Overlay    lipgloss.Color

	// Border colors
	Border      lipgloss.Color
	BorderFocus lipgloss.Color

	// Semantic colors
	User      lipgloss.Color
	Assistant lipgloss.Color
	System    lipgloss.Color
	Exec      lipgloss.Color

	// Risk level colors
	RiskLow      lipgloss.Color
	RiskMedium   lipgloss.Color
	RiskHigh     lipgloss.Color
	RiskCritical lipgloss.Color

	// Diff colors (color-blind safe)
	DiffAdded   lipgloss.Color
	DiffRemoved lipgloss.Color
	DiffContext lipgloss.Color

	// Syntax highlighting
	Keyword lipgloss.Color
	String  lipgloss.Color
	Number  lipgloss.Color
	Comment lipgloss.Color
}

// Theme contains all styling information for the UI
type Theme struct {
	Colors  ColorPalette
	Spacing struct {
		Small  int
		Medium int
		Large  int
	}
	BorderStyles struct {
		Normal  lipgloss.Border
		Rounded lipgloss.Border
		Thick   lipgloss.Border
	}
}

// Dark theme palette - WCAG AA compliant
var darkPalette = ColorPalette{
	Primary:   lipgloss.Color("255"), // White
	Secondary: lipgloss.Color("245"), // Light gray
	Muted:     lipgloss.Color("244"), // Medium gray
	Accent:    lipgloss.Color("75"),  // Cyan
	Success:   lipgloss.Color("82"),  // Green
	Warning:   lipgloss.Color("220"), // Yellow
	Error:     lipgloss.Color("196"), // Red
	Info:      lipgloss.Color("75"),  // Cyan

	Background: lipgloss.Color("0"),   // Black
	Surface:    lipgloss.Color("236"), // Dark gray
	Overlay:    lipgloss.Color("240"), // Medium dark gray

	Border:      lipgloss.Color("96"), // Light blue-gray
	BorderFocus: lipgloss.Color("75"), // Cyan

	User:      lipgloss.Color("63"),  // Blue
	Assistant: lipgloss.Color("82"),  // Green
	System:    lipgloss.Color("240"), // Gray
	Exec:      lipgloss.Color("220"), // Yellow

	RiskLow:      lipgloss.Color("82"),  // Green
	RiskMedium:   lipgloss.Color("220"), // Yellow
	RiskHigh:     lipgloss.Color("208"), // Orange
	RiskCritical: lipgloss.Color("196"), // Red

	// Color-blind safe diff colors using background highlighting
	DiffAdded:   lipgloss.Color("22"),  // Dark green background
	DiffRemoved: lipgloss.Color("52"),  // Dark red background
	DiffContext: lipgloss.Color("244"), // Gray

	Keyword: lipgloss.Color("178"), // Orange
	String:  lipgloss.Color("114"), // Light green
	Number:  lipgloss.Color("99"),  // Purple
	Comment: lipgloss.Color("244"), // Gray
}

// Light theme palette - Solarized-inspired, WCAG AA compliant
var lightPalette = ColorPalette{
	Primary:   lipgloss.Color("235"), // Dark gray
	Secondary: lipgloss.Color("240"), // Medium gray
	Muted:     lipgloss.Color("244"), // Light gray
	Accent:    lipgloss.Color("33"),  // Blue
	Success:   lipgloss.Color("28"),  // Dark green
	Warning:   lipgloss.Color("166"), // Orange
	Error:     lipgloss.Color("160"), // Dark red
	Info:      lipgloss.Color("33"),  // Blue

	Background: lipgloss.Color("253"), // Off-white
	Surface:    lipgloss.Color("254"), // Light gray
	Overlay:    lipgloss.Color("250"), // Medium light gray

	Border:      lipgloss.Color("248"), // Light gray
	BorderFocus: lipgloss.Color("33"),  // Blue

	User:      lipgloss.Color("33"),  // Blue
	Assistant: lipgloss.Color("28"),  // Dark green
	System:    lipgloss.Color("244"), // Gray
	Exec:      lipgloss.Color("166"), // Orange

	RiskLow:      lipgloss.Color("28"),  // Dark green
	RiskMedium:   lipgloss.Color("166"), // Orange
	RiskHigh:     lipgloss.Color("196"), // Red
	RiskCritical: lipgloss.Color("124"), // Dark red

	// Color-blind safe diff colors
	DiffAdded:   lipgloss.Color("194"), // Light green background
	DiffRemoved: lipgloss.Color("224"), // Light red background
	DiffContext: lipgloss.Color("250"), // Light gray

	Keyword: lipgloss.Color("166"), // Orange
	String:  lipgloss.Color("64"),  // Green
	Number:  lipgloss.Color("125"), // Purple
	Comment: lipgloss.Color("244"), // Gray
}

// Current theme instance
var currentThemeInstance *Theme

// Initialize themes
func init() {
	currentThemeInstance = &Theme{
		Colors: darkPalette,
		Spacing: struct {
			Small  int
			Medium int
			Large  int
		}{
			Small:  1,
			Medium: 2,
			Large:  4,
		},
		BorderStyles: struct {
			Normal  lipgloss.Border
			Rounded lipgloss.Border
			Thick   lipgloss.Border
		}{
			Normal:  lipgloss.NormalBorder(),
			Rounded: lipgloss.RoundedBorder(),
			Thick:   lipgloss.ThickBorder(),
		},
	}
}

// CurrentTheme returns the current theme instance
func CurrentTheme() *Theme {
	return currentThemeInstance
}

// ToggleTheme switches between dark and light themes
func ToggleTheme() {
	if currentTheme == ThemeDark {
		currentTheme = ThemeLight
		currentThemeInstance.Colors = lightPalette
	} else {
		currentTheme = ThemeDark
		currentThemeInstance.Colors = darkPalette
	}
}

// GetThemeType returns the current theme type
func GetThemeType() ThemeType {
	return currentTheme
}

// Style factory functions for common UI components

// HeaderStyle creates a styled header
func (t *Theme) HeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Colors.Primary).
		Background(t.Colors.Accent).
		Padding(0, t.Spacing.Medium)
}

// StatusStyle creates a styled status bar
func (t *Theme) StatusStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Colors.Secondary).
		Background(t.Colors.Surface).
		Padding(0, t.Spacing.Medium)
}

// PaneStyle creates a styled pane with border
func (t *Theme) PaneStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(t.BorderStyles.Rounded).
		BorderForeground(t.Colors.Border).
		Padding(0, t.Spacing.Small)
}

// PaneFocusedStyle creates a styled focused pane
func (t *Theme) PaneFocusedStyle() lipgloss.Style {
	return t.PaneStyle().
		BorderForeground(t.Colors.BorderFocus)
}

// InputStyle creates a styled input field
func (t *Theme) InputStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(t.BorderStyles.Rounded).
		BorderForeground(t.Colors.BorderFocus).
		Padding(0, t.Spacing.Small)
}

// ToastStyle creates a styled toast notification
func (t *Theme) ToastStyle(level string) lipgloss.Style {
	var color lipgloss.Color
	switch level {
	case "success":
		color = t.Colors.Success
	case "warning":
		color = t.Colors.Warning
	case "error":
		color = t.Colors.Error
	default:
		color = t.Colors.Info
	}

	return lipgloss.NewStyle().
		Foreground(t.Colors.Primary).
		Background(color).
		Padding(0, t.Spacing.Medium).
		Border(t.BorderStyles.Rounded).
		BorderForeground(color)
}

// ModalStyle creates a styled modal dialog
func (t *Theme) ModalStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(t.BorderStyles.Thick).
		BorderForeground(t.Colors.BorderFocus).
		Background(t.Colors.Surface).
		Padding(t.Spacing.Medium)
}

// ChipStyle creates a styled action chip
func (t *Theme) ChipStyle(focused bool) lipgloss.Style {
	style := lipgloss.NewStyle().
		Foreground(t.Colors.Accent).
		Background(t.Colors.Surface).
		Padding(0, t.Spacing.Small).
		Border(t.BorderStyles.Rounded).
		BorderForeground(t.Colors.Border)

	if focused {
		style = style.
			BorderForeground(t.Colors.BorderFocus).
			Bold(true)
	}

	return style
}

// RiskStyle returns appropriate styling for risk levels
func (t *Theme) RiskStyle(level string) lipgloss.Style {
	var color lipgloss.Color
	switch level {
	case "low":
		color = t.Colors.RiskLow
	case "medium", "med":
		color = t.Colors.RiskMedium
	case "high":
		color = t.Colors.RiskHigh
	case "critical":
		color = t.Colors.RiskCritical
	default:
		color = t.Colors.Muted
	}

	return lipgloss.NewStyle().Foreground(color)
}

// SenderStyle returns appropriate styling for message senders
func (t *Theme) SenderStyle(sender string) lipgloss.Style {
	var color lipgloss.Color
	switch sender {
	case "user", "User", "You":
		color = t.Colors.User
	case "assistant", "Assistant", "KubeMage":
		color = t.Colors.Assistant
	case "system", "System":
		color = t.Colors.System
	case "exec", "Exec", "Command":
		color = t.Colors.Exec
	default:
		color = t.Colors.Primary
	}

	return lipgloss.NewStyle().
		Foreground(color).
		Bold(true)
}

// DiffStyle returns appropriate styling for diff lines
func (t *Theme) DiffStyle(diffType string) lipgloss.Style {
	switch diffType {
	case "added", "+":
		return lipgloss.NewStyle().
			Background(t.Colors.DiffAdded).
			Foreground(t.Colors.Primary)
	case "removed", "-":
		return lipgloss.NewStyle().
			Background(t.Colors.DiffRemoved).
			Foreground(t.Colors.Primary)
	default:
		return lipgloss.NewStyle().
			Foreground(t.Colors.DiffContext)
	}
}

// HighlightStyle returns appropriate styling for syntax highlighting
func (t *Theme) HighlightStyle(tokenType string) lipgloss.Style {
	var color lipgloss.Color
	switch tokenType {
	case "keyword":
		color = t.Colors.Keyword
	case "string":
		color = t.Colors.String
	case "number":
		color = t.Colors.Number
	case "comment":
		color = t.Colors.Comment
	default:
		color = t.Colors.Primary
	}

	return lipgloss.NewStyle().Foreground(color)
}
