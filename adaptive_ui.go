// adaptive_ui.go - Adaptive UI management with content-aware layout switching
package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// AdaptiveUIManager manages dynamic UI adaptations
type AdaptiveUIManager struct {
	streamingManager    *StreamingIntelligenceManager
	layoutEngine        *AdaptiveLayoutEngine
	contentAnalyzer     *ContentAnalyzer
	userBehaviorTracker *UserBehaviorTracker
	performanceMonitor  *PerformanceMonitor
	mu                  sync.RWMutex
}

// AdaptiveLayoutEngine handles intelligent layout switching
type AdaptiveLayoutEngine struct {
	currentLayout       layoutMode
	layoutHistory       []LayoutTransition
	contentTypeRules    map[ContentType]layoutMode
	screenSizeRules     map[ScreenSize]layoutMode
	userPreferenceRules map[string]layoutMode
	autoSwitchEnabled   bool
	transitionDuration  time.Duration
	mu                  sync.RWMutex
}

// ContentAnalyzer analyzes content to determine optimal layout
type ContentAnalyzer struct {
	yamlAnalyzer    *YAMLAnalyzer
	logAnalyzer     *LogAnalyzer
	diffAnalyzer    *DiffAnalyzer
	commandAnalyzer *CommandAnalyzer
	riskAnalyzer    *RiskAnalyzer
	patterns        map[string]*ContentPattern
	mu              sync.RWMutex
}

// UserBehaviorTracker tracks user behavior for UI adaptation
type UserBehaviorTracker struct {
	activityLevel       ActivityLevel
	interactionHistory  []UIInteraction
	preferences         *UserUIPreferences
	adaptationRules     []AdaptationRule
	lastActivity        time.Time
	mu                  sync.RWMutex
}

// PerformanceMonitor monitors UI performance for optimization
type PerformanceMonitor struct {
	renderTimes         []time.Duration
	layoutSwitchTimes   []time.Duration
	memoryUsage         []float64
	cpuUsage            []float64
	userSatisfaction    float64
	optimizationEnabled bool
	mu                  sync.RWMutex
}

// ContentType represents different types of content
type ContentType int

const (
	ContentTypeYAML ContentType = iota
	ContentTypeLogs
	ContentTypeDiff
	ContentTypeCommands
	ContentTypeError
	ContentTypeHelp
	ContentTypeMetrics
	ContentTypeGeneral
)

// ScreenSize represents different screen size categories
type ScreenSize int

const (
	ScreenSizeSmall ScreenSize = iota
	ScreenSizeMedium
	ScreenSizeLarge
	ScreenSizeExtraLarge
)

// ActivityLevel represents user activity levels
type ActivityLevel int

const (
	ActivityIdle ActivityLevel = iota
	ActivityLight
	ActivityModerate
	ActivityIntense
)

// LayoutTransition represents a layout change event
type LayoutTransition struct {
	FromLayout  layoutMode    `json:"from_layout"`
	ToLayout    layoutMode    `json:"to_layout"`
	Trigger     string        `json:"trigger"`
	Timestamp   time.Time     `json:"timestamp"`
	Duration    time.Duration `json:"duration"`
	UserAction  bool          `json:"user_action"`
	Successful  bool          `json:"successful"`
}

// ContentPattern represents learned content patterns
type ContentPattern struct {
	Pattern         string        `json:"pattern"`
	ContentType     ContentType   `json:"content_type"`
	OptimalLayout   layoutMode    `json:"optimal_layout"`
	Confidence      float64       `json:"confidence"`
	UsageCount      int           `json:"usage_count"`
	SuccessRate     float64       `json:"success_rate"`
	LastUsed        time.Time     `json:"last_used"`
}

// UIInteraction represents a user interaction with the UI
type UIInteraction struct {
	Type        string        `json:"type"`
	Target      string        `json:"target"`
	Timestamp   time.Time     `json:"timestamp"`
	Duration    time.Duration `json:"duration"`
	Successful  bool          `json:"successful"`
	Context     string        `json:"context"`
}

// UserUIPreferences represents learned user preferences
type UserUIPreferences struct {
	PreferredLayouts    map[ContentType]layoutMode `json:"preferred_layouts"`
	AutoSwitchEnabled   bool                       `json:"auto_switch_enabled"`
	TransitionSpeed     string                     `json:"transition_speed"`
	PanelPreferences    map[string]bool            `json:"panel_preferences"`
	ColorScheme         string                     `json:"color_scheme"`
	LastUpdated         time.Time                  `json:"last_updated"`
}

// AdaptationRule represents a rule for UI adaptation
type AdaptationRule struct {
	Name        string                    `json:"name"`
	Condition   func(*AdaptationContext) bool `json:"-"`
	Action      func(*AdaptationContext)      `json:"-"`
	Priority    int                       `json:"priority"`
	Enabled     bool                      `json:"enabled"`
	Description string                    `json:"description"`
}

// AdaptationContext provides context for adaptation decisions
type AdaptationContext struct {
	ContentType     ContentType
	ScreenSize      ScreenSize
	ActivityLevel   ActivityLevel
	CurrentLayout   layoutMode
	UserPreferences *UserUIPreferences
	Performance     *PerformanceMetrics
	Content         string
	Context         *KubeContextSummary
}

// YAMLAnalyzer analyzes YAML content
type YAMLAnalyzer struct {
	patterns map[string]float64
}

// LogAnalyzer analyzes log content
type LogAnalyzer struct {
	errorPatterns   []string
	warningPatterns []string
	infoPatterns    []string
}

// DiffAnalyzer analyzes diff content
type DiffAnalyzer struct {
	additionPatterns []string
	deletionPatterns []string
	modifyPatterns   []string
}

// CommandAnalyzer analyzes command content
type CommandAnalyzer struct {
	kubectlPatterns []string
	helmPatterns    []string
	bashPatterns    []string
}

// RiskAnalyzer analyzes content for risk indicators
type RiskAnalyzer struct {
	highRiskPatterns   []string
	mediumRiskPatterns []string
	lowRiskPatterns    []string
}

// PerformanceMetrics represents UI performance metrics
type PerformanceMetrics struct {
	AvgRenderTime     time.Duration `json:"avg_render_time"`
	AvgLayoutSwitch   time.Duration `json:"avg_layout_switch"`
	MemoryUsage       float64       `json:"memory_usage"`
	CPUUsage          float64       `json:"cpu_usage"`
	UserSatisfaction  float64       `json:"user_satisfaction"`
	ResponsivenessScore float64     `json:"responsiveness_score"`
}

// NewAdaptiveUIManager creates a new adaptive UI manager
func NewAdaptiveUIManager(streamingManager *StreamingIntelligenceManager) *AdaptiveUIManager {
	return &AdaptiveUIManager{
		streamingManager: streamingManager,
		layoutEngine: &AdaptiveLayoutEngine{
			currentLayout:       layoutThreePane,
			layoutHistory:       make([]LayoutTransition, 0),
			contentTypeRules:    make(map[ContentType]layoutMode),
			screenSizeRules:     make(map[ScreenSize]layoutMode),
			userPreferenceRules: make(map[string]layoutMode),
			autoSwitchEnabled:   true,
			transitionDuration:  200 * time.Millisecond,
		},
		contentAnalyzer: &ContentAnalyzer{
			yamlAnalyzer:    &YAMLAnalyzer{patterns: make(map[string]float64)},
			logAnalyzer:     &LogAnalyzer{},
			diffAnalyzer:    &DiffAnalyzer{},
			commandAnalyzer: &CommandAnalyzer{},
			riskAnalyzer:    &RiskAnalyzer{},
			patterns:        make(map[string]*ContentPattern),
		},
		userBehaviorTracker: &UserBehaviorTracker{
			activityLevel:      ActivityIdle,
			interactionHistory: make([]UIInteraction, 0),
			preferences: &UserUIPreferences{
				PreferredLayouts:  make(map[ContentType]layoutMode),
				AutoSwitchEnabled: true,
				TransitionSpeed:   "normal",
				PanelPreferences:  make(map[string]bool),
				ColorScheme:       "default",
				LastUpdated:       time.Now(),
			},
			adaptationRules: make([]AdaptationRule, 0),
			lastActivity:    time.Now(),
		},
		performanceMonitor: &PerformanceMonitor{
			renderTimes:         make([]time.Duration, 0),
			layoutSwitchTimes:   make([]time.Duration, 0),
			memoryUsage:         make([]float64, 0),
			cpuUsage:            make([]float64, 0),
			userSatisfaction:    0.8,
			optimizationEnabled: true,
		},
	}
}

// AdaptLayout adapts the layout based on content and context
func (aui *AdaptiveUIManager) AdaptLayout(content string, context *KubeContextSummary, currentLayout layoutMode) layoutMode {
	aui.mu.Lock()
	defer aui.mu.Unlock()

	// Analyze content
	contentType := aui.contentAnalyzer.AnalyzeContent(content)
	
	// Create adaptation context
	adaptationContext := &AdaptationContext{
		ContentType:     contentType,
		ScreenSize:      aui.determineScreenSize(),
		ActivityLevel:   aui.userBehaviorTracker.activityLevel,
		CurrentLayout:   currentLayout,
		UserPreferences: aui.userBehaviorTracker.preferences,
		Content:         content,
		Context:         context,
	}

	// Determine optimal layout
	optimalLayout := aui.layoutEngine.DetermineOptimalLayout(adaptationContext)

	// Check if layout change is needed
	if optimalLayout != currentLayout && aui.layoutEngine.autoSwitchEnabled {
		// Record transition
		transition := LayoutTransition{
			FromLayout: currentLayout,
			ToLayout:   optimalLayout,
			Trigger:    fmt.Sprintf("content_type_%d", contentType),
			Timestamp:  time.Now(),
			UserAction: false,
		}

		// Perform layout switch
		success := aui.layoutEngine.SwitchLayout(optimalLayout, transition)
		transition.Successful = success

		// Stream the layout change
		if aui.streamingManager != nil {
			aui.streamingManager.StreamUpdate(IntelligenceUpdate{
				Type:      "layout_change",
				Data:      map[string]interface{}{
					"from":    currentLayout,
					"to":      optimalLayout,
					"trigger": transition.Trigger,
				},
				Priority:  Medium,
				Timestamp: time.Now(),
			})
		}

		return optimalLayout
	}

	return currentLayout
}

// AnalyzeContent analyzes content to determine its type
func (ca *ContentAnalyzer) AnalyzeContent(content string) ContentType {
	ca.mu.RLock()
	defer ca.mu.RUnlock()

	contentLower := strings.ToLower(content)

	// Check for YAML content
	if ca.yamlAnalyzer.IsYAML(content) {
		return ContentTypeYAML
	}

	// Check for diff content
	if ca.diffAnalyzer.IsDiff(content) {
		return ContentTypeDiff
	}

	// Check for log content
	if ca.logAnalyzer.IsLog(content) {
		return ContentTypeLogs
	}

	// Check for command content
	if ca.commandAnalyzer.IsCommand(content) {
		return ContentTypeCommands
	}

	// Check for error content
	if strings.Contains(contentLower, "error") || strings.Contains(contentLower, "failed") {
		return ContentTypeError
	}

	// Check for help content
	if strings.Contains(contentLower, "help") || strings.Contains(contentLower, "usage") {
		return ContentTypeHelp
	}

	// Check for metrics content
	if strings.Contains(contentLower, "metrics") || strings.Contains(contentLower, "stats") {
		return ContentTypeMetrics
	}

	return ContentTypeGeneral
}

// DetermineOptimalLayout determines the optimal layout for given context
func (ale *AdaptiveLayoutEngine) DetermineOptimalLayout(ctx *AdaptationContext) layoutMode {
	ale.mu.RLock()
	defer ale.mu.RUnlock()

	// Check user preferences first
	if preferredLayout, exists := ctx.UserPreferences.PreferredLayouts[ctx.ContentType]; exists {
		return preferredLayout
	}

	// Check content type rules
	if layout, exists := ale.contentTypeRules[ctx.ContentType]; exists {
		return layout
	}

	// Apply default rules based on content type
	switch ctx.ContentType {
	case ContentTypeYAML, ContentTypeDiff:
		return layoutVerticalSplit // Better for wide content
	case ContentTypeLogs:
		return layoutHorizontalSplit // Better for scrolling logs
	case ContentTypeCommands:
		return layoutThreePane // Good for command + output + preview
	case ContentTypeError:
		return layoutChatOnly // Focus on error details
	case ContentTypeHelp:
		return layoutChatOnly // Focus on help content
	default:
		return layoutThreePane // Default layout
	}
}

// SwitchLayout performs a layout switch with transition tracking
func (ale *AdaptiveLayoutEngine) SwitchLayout(newLayout layoutMode, transition LayoutTransition) bool {
	ale.mu.Lock()
	defer ale.mu.Unlock()

	startTime := time.Now()
	
	// Perform the layout switch
	ale.currentLayout = newLayout
	
	// Record transition time
	transition.Duration = time.Since(startTime)
	ale.layoutHistory = append(ale.layoutHistory, transition)

	// Maintain history size
	if len(ale.layoutHistory) > 100 {
		ale.layoutHistory = ale.layoutHistory[len(ale.layoutHistory)-100:]
	}

	return true
}

// TrackInteraction tracks user interactions for behavior analysis
func (ubt *UserBehaviorTracker) TrackInteraction(interaction UIInteraction) {
	ubt.mu.Lock()
	defer ubt.mu.Unlock()

	ubt.interactionHistory = append(ubt.interactionHistory, interaction)
	ubt.lastActivity = time.Now()

	// Update activity level based on recent interactions
	ubt.updateActivityLevel()

	// Maintain history size
	if len(ubt.interactionHistory) > 1000 {
		ubt.interactionHistory = ubt.interactionHistory[len(ubt.interactionHistory)-1000:]
	}
}

// updateActivityLevel updates the current activity level
func (ubt *UserBehaviorTracker) updateActivityLevel() {
	now := time.Now()
	recentInteractions := 0

	// Count interactions in the last minute
	for _, interaction := range ubt.interactionHistory {
		if now.Sub(interaction.Timestamp) < time.Minute {
			recentInteractions++
		}
	}

	// Determine activity level
	switch {
	case recentInteractions >= 10:
		ubt.activityLevel = ActivityIntense
	case recentInteractions >= 5:
		ubt.activityLevel = ActivityModerate
	case recentInteractions >= 1:
		ubt.activityLevel = ActivityLight
	default:
		ubt.activityLevel = ActivityIdle
	}
}

// RecordPerformanceMetric records a performance metric
func (pm *PerformanceMonitor) RecordPerformanceMetric(metricType string, value interface{}) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	switch metricType {
	case "render_time":
		if duration, ok := value.(time.Duration); ok {
			pm.renderTimes = append(pm.renderTimes, duration)
			if len(pm.renderTimes) > 100 {
				pm.renderTimes = pm.renderTimes[len(pm.renderTimes)-100:]
			}
		}
	case "layout_switch_time":
		if duration, ok := value.(time.Duration); ok {
			pm.layoutSwitchTimes = append(pm.layoutSwitchTimes, duration)
			if len(pm.layoutSwitchTimes) > 100 {
				pm.layoutSwitchTimes = pm.layoutSwitchTimes[len(pm.layoutSwitchTimes)-100:]
			}
		}
	case "memory_usage":
		if usage, ok := value.(float64); ok {
			pm.memoryUsage = append(pm.memoryUsage, usage)
			if len(pm.memoryUsage) > 100 {
				pm.memoryUsage = pm.memoryUsage[len(pm.memoryUsage)-100:]
			}
		}
	case "cpu_usage":
		if usage, ok := value.(float64); ok {
			pm.cpuUsage = append(pm.cpuUsage, usage)
			if len(pm.cpuUsage) > 100 {
				pm.cpuUsage = pm.cpuUsage[len(pm.cpuUsage)-100:]
			}
		}
	}
}

// Helper methods for content analysis

func (ya *YAMLAnalyzer) IsYAML(content string) bool {
	yamlIndicators := []string{
		"apiVersion:",
		"kind:",
		"metadata:",
		"spec:",
		"---",
	}

	for _, indicator := range yamlIndicators {
		if strings.Contains(content, indicator) {
			return true
		}
	}
	return false
}

func (da *DiffAnalyzer) IsDiff(content string) bool {
	diffIndicators := []string{
		"@@",
		"+++",
		"---",
		"+++ ",
		"--- ",
	}

	for _, indicator := range diffIndicators {
		if strings.Contains(content, indicator) {
			return true
		}
	}
	return false
}

func (la *LogAnalyzer) IsLog(content string) bool {
	logIndicators := []string{
		"INFO",
		"WARN",
		"ERROR",
		"DEBUG",
		"FATAL",
		"timestamp",
		"level=",
	}

	contentUpper := strings.ToUpper(content)
	for _, indicator := range logIndicators {
		if strings.Contains(contentUpper, indicator) {
			return true
		}
	}
	return false
}

func (ca *CommandAnalyzer) IsCommand(content string) bool {
	commandIndicators := []string{
		"kubectl",
		"helm",
		"$ ",
		"# ",
		"bash",
		"sh",
	}

	for _, indicator := range commandIndicators {
		if strings.Contains(content, indicator) {
			return true
		}
	}
	return false
}

func (aui *AdaptiveUIManager) determineScreenSize() ScreenSize {
	// This would be determined from actual screen dimensions
	// For now, return a default
	return ScreenSizeMedium
}

// GetAdaptationStats returns adaptation statistics
func (aui *AdaptiveUIManager) GetAdaptationStats() map[string]interface{} {
	aui.mu.RLock()
	defer aui.mu.RUnlock()

	return map[string]interface{}{
		"current_layout":     aui.layoutEngine.currentLayout,
		"auto_switch_enabled": aui.layoutEngine.autoSwitchEnabled,
		"activity_level":     aui.userBehaviorTracker.activityLevel,
		"layout_transitions": len(aui.layoutEngine.layoutHistory),
		"user_interactions":  len(aui.userBehaviorTracker.interactionHistory),
		"content_patterns":   len(aui.contentAnalyzer.patterns),
		"user_satisfaction":  aui.performanceMonitor.userSatisfaction,
	}
}

// Global adaptive UI manager instance
var AdaptiveUI *AdaptiveUIManager

// InitializeAdaptiveUI initializes the global adaptive UI manager
func InitializeAdaptiveUI(streamingManager *StreamingIntelligenceManager) {
	AdaptiveUI = NewAdaptiveUIManager(streamingManager)
}

