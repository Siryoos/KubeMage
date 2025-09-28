package ui

import (
	"fmt"
	"strings"
	"time"
)

// StatusInfo holds all information needed for the status bar
type StatusInfo struct {
	Context     string
	Namespace   string
	User        string
	ChatModel   string
	GenModel    string
	LiveTokens  int
	MaxTokens   int
	RiskLevel   string
	CurrentTime time.Time
	IsStreaming bool
	Layout      LayoutMode
}

// StatusBar manages the status bar display
type StatusBar struct {
	theme        *Theme
	info         StatusInfo
	lastUpdate   time.Time
	isProdWarn   bool
}

// NewStatusBar creates a new status bar
func NewStatusBar(theme *Theme) *StatusBar {
	return &StatusBar{
		theme: theme,
		info: StatusInfo{
			Context:     "(ctx)",
			Namespace:   "(ns)",
			User:        "(user)",
			ChatModel:   "(model)",
			GenModel:    "(model)",
			LiveTokens:  0,
			MaxTokens:   9999,
			RiskLevel:   "low",
			CurrentTime: time.Now(),
			IsStreaming: false,
			Layout:      LayoutThreePane,
		},
	}
}

// UpdateInfo updates the status bar information
func (sb *StatusBar) UpdateInfo(info StatusInfo) {
	sb.info = info
	sb.lastUpdate = time.Now()

	// Check if we're in a production-like environment
	sb.isProdWarn = strings.Contains(strings.ToLower(info.Namespace), "prod") ||
		strings.Contains(strings.ToLower(info.Context), "prod") ||
		strings.Contains(strings.ToLower(info.Context), "production")
}

// SetContext updates the context information
func (sb *StatusBar) SetContext(context, namespace string) {
	sb.info.Context = context
	sb.info.Namespace = namespace
	sb.isProdWarn = strings.Contains(strings.ToLower(namespace), "prod") ||
		strings.Contains(strings.ToLower(context), "prod")
}

// SetModels updates the model information
func (sb *StatusBar) SetModels(chat, generation string) {
	sb.info.ChatModel = chat
	sb.info.GenModel = generation
}

// SetTokens updates the token count
func (sb *StatusBar) SetTokens(live, max int) {
	sb.info.LiveTokens = live
	sb.info.MaxTokens = max
}

// SetRiskLevel updates the risk level
func (sb *StatusBar) SetRiskLevel(level string) {
	sb.info.RiskLevel = level
}

// SetStreaming updates the streaming status
func (sb *StatusBar) SetStreaming(streaming bool) {
	sb.info.IsStreaming = streaming
}

// SetLayout updates the layout mode
func (sb *StatusBar) SetLayout(layout LayoutMode) {
	sb.info.Layout = layout
}

// Render returns the formatted status bar
func (sb *StatusBar) Render() string {
	sb.info.CurrentTime = time.Now()

	// Build status components
	components := []string{
		sb.renderContext(),
		sb.renderNamespace(),
		sb.renderUser(),
		sb.renderModel(),
		sb.renderTokens(),
		sb.renderRisk(),
		sb.renderLayout(),
		sb.renderTime(),
	}

	// Filter out empty components
	var filteredComponents []string
	for _, comp := range components {
		if strings.TrimSpace(comp) != "" {
			filteredComponents = append(filteredComponents, comp)
		}
	}

	statusText := strings.Join(filteredComponents, "  ")

	// Apply appropriate styling based on production warning
	if sb.isProdWarn {
		return sb.theme.StatusStyle().
			Background(sb.theme.Colors.RiskHigh).
			Foreground(sb.theme.Colors.Primary).
			Bold(true).
			Render(statusText)
	}

	return sb.theme.StatusStyle().Render(statusText)
}

// renderContext formats the context component
func (sb *StatusBar) renderContext() string {
	ctx := strings.TrimSpace(sb.info.Context)
	if ctx == "" {
		ctx = "(ctx)"
	}
	return fmt.Sprintf("ctx:%s", ctx)
}

// renderNamespace formats the namespace component
func (sb *StatusBar) renderNamespace() string {
	ns := strings.TrimSpace(sb.info.Namespace)
	if ns == "" {
		ns = "(ns)"
	}
	return fmt.Sprintf("ns:%s", ns)
}

// renderUser formats the user component
func (sb *StatusBar) renderUser() string {
	user := strings.TrimSpace(sb.info.User)
	if user == "" {
		user = "(user)"
	}
	return fmt.Sprintf("user:%s", user)
}

// renderModel formats the model component
func (sb *StatusBar) renderModel() string {
	chatModel := sb.formatModelName(sb.info.ChatModel)
	genModel := sb.formatModelName(sb.info.GenModel)

	if chatModel == genModel {
		return fmt.Sprintf("model:%s", chatModel)
	}

	return fmt.Sprintf("model:%s/%s", chatModel, genModel)
}

// formatModelName formats a model name for display
func (sb *StatusBar) formatModelName(modelName string) string {
	trimmed := strings.TrimSpace(modelName)
	if trimmed == "" {
		return "(model)"
	}

	// Extract name@quant format
	parts := strings.Split(trimmed, ":")
	if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
		name := strings.TrimSpace(parts[0])
		quant := strings.TrimSpace(parts[1])
		return fmt.Sprintf("%s@%s", name, quant)
	}

	return trimmed
}

// renderTokens formats the token component
func (sb *StatusBar) renderTokens() string {
	tokens := sb.info.LiveTokens
	if tokens > sb.info.MaxTokens {
		tokens = sb.info.MaxTokens
	}

	if sb.info.IsStreaming {
		return fmt.Sprintf("tokens:%d~", tokens)
	}

	return fmt.Sprintf("tokens:%d", tokens)
}

// renderRisk formats the risk component with appropriate styling
func (sb *StatusBar) renderRisk() string {
	risk := strings.TrimSpace(sb.info.RiskLevel)
	if risk == "" {
		risk = "low"
	}

	return fmt.Sprintf("risk:%s", risk)
}

// renderLayout formats the layout component
func (sb *StatusBar) renderLayout() string {
	var layoutSymbol string
	switch sb.info.Layout {
	case LayoutThreePane:
		layoutSymbol = "▣"
	case LayoutVerticalSplit:
		layoutSymbol = "▥"
	case LayoutHorizontalSplit:
		layoutSymbol = "▤"
	case LayoutChatOnly:
		layoutSymbol = "□"
	default:
		layoutSymbol = "?"
	}

	return fmt.Sprintf("layout:%s", layoutSymbol)
}

// renderTime formats the time component
func (sb *StatusBar) renderTime() string {
	return fmt.Sprintf("time:%s", sb.info.CurrentTime.Format("15:04:05"))
}

// RenderRisk returns a styled risk indicator for use in other components
func (sb *StatusBar) RenderRisk() string {
	riskStyle := sb.theme.RiskStyle(sb.info.RiskLevel)
	return riskStyle.Render(strings.ToUpper(sb.info.RiskLevel))
}

// RenderCompact returns a compact version of the status bar for small screens
func (sb *StatusBar) RenderCompact() string {
	sb.info.CurrentTime = time.Now()

	// Compact format: ctx:prod ns:default risk:HIGH 15:04:05
	components := []string{
		sb.renderContext(),
		sb.renderNamespace(),
		sb.renderRisk(),
		sb.renderTime(),
	}

	statusText := strings.Join(components, " ")

	if sb.isProdWarn {
		return sb.theme.StatusStyle().
			Background(sb.theme.Colors.RiskHigh).
			Foreground(sb.theme.Colors.Primary).
			Bold(true).
			Render(statusText)
	}

	return sb.theme.StatusStyle().Render(statusText)
}

// GetStatusInfo returns the current status information
func (sb *StatusBar) GetStatusInfo() StatusInfo {
	return sb.info
}

// IsProductionWarning returns true if we're in a production-like environment
func (sb *StatusBar) IsProductionWarning() bool {
	return sb.isProdWarn
}

// RenderSpinner returns a spinner indicator for streaming states
func (sb *StatusBar) RenderSpinner() string {
	if !sb.info.IsStreaming {
		return ""
	}

	// Simple spinner animation based on current time
	spinners := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	index := int(time.Now().UnixNano()/100000000) % len(spinners)

	return sb.theme.HighlightStyle("keyword").Render(spinners[index])
}