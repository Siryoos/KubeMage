// ui_intelligence.go - Intelligent UI enhancements and smart features
package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// IntelligentUI manages smart UI features and enhancements
type IntelligentUI struct {
	riskLevel      RiskLevel
	quickActions   []QuickAction
	suggestions    []AISuggestion
	insights       []IntelligentInsight
	currentSession *AnalysisSession
	lastUpdate     time.Time
}

// AISuggestion represents intelligent suggestions for the user
type AISuggestion struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"` // "command", "explanation", "optimization"
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Action      string  `json:"action"`
	Confidence  float64 `json:"confidence"`
	Icon        string  `json:"icon"`
	Hotkey      string  `json:"hotkey"`
}

// NewIntelligentUI creates a new intelligent UI instance
func NewIntelligentUI() *IntelligentUI {
	return &IntelligentUI{
		quickActions: make([]QuickAction, 0),
		suggestions:  make([]AISuggestion, 0),
		insights:     make([]IntelligentInsight, 0),
		lastUpdate:   time.Now(),
	}
}

// UpdateIntelligence refreshes UI intelligence based on current context
func (ui *IntelligentUI) UpdateIntelligence(input string, context *KubeContextSummary) error {
	// Analyze current situation
	session, err := Intelligence.AnalyzeIntelligently(input, context)
	if err != nil {
		return err
	}

	ui.currentSession = session
	ui.riskLevel = ui.calculateUIRiskLevel(session, context)
	ui.quickActions = GenerateQuickActions(input, session.Intent.Alternatives, context)
	ui.suggestions = ui.generateAISuggestions(session)
	ui.insights = Intelligence.GetIntelligentInsights(session)
	ui.lastUpdate = time.Now()

	return nil
}

// calculateUIRiskLevel determines overall risk level for UI display
func (ui *IntelligentUI) calculateUIRiskLevel(session *AnalysisSession, context *KubeContextSummary) RiskLevel {
	risk := RiskLevel{
		Level:      "low",
		Factors:    []string{},
		Reversible: true,
	}

	// Check context risk factors
	if context != nil {
		// Production namespace detection
		prodPatterns := []string{"prod", "production", "live", "staging"}
		for _, pattern := range prodPatterns {
			if strings.Contains(strings.ToLower(context.Namespace), pattern) {
				risk.Level = "high"
				risk.Factors = append(risk.Factors, "production environment")
				break
			}
		}

		// Cluster health risk
		if len(context.PodProblemCounts) > 0 {
			problemCount := 0
			for _, count := range context.PodProblemCounts {
				problemCount += count
			}
			if problemCount > 5 {
				risk.Level = ui.escalateRisk(risk.Level, "medium")
				risk.Factors = append(risk.Factors, "multiple pod issues")
			}
		}
	}

	// Check intent risk
	if session != nil && session.Intent != nil {
		switch session.Intent.Mode {
		case ModeEdit:
			risk.Level = ui.escalateRisk(risk.Level, "medium")
			risk.Factors = append(risk.Factors, "configuration changes")
		case ModeGenerate:
			risk.Level = ui.escalateRisk(risk.Level, "medium")
			risk.Factors = append(risk.Factors, "resource creation")
		}

		// Low confidence increases risk
		if session.Intent.Confidence < 0.7 {
			risk.Level = ui.escalateRisk(risk.Level, "medium")
			risk.Factors = append(risk.Factors, "uncertain intent")
		}
	}

	return risk
}

// escalateRisk safely escalates risk level
func (ui *IntelligentUI) escalateRisk(current, new string) string {
	riskLevels := map[string]int{
		"low":      1,
		"medium":   2,
		"high":     3,
		"critical": 4,
	}

	if riskLevels[new] > riskLevels[current] {
		return new
	}
	return current
}

// generateAISuggestions creates intelligent suggestions for the user
func (ui *IntelligentUI) generateAISuggestions(session *AnalysisSession) []AISuggestion {
	var suggestions []AISuggestion

	if session == nil {
		return suggestions
	}

	// Generate suggestions based on analysis
	if session.RootCause != nil && session.RootCause.Confidence > 0.7 {
		suggestions = append(suggestions, AISuggestion{
			ID:          "root-cause-fix",
			Type:        "diagnostic",
			Title:       fmt.Sprintf("Fix %s", session.RootCause.RootCause),
			Description: fmt.Sprintf("High confidence (%.0f%%) solution available", session.RootCause.Confidence*100),
			Action:      "Apply recommended fix",
			Confidence:  session.RootCause.Confidence,
			Icon:        "🔧",
			Hotkey:      "F1",
		})
	}

	// Predictive suggestions - get them from the intelligence engine
	if session.Context != nil {
		predictiveActions := Intelligence.GetPredictiveActions("", session.Context)
		for i, action := range predictiveActions {
			if i >= 3 { // Limit to top 3 predictive suggestions
				break
			}

			hotkey := fmt.Sprintf("F%d", i+4) // F4, F5, F6 for predictive actions
			icon := "🔮"
			switch action.Type {
			case "diagnostic":
				icon = "🔍"
			case "command":
				icon = "⚡"
			case "explanation":
				icon = "💡"
			}

			suggestions = append(suggestions, AISuggestion{
				ID:          fmt.Sprintf("predictive-%d", i),
				Type:        "predictive",
				Title:       fmt.Sprintf("Predicted: %s", action.Description),
				Description: fmt.Sprintf("Based on patterns (%.0f%% confidence)", action.Confidence*100),
				Action:      action.Command,
				Confidence:  action.Confidence,
				Icon:        icon,
				Hotkey:      hotkey,
			})
		}
	}

	// Optimization suggestions
	criticalOpts := 0
	for _, opt := range session.Optimization {
		if opt.Severity == "critical" {
			criticalOpts++
		}
	}

	if criticalOpts > 0 {
		suggestions = append(suggestions, AISuggestion{
			ID:          "critical-optimization",
			Type:        "optimization",
			Title:       fmt.Sprintf("%d Critical Optimizations", criticalOpts),
			Description: "Resource optimizations can improve performance",
			Action:      "Review optimizations",
			Confidence:  0.9,
			Icon:        "⚡",
			Hotkey:      "F2",
		})
	}

	// Intent clarification for low confidence
	if session.Intent.Confidence < 0.7 && len(session.Intent.Alternatives) > 0 {
		suggestions = append(suggestions, AISuggestion{
			ID:          "clarify-intent",
			Type:        "clarification",
			Title:       "Clarify Intent",
			Description: fmt.Sprintf("Multiple options available: %s", strings.Join(ui.modeStrings(session.Intent.Alternatives), ", ")),
			Action:      "Choose specific action",
			Confidence:  session.Intent.Confidence,
			Icon:        "❓",
			Hotkey:      "F3",
		})
	}

	return suggestions
}

// modeStrings converts AgentMode slice to string slice
func (ui *IntelligentUI) modeStrings(modes []AgentMode) []string {
	result := make([]string, len(modes))
	for i, mode := range modes {
		result[i] = string(mode)
	}
	return result
}

// FormatRiskIndicator creates a styled risk indicator for the status bar
func (ui *IntelligentUI) FormatRiskIndicator() string {
	var style lipgloss.Style
	var indicator string

	switch ui.riskLevel.Level {
	case "low":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // Green
		indicator = "🟢 LOW"
	case "medium":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // Yellow
		indicator = "🟡 MED"
	case "high":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("1")) // Red
		indicator = "🔴 HIGH"
	case "critical":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Blink(true) // Blinking red
		indicator = "🚨 CRIT"
	default:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // Gray
		indicator = "❓ UNK"
	}

	return style.Render(indicator)
}

// FormatQuickActions creates styled quick action buttons
func (ui *IntelligentUI) FormatQuickActions() string {
	if len(ui.quickActions) == 0 {
		return ""
	}

	var actions []string
	for i, action := range ui.quickActions {
		if i >= 3 { // Limit to 3 quick actions
			break
		}

		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("4")).
			Padding(0, 1).
			Margin(0, 1)

		actionText := fmt.Sprintf("%s %s", action.Icon, action.Label)
		actions = append(actions, style.Render(actionText))
	}

	return strings.Join(actions, "")
}

// FormatSuggestions creates a styled suggestions panel
func (ui *IntelligentUI) FormatSuggestions() string {
	if len(ui.suggestions) == 0 {
		return ""
	}

	var suggestionTexts []string

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("6")).
		Bold(true)

	suggestionTexts = append(suggestionTexts, headerStyle.Render("💡 AI Suggestions:"))

	for i, suggestion := range ui.suggestions {
		if i >= 3 { // Limit display
			break
		}

		confidence := fmt.Sprintf("(%.0f%%)", suggestion.Confidence*100)

		suggestionStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Margin(0, 2)

		confidenceStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true)

		text := fmt.Sprintf("%s %s %s %s",
			suggestion.Icon,
			suggestion.Title,
			confidenceStyle.Render(confidence),
			fmt.Sprintf("[%s]", suggestion.Hotkey))

		suggestionTexts = append(suggestionTexts, suggestionStyle.Render(text))
	}

	return strings.Join(suggestionTexts, "\n")
}

// FormatInsights creates a styled insights panel
func (ui *IntelligentUI) FormatInsights() string {
	if len(ui.insights) == 0 {
		return ""
	}

	var insightTexts []string

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("5")).
		Bold(true)

	insightTexts = append(insightTexts, headerStyle.Render("🧠 Intelligent Insights:"))

	for i, insight := range ui.insights {
		if i >= 2 { // Limit display
			break
		}

		titleStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Bold(true)

		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("7")).
			Margin(0, 2)

		confidenceStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true)

		title := titleStyle.Render(insight.Title)
		confidence := confidenceStyle.Render(fmt.Sprintf("(%.0f%% confidence)", insight.Confidence*100))
		description := descStyle.Render(insight.Description)

		insightText := fmt.Sprintf("%s %s\n%s", title, confidence, description)
		insightTexts = append(insightTexts, insightText)
	}

	return strings.Join(insightTexts, "\n\n")
}

// FormatIntelligentStatus creates an enhanced status line with intelligence
func (ui *IntelligentUI) FormatIntelligentStatus(context *KubeContextSummary, model string) string {
	var statusParts []string

	// Context info
	if context != nil {
		contextStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
		nsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))

		statusParts = append(statusParts,
			fmt.Sprintf("ctx:%s", contextStyle.Render(context.Context)),
			fmt.Sprintf("ns:%s", nsStyle.Render(context.Namespace)))
	}

	// Model info
	modelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	statusParts = append(statusParts, fmt.Sprintf("model:%s", modelStyle.Render(model)))

	// Risk indicator
	statusParts = append(statusParts, fmt.Sprintf("risk:%s", ui.FormatRiskIndicator()))

	// Time
	timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	statusParts = append(statusParts, fmt.Sprintf("time:%s", timeStyle.Render(time.Now().Format("15:04:05"))))

	// Intelligence indicator
	if ui.currentSession != nil {
		confStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
		confidence := fmt.Sprintf("%.0f%%", ui.currentSession.Confidence*100)
		statusParts = append(statusParts, fmt.Sprintf("ai:%s", confStyle.Render(confidence)))
	}

	return strings.Join(statusParts, " ")
}

// GetQuickAction retrieves a quick action by index
func (ui *IntelligentUI) GetQuickAction(index int) *QuickAction {
	if index >= 0 && index < len(ui.quickActions) {
		return &ui.quickActions[index]
	}
	return nil
}

// GetSuggestion retrieves a suggestion by ID
func (ui *IntelligentUI) GetSuggestion(id string) *AISuggestion {
	for _, suggestion := range ui.suggestions {
		if suggestion.ID == id {
			return &suggestion
		}
	}
	return nil
}

// HandleHotkey processes hotkey inputs for suggestions
func (ui *IntelligentUI) HandleHotkey(key string) *AISuggestion {
	for _, suggestion := range ui.suggestions {
		if suggestion.Hotkey == key {
			return &suggestion
		}
	}
	return nil
}

// FormatIntelligencePanel creates a comprehensive intelligence panel
func (ui *IntelligentUI) FormatIntelligencePanel() string {
	var sections []string

	// Quick actions
	if quickActions := ui.FormatQuickActions(); quickActions != "" {
		sections = append(sections, quickActions)
	}

	// Suggestions
	if suggestions := ui.FormatSuggestions(); suggestions != "" {
		sections = append(sections, suggestions)
	}

	// Insights
	if insights := ui.FormatInsights(); insights != "" {
		sections = append(sections, insights)
	}

	if len(sections) == 0 {
		return ""
	}

	// Wrap in a styled panel
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("4")).
		Padding(1).
		Margin(1)

	return panelStyle.Render(strings.Join(sections, "\n\n"))
}

// GetCurrentRiskLevel returns the current risk assessment
func (ui *IntelligentUI) GetCurrentRiskLevel() RiskLevel {
	return ui.riskLevel
}

// GetCurrentSession returns the current analysis session
func (ui *IntelligentUI) GetCurrentSession() *AnalysisSession {
	return ui.currentSession
}

// IsHighRisk checks if current situation is high risk
func (ui *IntelligentUI) IsHighRisk() bool {
	return ui.riskLevel.Level == "high" || ui.riskLevel.Level == "critical"
}

// GetIntelligenceAge returns how old the current intelligence is
func (ui *IntelligentUI) GetIntelligenceAge() time.Duration {
	return time.Since(ui.lastUpdate)
}

// NeedsRefresh checks if intelligence should be refreshed
func (ui *IntelligentUI) NeedsRefresh() bool {
	return ui.GetIntelligenceAge() > 30*time.Second
}
