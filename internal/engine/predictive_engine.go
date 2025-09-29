// predictive_engine.go - Core predictive intelligence engine implementation
package engine

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// NewPredictiveIntelligenceEngine creates a new predictive intelligence engine
func NewPredictiveIntelligenceEngine(smartCache *SmartCacheSystem, streamingManager *StreamingIntelligenceManager) *PredictiveIntelligenceEngine {
	return &PredictiveIntelligenceEngine{
		patternLearner: &PatternLearner{
			userPatterns:        make(map[string]*UserPattern),
			clusterPatterns:     make(map[string]*ClusterPattern),
			commandPatterns:     make(map[string]*CommandPattern),
			sessionPatterns:     make(map[string]*SessionPattern),
			learningRate:        0.1,
			confidenceThreshold: 0.7,
			maxPatterns:         1000,
		},
		contextPredictor: &ContextPredictor{
			contextHistory:   make([]ContextSnapshot, 0),
			transitionMatrix: make(map[string]map[string]float64),
			predictionWindow: 5 * time.Minute,
			accuracy:         0.8,
		},
		behaviorAnalyzer: &BehaviorAnalyzer{
			behaviorHistory: make([]BehaviorEvent, 0),
			patterns:        make(map[string]*BehaviorPattern),
			anomalyDetector: &AnomalyDetector{
				baselineMetrics: make(map[string]float64),
				thresholds:      make(map[string]float64),
				anomalyHistory:  make([]AnomalyEvent, 0),
				sensitivity:     0.8,
			},
			sessionTracker: &SessionTracker{
				sessionHistory:  make([]UserSession, 0),
				sessionPatterns: make(map[string]*SessionPattern),
				maxSessions:     100,
			},
		},
		smartCache:          smartCache,
		streamingManager:    streamingManager,
		confidenceThreshold: 0.75,
		learningRate:        0.1,
	}
}

// PredictNextActions predicts likely next user actions based on current context
func (pie *PredictiveIntelligenceEngine) PredictNextActions(context *KubeContextSummary, userInput string) []PredictedAction {
	pie.mu.RLock()
	defer pie.mu.RUnlock()

	// Check cache first
	cacheKey := fmt.Sprintf("predictions:%s:%s", context.Hash(), userInput)
	if cached := pie.smartCache.GetL1(cacheKey); cached != nil {
		return cached.([]PredictedAction)
	}

	var predictions []PredictedAction

	// Analyze current context and user behavior
	behaviorPredictions := pie.behaviorAnalyzer.PredictNextActions(userInput, context)
	contextPredictions := pie.contextPredictor.PredictContextChanges(context)
	patternPredictions := pie.patternLearner.PredictFromPatterns(userInput, context)

	// Combine predictions with weighted scoring
	predictions = append(predictions, behaviorPredictions...)
	predictions = append(predictions, contextPredictions...)
	predictions = append(predictions, patternPredictions...)

	// Sort by confidence and priority
	sort.Slice(predictions, func(i, j int) bool {
		if predictions[i].Priority != predictions[j].Priority {
			return predictions[i].Priority > predictions[j].Priority
		}
		return predictions[i].Confidence > predictions[j].Confidence
	})

	// Filter by confidence threshold
	var filteredPredictions []PredictedAction
	for _, pred := range predictions {
		if pred.Confidence >= pie.confidenceThreshold {
			filteredPredictions = append(filteredPredictions, pred)
		}
	}

	// Limit to top 5 predictions
	if len(filteredPredictions) > 5 {
		filteredPredictions = filteredPredictions[:5]
	}

	// Cache the results
	pie.smartCache.SetL1(cacheKey, filteredPredictions, 2*time.Minute)

	return filteredPredictions
}

// LearnFromInteraction learns from user interactions to improve predictions
func (pie *PredictiveIntelligenceEngine) LearnFromInteraction(userInput string, context *KubeContextSummary, action string, success bool, duration time.Duration) {
	pie.mu.Lock()
	defer pie.mu.Unlock()

	// Record behavior event
	event := BehaviorEvent{
		Timestamp: time.Now(),
		EventType: "user_interaction",
		Context:   context.Namespace,
		Action:    action,
		Duration:  duration,
		Success:   success,
		Metadata: map[string]interface{}{
			"user_input": userInput,
			"context":    context,
		},
	}

	pie.behaviorAnalyzer.RecordEvent(event)

	// Update patterns
	pie.patternLearner.UpdatePatterns(userInput, context, action, success)

	// Update context predictions
	pie.contextPredictor.UpdateTransitions(context)

	// Stream learning update
	if pie.streamingManager != nil {
		pie.streamingManager.StreamUpdate(IntelligenceUpdate{
			Type:      "learning_update",
			Data:      map[string]interface{}{"action": action, "success": success},
			Priority:  Medium,
			Timestamp: time.Now(),
		})
	}
}

// GetPredictionAccuracy returns the current prediction accuracy
func (pie *PredictiveIntelligenceEngine) GetPredictionAccuracy() float64 {
	pie.mu.RLock()
	defer pie.mu.RUnlock()

	return pie.contextPredictor.accuracy
}

// GetLearningStats returns learning statistics
func (pie *PredictiveIntelligenceEngine) GetLearningStats() map[string]interface{} {
	pie.mu.RLock()
	defer pie.mu.RUnlock()

	return map[string]interface{}{
		"user_patterns":    len(pie.patternLearner.userPatterns),
		"cluster_patterns": len(pie.patternLearner.clusterPatterns),
		"command_patterns": len(pie.patternLearner.commandPatterns),
		"behavior_events":  len(pie.behaviorAnalyzer.behaviorHistory),
		"behavior_patterns": len(pie.behaviorAnalyzer.patterns),
		"context_history":  len(pie.contextPredictor.contextHistory),
		"prediction_accuracy": pie.contextPredictor.accuracy,
		"confidence_threshold": pie.confidenceThreshold,
	}
}

// OptimizePredictions optimizes prediction algorithms based on performance
func (pie *PredictiveIntelligenceEngine) OptimizePredictions() {
	pie.mu.Lock()
	defer pie.mu.Unlock()

	// Adjust confidence threshold based on accuracy
	if pie.contextPredictor.accuracy > 0.9 {
		pie.confidenceThreshold = math.Max(0.6, pie.confidenceThreshold-0.05)
	} else if pie.contextPredictor.accuracy < 0.7 {
		pie.confidenceThreshold = math.Min(0.9, pie.confidenceThreshold+0.05)
	}

	// Clean up old patterns
	pie.patternLearner.CleanupOldPatterns()
	pie.behaviorAnalyzer.CleanupOldEvents()
	pie.contextPredictor.CleanupOldHistory()
}

// PredictFromPatterns predicts actions based on learned patterns
func (pl *PatternLearner) PredictFromPatterns(userInput string, context *KubeContextSummary) []PredictedAction {
	pl.mu.RLock()
	defer pl.mu.RUnlock()

	var predictions []PredictedAction

	// Check user patterns
	for _, userPattern := range pl.userPatterns {
		for _, queryPattern := range userPattern.CommonQueries {
			if pl.matchesPattern(userInput, queryPattern.Pattern) {
				for _, actionPattern := range userPattern.PreferredActions {
					if pl.contextMatches(context, actionPattern.Context) {
						predictions = append(predictions, PredictedAction{
							Action:        actionPattern.Action,
							Confidence:    queryPattern.SuccessRate * actionPattern.SuccessRate,
							Priority:      pl.calculatePriority(actionPattern),
							Context:       actionPattern.Context,
							EstimatedTime: time.Duration(actionPattern.Frequency) * time.Second,
							RiskLevel:     actionPattern.RiskLevel,
						})
					}
				}
			}
		}
	}

	// Check command patterns
	for _, cmdPattern := range pl.commandPatterns {
		if strings.Contains(userInput, cmdPattern.Command) {
			for _, followUp := range cmdPattern.CommonFollowUps {
				predictions = append(predictions, PredictedAction{
					Action:        followUp,
					Confidence:    cmdPattern.SuccessRate * 0.8,
					Priority:      2,
					Context:       cmdPattern.Context,
					EstimatedTime: cmdPattern.AverageTime,
					RiskLevel:     "low",
				})
			}
		}
	}

	return predictions
}

// UpdatePatterns updates learned patterns based on new interactions
func (pl *PatternLearner) UpdatePatterns(userInput string, context *KubeContextSummary, action string, success bool) {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	userID := "default" // In a real implementation, this would be the actual user ID

	// Update user patterns
	if userPattern, exists := pl.userPatterns[userID]; exists {
		pl.updateUserPattern(userPattern, userInput, context, action, success)
	} else {
		pl.userPatterns[userID] = pl.createUserPattern(userID, userInput, context, action, success)
	}

	// Update command patterns
	cmdKey := fmt.Sprintf("%s:%s", action, context.Namespace)
	if cmdPattern, exists := pl.commandPatterns[cmdKey]; exists {
		pl.updateCommandPattern(cmdPattern, success)
	} else {
		pl.commandPatterns[cmdKey] = pl.createCommandPattern(action, context, success)
	}

	// Update cluster patterns
	clusterKey := context.Context
	if clusterPattern, exists := pl.clusterPatterns[clusterKey]; exists {
		pl.updateClusterPattern(clusterPattern, context, action, success)
	} else {
		pl.clusterPatterns[clusterKey] = pl.createClusterPattern(context, action, success)
	}
}

// CleanupOldPatterns removes old or low-confidence patterns
func (pl *PatternLearner) CleanupOldPatterns() {
	cutoff := time.Now().Add(-24 * time.Hour)

	// Clean user patterns
	for id, pattern := range pl.userPatterns {
		if pattern.LastUpdated.Before(cutoff) && pattern.Confidence < 0.3 {
			delete(pl.userPatterns, id)
		}
	}

	// Clean command patterns
	for id, pattern := range pl.commandPatterns {
		if pattern.LastUsed.Before(cutoff) && pattern.Confidence < 0.3 {
			delete(pl.commandPatterns, id)
		}
	}

	// Clean cluster patterns
	for id, pattern := range pl.clusterPatterns {
		if pattern.LastUpdated.Before(cutoff) && pattern.Confidence < 0.3 {
			delete(pl.clusterPatterns, id)
		}
	}
}

// Helper methods for pattern matching and updates
func (pl *PatternLearner) matchesPattern(input, pattern string) bool {
	// Simple pattern matching - in a real implementation, this would be more sophisticated
	return strings.Contains(strings.ToLower(input), strings.ToLower(pattern))
}

func (pl *PatternLearner) contextMatches(context *KubeContextSummary, pattern string) bool {
	return strings.Contains(context.Namespace, pattern) || strings.Contains(context.Context, pattern)
}

func (pl *PatternLearner) calculatePriority(actionPattern ActionPattern) int {
	switch actionPattern.RiskLevel {
	case "low":
		return 3
	case "medium":
		return 2
	case "high":
		return 1
	default:
		return 2
	}
}

func (pl *PatternLearner) updateUserPattern(pattern *UserPattern, userInput string, context *KubeContextSummary, action string, success bool) {
	// Update success rate using exponential moving average
	if success {
		pattern.SuccessRate = pattern.SuccessRate*0.9 + 0.1
	} else {
		pattern.SuccessRate = pattern.SuccessRate * 0.9
	}

	pattern.LastUpdated = time.Now()
	pattern.Frequency++

	// Update confidence based on frequency and success rate
	pattern.Confidence = math.Min(1.0, float64(pattern.Frequency)/100.0*pattern.SuccessRate)
}

func (pl *PatternLearner) createUserPattern(userID, userInput string, context *KubeContextSummary, action string, success bool) *UserPattern {
	successRate := 0.5
	if success {
		successRate = 1.0
	}

	return &UserPattern{
		UserID: userID,
		CommonQueries: []QueryPattern{{
			Pattern:     userInput,
			Intent:      "unknown",
			Frequency:   1,
			SuccessRate: successRate,
		}},
		PreferredActions: []ActionPattern{{
			Action:      action,
			Context:     context.Namespace,
			Frequency:   1,
			SuccessRate: successRate,
			RiskLevel:   "medium",
		}},
		RiskTolerance: "medium",
		SuccessRate:   successRate,
		LastUpdated:   time.Now(),
		Frequency:     1,
		Confidence:    0.1,
	}
}

func (pl *PatternLearner) updateCommandPattern(pattern *CommandPattern, success bool) {
	// Update success rate
	if success {
		pattern.SuccessRate = pattern.SuccessRate*0.9 + 0.1
	} else {
		pattern.SuccessRate = pattern.SuccessRate * 0.9
	}

	pattern.Frequency++
	pattern.LastUsed = time.Now()
	pattern.Confidence = math.Min(1.0, float64(pattern.Frequency)/50.0*pattern.SuccessRate)
}

func (pl *PatternLearner) createCommandPattern(action string, context *KubeContextSummary, success bool) *CommandPattern {
	successRate := 0.5
	if success {
		successRate = 1.0
	}

	return &CommandPattern{
		Command:     action,
		Context:     context.Namespace,
		Frequency:   1,
		SuccessRate: successRate,
		AverageTime: 2 * time.Second,
		LastUsed:    time.Now(),
		Confidence:  0.1,
	}
}

func (pl *PatternLearner) updateClusterPattern(pattern *ClusterPattern, context *KubeContextSummary, action string, success bool) {
	pattern.LastUpdated = time.Now()
	
	// Update namespace patterns
	if nsPattern, exists := pattern.NamespacePatterns[context.Namespace]; exists {
		nsPattern.LastAnalyzed = time.Now()
		if success {
			nsPattern.HealthScore = math.Min(1.0, nsPattern.HealthScore+0.01)
		} else {
			nsPattern.HealthScore = math.Max(0.0, nsPattern.HealthScore-0.02)
		}
	}
}

func (pl *PatternLearner) createClusterPattern(context *KubeContextSummary, action string, success bool) *ClusterPattern {
	healthScore := 0.5
	if success {
		healthScore = 0.7
	} else {
		healthScore = 0.3
	}

	return &ClusterPattern{
		ClusterID: context.Context,
		NamespacePatterns: map[string]*NamespacePattern{
			context.Namespace: {
				Namespace:    context.Namespace,
				CommonIssues: []string{},
				ResourceUsage: make(map[string]float64),
				HealthScore:  healthScore,
				LastAnalyzed: time.Now(),
			},
		},
		ResourcePatterns:  make(map[string]*ResourcePattern),
		ProblemPatterns:   []*ProblemPattern{},
		OptimizationHints: []*OptimizationHint{},
		HealthPatterns:    []*HealthPattern{},
		LastUpdated:       time.Now(),
		Confidence:        0.1,
	}
}

// Removed global instance - now created via dependency injection

// InitializePredictiveIntelligence creates a new predictive intelligence engine
func InitializePredictiveIntelligence(smartCache *SmartCacheSystem, streamingManager *StreamingIntelligenceManager) *PredictiveIntelligenceEngine {
	return NewPredictiveIntelligenceEngine(smartCache, streamingManager)
}

