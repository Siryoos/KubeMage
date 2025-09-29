// learning.go - Flight recorder and learning system for continuous improvement
package engine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// FlightRecorder logs interactions and outcomes for learning
type FlightRecorder struct {
	sessions     []SessionRecord
	storage      string
	maxSessions  int
	autoSave     bool
	lastSaved    time.Time
	learningData *LearningData
}

// SessionRecord captures a complete interaction session
type SessionRecord struct {
	ID                string              `json:"id"`
	Timestamp         time.Time           `json:"timestamp"`
	UserInput         string              `json:"user_input"`
	Context           *KubeContextSummary `json:"context"`
	IntentRouter      *IntentRouter       `json:"intent_router"`
	Observations      []ObservationRecord `json:"observations"`
	Suggestions       []SuggestionRecord  `json:"suggestions"`
	UserActions       []UserActionRecord  `json:"user_actions"`
	ValidationResults []ValidationRecord  `json:"validation_results"`
	Outcome           string              `json:"outcome"`           // "success", "failure", "partial"
	UserSatisfaction  int                 `json:"user_satisfaction"` // 1-5 rating
	Duration          time.Duration       `json:"duration"`
	ErrorMessages     []string            `json:"error_messages"`
}

type ObservationRecord struct {
	Command       string        `json:"command"`
	Output        string        `json:"output"`
	Error         string        `json:"error"`
	Timestamp     time.Time     `json:"timestamp"`
	Truncated     bool          `json:"truncated"`
	ExecutionTime time.Duration `json:"execution_time"`
}

type SuggestionRecord struct {
	Type       string    `json:"type"` // "command", "diff", "explanation"
	Content    string    `json:"content"`
	Confidence float64   `json:"confidence"`
	Risk       string    `json:"risk"`
	Accepted   bool      `json:"accepted"`
	Applied    bool      `json:"applied"`
	Timestamp  time.Time `json:"timestamp"`
}

type UserActionRecord struct {
	Action     string    `json:"action"` // "accept", "reject", "modify", "cancel"
	Target     string    `json:"target"` // What was acted upon
	Timestamp  time.Time `json:"timestamp"`
	ModifiedTo string    `json:"modified_to"` // If user modified the suggestion
}

type ValidationRecord struct {
	Command   string    `json:"command"`
	Result    string    `json:"result"` // "pass", "fail", "warning"
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Validator string    `json:"validator"` // "kubectl", "helm", "custom"
}

// LearningData aggregates patterns and insights from sessions
type LearningData struct {
	IntentAccuracy        map[AgentMode]float64  `json:"intent_accuracy"`
	CommonPatterns        []PatternLearning      `json:"common_patterns"`
	SuccessfulSuggestions []SuggestionPattern    `json:"successful_suggestions"`
	FailurePatterns       []FailurePattern       `json:"failure_patterns"`
	OptimizationImpact    []OptimizationLearning `json:"optimization_impact"`
	UserPreferences       UserPreferenceLearning `json:"user_preferences"`
	PerformanceMetrics    PerformanceMetrics     `json:"performance_metrics"`
	LastUpdated           time.Time              `json:"last_updated"`
}

type PatternLearning struct {
	InputPattern   string    `json:"input_pattern"`
	ContextPattern string    `json:"context_pattern"`
	SuccessfulMode AgentMode `json:"successful_mode"`
	Frequency      int       `json:"frequency"`
	SuccessRate    float64   `json:"success_rate"`
	AvgConfidence  float64   `json:"avg_confidence"`
}

type SuggestionPattern struct {
	InputType      string  `json:"input_type"`
	SuggestionType string  `json:"suggestion_type"`
	Pattern        string  `json:"pattern"`
	SuccessRate    float64 `json:"success_rate"`
	UserAcceptance float64 `json:"user_acceptance"`
	AvgConfidence  float64 `json:"avg_confidence"`
}

type FailurePattern struct {
	ErrorType       string   `json:"error_type"`
	Context         string   `json:"context"`
	Frequency       int      `json:"frequency"`
	CommonCauses    []string `json:"common_causes"`
	SuccessfulFixes []string `json:"successful_fixes"`
}

type OptimizationLearning struct {
	Type             string  `json:"type"`
	AppliedCount     int     `json:"applied_count"`
	SuccessRate      float64 `json:"success_rate"`
	AvgImprovement   float64 `json:"avg_improvement"`
	UserSatisfaction float64 `json:"user_satisfaction"`
}

type UserPreferenceLearning struct {
	PreferredModes      map[AgentMode]int `json:"preferred_modes"`
	RiskTolerance       string            `json:"risk_tolerance"`   // "low", "medium", "high"
	AutomationLevel     string            `json:"automation_level"` // "manual", "assisted", "automated"
	PreferredNamespaces []string          `json:"preferred_namespaces"`
	CommonCommands      []CommandPattern  `json:"common_commands"`
}

// CommandPattern is now defined in types.go

// PerformanceMetrics is now defined in types.go

// TrainingExample represents data suitable for model fine-tuning
type TrainingExample struct {
	Input    string                 `json:"input"`
	Context  map[string]string      `json:"context"`
	Output   string                 `json:"output"`
	Quality  float64                `json:"quality"` // 0.0-1.0 based on user feedback
	Type     string                 `json:"type"`    // "command", "diagnostic", "explanation", etc.
	Metadata map[string]interface{} `json:"metadata"`
}

// LoRADataset represents a dataset for LoRA fine-tuning
type LoRADataset struct {
	Examples   []TrainingExample `json:"examples"`
	Categories map[string]int    `json:"categories"`
	Quality    float64           `json:"quality"`
	CreatedAt  time.Time         `json:"created_at"`
	Version    string            `json:"version"`
}

// NewFlightRecorder creates a new flight recorder instance
func NewFlightRecorder(storagePath string) *FlightRecorder {
	return &FlightRecorder{
		sessions:    make([]SessionRecord, 0),
		storage:     storagePath,
		maxSessions: 1000, // Keep last 1000 sessions
		autoSave:    true,
		learningData: &LearningData{
			IntentAccuracy:        make(map[AgentMode]float64),
			CommonPatterns:        make([]PatternLearning, 0),
			SuccessfulSuggestions: make([]SuggestionPattern, 0),
			FailurePatterns:       make([]FailurePattern, 0),
			OptimizationImpact:    make([]OptimizationLearning, 0),
			UserPreferences: UserPreferenceLearning{
				PreferredModes: make(map[AgentMode]int),
				CommonCommands: make([]CommandPattern, 0),
			},
		},
	}
}

// RecordSession logs a complete interaction session
func (fr *FlightRecorder) RecordSession(session SessionRecord) error {
	session.Timestamp = time.Now()
	session.ID = fmt.Sprintf("session-%d", session.Timestamp.Unix())

	// Add to sessions
	fr.sessions = append(fr.sessions, session)

	// Maintain max sessions limit
	if len(fr.sessions) > fr.maxSessions {
		fr.sessions = fr.sessions[len(fr.sessions)-fr.maxSessions:]
	}

	// Update learning data
	fr.updateLearningData(session)

	// Auto-save if enabled
	if fr.autoSave && time.Since(fr.lastSaved) > 5*time.Minute {
		if err := fr.Save(); err != nil {
			return fmt.Errorf("failed to auto-save: %w", err)
		}
	}

	return nil
}

// updateLearningData extracts learning insights from new session
func (fr *FlightRecorder) updateLearningData(session SessionRecord) {
	// Update intent accuracy
	if session.IntentRouter != nil {
		mode := session.IntentRouter.Mode
		currentAccuracy := fr.learningData.IntentAccuracy[mode]

		// Simple accuracy calculation based on outcome
		sessionAccuracy := 0.0
		if session.Outcome == "success" {
			sessionAccuracy = 1.0
		} else if session.Outcome == "partial" {
			sessionAccuracy = 0.5
		}

		// Running average
		fr.learningData.IntentAccuracy[mode] = (currentAccuracy + sessionAccuracy) / 2.0
	}

	// Update user preferences
	if session.IntentRouter != nil {
		fr.learningData.UserPreferences.PreferredModes[session.IntentRouter.Mode]++
	}

	// Learn from successful suggestions
	for _, suggestion := range session.Suggestions {
		if suggestion.Accepted && suggestion.Applied {
			fr.learnFromSuccessfulSuggestion(session.UserInput, suggestion)
		}
	}

	// Learn from failures
	if session.Outcome == "failure" && len(session.ErrorMessages) > 0 {
		fr.learnFromFailure(session)
	}

	// Update performance metrics
	fr.updatePerformanceMetrics(session)

	fr.learningData.LastUpdated = time.Now()
}

// learnFromSuccessfulSuggestion extracts patterns from successful interactions
func (fr *FlightRecorder) learnFromSuccessfulSuggestion(input string, suggestion SuggestionRecord) {
	pattern := SuggestionPattern{
		InputType:      fr.categorizeInput(input),
		SuggestionType: suggestion.Type,
		Pattern:        fr.extractPattern(input),
		SuccessRate:    1.0, // Will be averaged over time
		UserAcceptance: 1.0,
		AvgConfidence:  suggestion.Confidence,
	}

	// Check if we already have this pattern
	for i, existing := range fr.learningData.SuccessfulSuggestions {
		if existing.InputType == pattern.InputType && existing.SuggestionType == pattern.SuggestionType {
			// Update existing pattern
			fr.learningData.SuccessfulSuggestions[i].SuccessRate = (existing.SuccessRate + 1.0) / 2.0
			return
		}
	}

	// Add new pattern
	fr.learningData.SuccessfulSuggestions = append(fr.learningData.SuccessfulSuggestions, pattern)
}

// learnFromFailure extracts patterns from failed interactions
func (fr *FlightRecorder) learnFromFailure(session SessionRecord) {
	errorType := "unknown"
	if len(session.ErrorMessages) > 0 {
		errorType = fr.categorizeError(session.ErrorMessages[0])
	}

	contextStr := ""
	if session.Context != nil {
		contextStr = fmt.Sprintf("ns:%s", session.Context.Namespace)
	}

	// Check if we have this failure pattern
	for i, existing := range fr.learningData.FailurePatterns {
		if existing.ErrorType == errorType && existing.Context == contextStr {
			fr.learningData.FailurePatterns[i].Frequency++
			return
		}
	}

	// Add new failure pattern
	pattern := FailurePattern{
		ErrorType:    errorType,
		Context:      contextStr,
		Frequency:    1,
		CommonCauses: []string{},
	}

	fr.learningData.FailurePatterns = append(fr.learningData.FailurePatterns, pattern)
}

// updatePerformanceMetrics updates aggregate performance statistics
func (fr *FlightRecorder) updatePerformanceMetrics(session SessionRecord) {
	metrics := &fr.learningData.PerformanceMetrics

	// Update average response time (simplified)
	if session.Duration > 0 {
		if metrics.AvgResponseTime == 0 {
			metrics.AvgResponseTime = session.Duration
		} else {
			metrics.AvgResponseTime = (metrics.AvgResponseTime + session.Duration) / 2
		}
	}

	// Update command accuracy
	successfulValidations := 0
	totalValidations := len(session.ValidationResults)
	for _, validation := range session.ValidationResults {
		if validation.Result == "pass" {
			successfulValidations++
		}
	}

	if totalValidations > 0 {
		sessionAccuracy := float64(successfulValidations) / float64(totalValidations)
		if metrics.CommandAccuracy == 0 {
			metrics.CommandAccuracy = sessionAccuracy
		} else {
			metrics.CommandAccuracy = (metrics.CommandAccuracy + sessionAccuracy) / 2
		}
	}
}

// ExportTrainingData generates training examples from recorded sessions
func (fr *FlightRecorder) ExportTrainingData() []TrainingExample {
	var examples []TrainingExample

	for _, session := range fr.sessions {
		// Only export successful interactions
		if session.Outcome != "success" {
			continue
		}

		for _, suggestion := range session.Suggestions {
			if suggestion.Accepted && suggestion.Applied {
				example := TrainingExample{
					Input:   session.UserInput,
					Output:  suggestion.Content,
					Quality: fr.calculateQuality(session, suggestion),
					Type:    suggestion.Type,
					Context: fr.buildContext(session.Context),
					Metadata: map[string]interface{}{
						"confidence":        suggestion.Confidence,
						"user_satisfaction": session.UserSatisfaction,
						"duration":          session.Duration.Seconds(),
					},
				}
				examples = append(examples, example)
			}
		}
	}

	// Sort by quality (highest first)
	sort.Slice(examples, func(i, j int) bool {
		return examples[i].Quality > examples[j].Quality
	})

	return examples
}

// GenerateLoRATrainingSet creates a dataset optimized for LoRA fine-tuning
func (fr *FlightRecorder) GenerateLoRATrainingSet() *LoRADataset {
	examples := fr.ExportTrainingData()

	// Filter high-quality examples
	highQualityExamples := make([]TrainingExample, 0)
	qualityThreshold := 0.7

	categories := make(map[string]int)
	totalQuality := 0.0

	for _, example := range examples {
		if example.Quality >= qualityThreshold {
			highQualityExamples = append(highQualityExamples, example)
			categories[example.Type]++
			totalQuality += example.Quality
		}
	}

	avgQuality := 0.0
	if len(highQualityExamples) > 0 {
		avgQuality = totalQuality / float64(len(highQualityExamples))
	}

	return &LoRADataset{
		Examples:   highQualityExamples,
		Categories: categories,
		Quality:    avgQuality,
		CreatedAt:  time.Now(),
		Version:    "1.0",
	}
}

// Save persists flight recorder data to disk
func (fr *FlightRecorder) Save() error {
	if err := os.MkdirAll(fr.storage, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Save sessions
	sessionsFile := filepath.Join(fr.storage, "sessions.json")
	if data, err := json.MarshalIndent(fr.sessions, "", "  "); err != nil {
		return fmt.Errorf("failed to marshal sessions: %w", err)
	} else if err := ioutil.WriteFile(sessionsFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write sessions file: %w", err)
	}

	// Save learning data
	learningFile := filepath.Join(fr.storage, "learning_data.json")
	if data, err := json.MarshalIndent(fr.learningData, "", "  "); err != nil {
		return fmt.Errorf("failed to marshal learning data: %w", err)
	} else if err := ioutil.WriteFile(learningFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write learning data file: %w", err)
	}

	fr.lastSaved = time.Now()
	return nil
}

// Load restores flight recorder data from disk
func (fr *FlightRecorder) Load() error {
	// Load sessions
	sessionsFile := filepath.Join(fr.storage, "sessions.json")
	if data, err := ioutil.ReadFile(sessionsFile); err == nil {
		if err := json.Unmarshal(data, &fr.sessions); err != nil {
			return fmt.Errorf("failed to unmarshal sessions: %w", err)
		}
	}

	// Load learning data
	learningFile := filepath.Join(fr.storage, "learning_data.json")
	if data, err := ioutil.ReadFile(learningFile); err == nil {
		if err := json.Unmarshal(data, &fr.learningData); err != nil {
			return fmt.Errorf("failed to unmarshal learning data: %w", err)
		}
	}

	return nil
}

// Helper methods
func (fr *FlightRecorder) categorizeInput(input string) string {
	inputLower := strings.ToLower(input)

	if strings.Contains(inputLower, "debug") || strings.Contains(inputLower, "error") || strings.Contains(inputLower, "failing") {
		return "diagnostic"
	} else if strings.Contains(inputLower, "create") || strings.Contains(inputLower, "generate") {
		return "creation"
	} else if strings.Contains(inputLower, "edit") || strings.Contains(inputLower, "modify") || strings.Contains(inputLower, "change") {
		return "modification"
	} else if strings.Contains(inputLower, "explain") || strings.Contains(inputLower, "what") || strings.Contains(inputLower, "how") {
		return "explanation"
	} else {
		return "command"
	}
}

func (fr *FlightRecorder) extractPattern(input string) string {
	// Simplified pattern extraction - replace specific names with placeholders
	pattern := input
	pattern = regexp.MustCompile(`\b[a-zA-Z0-9-]+\b`).ReplaceAllString(pattern, "{name}")
	return strings.ToLower(pattern)
}

func (fr *FlightRecorder) categorizeError(error string) string {
	errorLower := strings.ToLower(error)

	if strings.Contains(errorLower, "not found") || strings.Contains(errorLower, "does not exist") {
		return "not-found"
	} else if strings.Contains(errorLower, "permission") || strings.Contains(errorLower, "forbidden") {
		return "permission"
	} else if strings.Contains(errorLower, "timeout") || strings.Contains(errorLower, "deadline") {
		return "timeout"
	} else if strings.Contains(errorLower, "syntax") || strings.Contains(errorLower, "invalid") {
		return "syntax"
	} else {
		return "unknown"
	}
}

func (fr *FlightRecorder) calculateQuality(session SessionRecord, suggestion SuggestionRecord) float64 {
	quality := suggestion.Confidence

	// Boost quality for user satisfaction
	if session.UserSatisfaction > 3 {
		quality += 0.2
	}

	// Boost quality for quick resolution
	if session.Duration < 2*time.Minute {
		quality += 0.1
	}

	// Reduce quality for validation failures
	for _, validation := range session.ValidationResults {
		if validation.Result == "fail" {
			quality -= 0.1
		}
	}

	return min(quality, 1.0)
}

func (fr *FlightRecorder) buildContext(context *KubeContextSummary) map[string]string {
	result := make(map[string]string)

	if context != nil {
		result["namespace"] = context.Namespace
		result["context"] = context.Context

		if len(context.PodProblemCounts) > 0 {
			result["has_problems"] = "true"
		} else {
			result["has_problems"] = "false"
		}
	}

	return result
}

// GetLearningInsights provides insights from accumulated learning data
func (fr *FlightRecorder) GetLearningInsights() *LearningData {
	return fr.learningData
}

// Global flight recorder instance
var Recorder = NewFlightRecorder("./kubemage_data")
