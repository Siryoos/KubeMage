// predictive_types.go - Type definitions for predictive intelligence engine
package main

import (
	"sync"
	"time"
)

// PredictiveIntelligenceEngine provides advanced predictive capabilities
type PredictiveIntelligenceEngine struct {
	patternLearner      *PatternLearner
	contextPredictor    *ContextPredictor
	behaviorAnalyzer    *BehaviorAnalyzer
	smartCache          *SmartCache
	streamingManager    *StreamingIntelligenceManager
	confidenceThreshold float64
	learningRate        float64
	mu                  sync.RWMutex
}

// PatternLearner learns from user interactions and cluster patterns
type PatternLearner struct {
	userPatterns        map[string]*UserPattern
	clusterPatterns     map[string]*ClusterPattern
	commandPatterns     map[string]*CommandPattern
	sessionPatterns     map[string]*SessionPattern
	learningRate        float64
	confidenceThreshold float64
	maxPatterns         int
	mu                  sync.RWMutex
}

// UserPattern represents learned user behavior patterns
type UserPattern struct {
	UserID            string              `json:"user_id"`
	CommonQueries     []QueryPattern      `json:"common_queries"`
	PreferredActions  []ActionPattern     `json:"preferred_actions"`
	RiskTolerance     string              `json:"risk_tolerance"`
	WorkingHours      []TimePattern       `json:"working_hours"`
	ContextSwitches   []ContextSwitch     `json:"context_switches"`
	SuccessRate       float64             `json:"success_rate"`
	LastUpdated       time.Time           `json:"last_updated"`
	Frequency         int                 `json:"frequency"`
	Confidence        float64             `json:"confidence"`
}

// ClusterPattern represents learned cluster behavior patterns
type ClusterPattern struct {
	ClusterID         string                         `json:"cluster_id"`
	NamespacePatterns map[string]*NamespacePattern   `json:"namespace_patterns"`
	ResourcePatterns  map[string]*ResourcePattern    `json:"resource_patterns"`
	ProblemPatterns   []*ProblemPattern              `json:"problem_patterns"`
	OptimizationHints []*OptimizationHint            `json:"optimization_hints"`
	HealthPatterns    []*HealthPattern               `json:"health_patterns"`
	LastUpdated       time.Time                      `json:"last_updated"`
	Confidence        float64                        `json:"confidence"`
}

// CommandPattern represents learned command usage patterns
type CommandPattern struct {
	Command           string            `json:"command"`
	Context           string            `json:"context"`
	Frequency         int               `json:"frequency"`
	SuccessRate       float64           `json:"success_rate"`
	AverageTime       time.Duration     `json:"average_time"`
	CommonFollowUps   []string          `json:"common_follow_ups"`
	ErrorPatterns     []string          `json:"error_patterns"`
	OptimizedVersion  string            `json:"optimized_version"`
	LastUsed          time.Time         `json:"last_used"`
	Confidence        float64           `json:"confidence"`
}

// SessionPattern represents learned session flow patterns
type SessionPattern struct {
	SessionType       string            `json:"session_type"`
	CommonFlow        []string          `json:"common_flow"`
	Duration          time.Duration     `json:"duration"`
	SuccessRate       float64           `json:"success_rate"`
	CriticalPoints    []string          `json:"critical_points"`
	OptimizationTips  []string          `json:"optimization_tips"`
	LastSeen          time.Time         `json:"last_seen"`
	Frequency         int               `json:"frequency"`
}

// QueryPattern represents a learned query pattern
type QueryPattern struct {
	Pattern           string            `json:"pattern"`
	Intent            string            `json:"intent"`
	Frequency         int               `json:"frequency"`
	SuccessRate       float64           `json:"success_rate"`
	AverageConfidence float64           `json:"average_confidence"`
	CommonContext     []string          `json:"common_context"`
}

// ActionPattern represents a learned action pattern
type ActionPattern struct {
	Action            string            `json:"action"`
	Context           string            `json:"context"`
	Frequency         int               `json:"frequency"`
	SuccessRate       float64           `json:"success_rate"`
	RiskLevel         string            `json:"risk_level"`
	PreferredTime     []time.Time       `json:"preferred_time"`
}

// TimePattern represents time-based usage patterns
type TimePattern struct {
	Hour              int               `json:"hour"`
	DayOfWeek         int               `json:"day_of_week"`
	ActivityLevel     string            `json:"activity_level"`
	CommonActions     []string          `json:"common_actions"`
	Frequency         int               `json:"frequency"`
}

// ContextSwitch represents context switching patterns
type ContextSwitch struct {
	FromContext       string            `json:"from_context"`
	ToContext         string            `json:"to_context"`
	Trigger           string            `json:"trigger"`
	Frequency         int               `json:"frequency"`
	Duration          time.Duration     `json:"duration"`
}

// NamespacePattern represents namespace-specific patterns
type NamespacePattern struct {
	Namespace         string            `json:"namespace"`
	CommonIssues      []string          `json:"common_issues"`
	ResourceUsage     map[string]float64 `json:"resource_usage"`
	DeploymentPatterns []string          `json:"deployment_patterns"`
	HealthScore       float64           `json:"health_score"`
	LastAnalyzed      time.Time         `json:"last_analyzed"`
}

// ResourcePattern represents resource usage patterns
type ResourcePattern struct {
	ResourceType      string            `json:"resource_type"`
	UsagePattern      []float64         `json:"usage_pattern"`
	PeakTimes         []time.Time       `json:"peak_times"`
	OptimalConfig     map[string]string `json:"optimal_config"`
	PredictedGrowth   float64           `json:"predicted_growth"`
	LastUpdated       time.Time         `json:"last_updated"`
}

// ProblemPattern represents recurring problem patterns
type ProblemPattern struct {
	ProblemType       string            `json:"problem_type"`
	Symptoms          []string          `json:"symptoms"`
	RootCauses        []string          `json:"root_causes"`
	Solutions         []string          `json:"solutions"`
	Frequency         int               `json:"frequency"`
	Severity          string            `json:"severity"`
	PreventionTips    []string          `json:"prevention_tips"`
}

// OptimizationHint represents optimization suggestions
type OptimizationHint struct {
	Type              string            `json:"type"`
	Description       string            `json:"description"`
	Impact            string            `json:"impact"`
	Difficulty        string            `json:"difficulty"`
	EstimatedSavings  string            `json:"estimated_savings"`
	Implementation    []string          `json:"implementation"`
	Confidence        float64           `json:"confidence"`
}

// HealthPattern represents cluster health patterns
type HealthPattern struct {
	MetricName        string            `json:"metric_name"`
	NormalRange       [2]float64        `json:"normal_range"`
	WarningThreshold  float64           `json:"warning_threshold"`
	CriticalThreshold float64           `json:"critical_threshold"`
	TrendDirection    string            `json:"trend_direction"`
	Seasonality       []float64         `json:"seasonality"`
	LastUpdated       time.Time         `json:"last_updated"`
}

// ContextPredictor predicts future context changes
type ContextPredictor struct {
	contextHistory    []ContextSnapshot
	transitionMatrix  map[string]map[string]float64
	predictionWindow  time.Duration
	accuracy          float64
	mu                sync.RWMutex
}

// ContextSnapshot represents a point-in-time context state
type ContextSnapshot struct {
	Timestamp         time.Time         `json:"timestamp"`
	Context           *KubeContextSummary `json:"context"`
	UserActivity      string            `json:"user_activity"`
	SystemLoad        float64           `json:"system_load"`
	PredictedNext     []string          `json:"predicted_next"`
}

// BehaviorAnalyzer analyzes user behavior patterns
type BehaviorAnalyzer struct {
	behaviorHistory   []BehaviorEvent
	patterns          map[string]*BehaviorPattern
	anomalyDetector   *AnomalyDetector
	sessionTracker    *SessionTracker
	mu                sync.RWMutex
}

// BehaviorEvent represents a user behavior event
type BehaviorEvent struct {
	Timestamp         time.Time         `json:"timestamp"`
	EventType         string            `json:"event_type"`
	Context           string            `json:"context"`
	Action            string            `json:"action"`
	Duration          time.Duration     `json:"duration"`
	Success           bool              `json:"success"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// BehaviorPattern represents a learned behavior pattern
type BehaviorPattern struct {
	PatternID         string            `json:"pattern_id"`
	EventSequence     []string          `json:"event_sequence"`
	Frequency         int               `json:"frequency"`
	Confidence        float64           `json:"confidence"`
	PredictiveValue   float64           `json:"predictive_value"`
	NextActions       []PredictedAction `json:"next_actions"`
	LastSeen          time.Time         `json:"last_seen"`
}

// PredictedAction represents a predicted user action
type PredictedAction struct {
	Action            string            `json:"action"`
	Confidence        float64           `json:"confidence"`
	Priority          int               `json:"priority"`
	Context           string            `json:"context"`
	EstimatedTime     time.Duration     `json:"estimated_time"`
	Prerequisites     []string          `json:"prerequisites"`
	RiskLevel         string            `json:"risk_level"`
}

// AnomalyDetector detects unusual behavior patterns
type AnomalyDetector struct {
	baselineMetrics   map[string]float64
	thresholds        map[string]float64
	anomalyHistory    []AnomalyEvent
	sensitivity       float64
}

// AnomalyEvent represents a detected anomaly
type AnomalyEvent struct {
	Timestamp         time.Time         `json:"timestamp"`
	AnomalyType       string            `json:"anomaly_type"`
	Severity          string            `json:"severity"`
	Description       string            `json:"description"`
	Metrics           map[string]float64 `json:"metrics"`
	Resolved          bool              `json:"resolved"`
}

// SessionTracker tracks user session patterns
type SessionTracker struct {
	currentSession    *UserSession
	sessionHistory    []UserSession
	sessionPatterns   map[string]*SessionPattern
	maxSessions       int
}

// UserSession represents a user interaction session
type UserSession struct {
	SessionID         string            `json:"session_id"`
	StartTime         time.Time         `json:"start_time"`
	EndTime           time.Time         `json:"end_time"`
	Actions           []BehaviorEvent   `json:"actions"`
	Context           []ContextSnapshot `json:"context"`
	Outcome           string            `json:"outcome"`
	Satisfaction      float64           `json:"satisfaction"`
	Efficiency        float64           `json:"efficiency"`
}

