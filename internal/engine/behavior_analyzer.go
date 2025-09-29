// behavior_analyzer.go - Behavior analysis and anomaly detection for predictive intelligence
package engine

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// PredictNextActions predicts next actions based on behavior analysis
func (ba *BehaviorAnalyzer) PredictNextActions(userInput string, context *KubeContextSummary) []PredictedAction {
	ba.mu.RLock()
	defer ba.mu.RUnlock()

	var predictions []PredictedAction

	// Analyze recent behavior patterns
	recentEvents := ba.getRecentEvents(10)
	if len(recentEvents) < 2 {
		return predictions
	}

	// Find matching behavior patterns
	for _, pattern := range ba.patterns {
		if ba.matchesEventSequence(recentEvents, pattern.EventSequence) {
			for _, nextAction := range pattern.NextActions {
				predictions = append(predictions, nextAction)
			}
		}
	}

	// Check for anomalies
	if ba.anomalyDetector.DetectAnomaly(userInput, context) {
		// Add conservative predictions for anomalous behavior
		predictions = append(predictions, PredictedAction{
			Action:     "kubectl get pods",
			Confidence: 0.6,
			Priority:   1,
			Context:    "diagnostic",
			RiskLevel:  "low",
		})
	}

	// Add context-based predictions
	contextPredictions := ba.predictFromContext(context, userInput)
	predictions = append(predictions, contextPredictions...)

	return predictions
}

// RecordEvent records a behavior event and updates patterns
func (ba *BehaviorAnalyzer) RecordEvent(event BehaviorEvent) {
	ba.mu.Lock()
	defer ba.mu.Unlock()

	ba.behaviorHistory = append(ba.behaviorHistory, event)

	// Maintain history size
	if len(ba.behaviorHistory) > 1000 {
		ba.behaviorHistory = ba.behaviorHistory[len(ba.behaviorHistory)-1000:]
	}

	// Update patterns
	ba.updateBehaviorPatterns(event)

	// Update session tracking
	ba.sessionTracker.UpdateSession(event)

	// Update anomaly detection baselines
	ba.anomalyDetector.UpdateBaselines(event)
}

// updateBehaviorPatterns updates behavior patterns based on new events
func (ba *BehaviorAnalyzer) updateBehaviorPatterns(event BehaviorEvent) {
	// Extract patterns from recent events
	recentEvents := ba.getRecentEvents(5)
	if len(recentEvents) < 3 {
		return
	}

	// Create pattern from event sequence
	sequence := make([]string, len(recentEvents))
	for i, e := range recentEvents {
		sequence[i] = e.Action
	}

	patternID := strings.Join(sequence, "->")
	if pattern, exists := ba.patterns[patternID]; exists {
		pattern.Frequency++
		pattern.LastSeen = time.Now()
		pattern.Confidence = math.Min(1.0, float64(pattern.Frequency)/20.0)
		
		// Update predictive value based on success rate
		if event.Success {
			pattern.PredictiveValue = pattern.PredictiveValue*0.9 + 0.1
		} else {
			pattern.PredictiveValue = pattern.PredictiveValue * 0.9
		}
	} else {
		ba.patterns[patternID] = &BehaviorPattern{
			PatternID:       patternID,
			EventSequence:   sequence,
			Frequency:       1,
			Confidence:      0.05,
			PredictiveValue: 0.5,
			LastSeen:        time.Now(),
			NextActions:     ba.generateNextActions(sequence, event),
		}
	}
}

// generateNextActions generates predicted next actions based on event sequence
func (ba *BehaviorAnalyzer) generateNextActions(sequence []string, lastEvent BehaviorEvent) []PredictedAction {
	var actions []PredictedAction

	// Analyze the sequence to predict likely next actions
	lastAction := sequence[len(sequence)-1]

	switch {
	case strings.Contains(lastAction, "get pods"):
		actions = append(actions, PredictedAction{
			Action:     "kubectl describe pod",
			Confidence: 0.7,
			Priority:   2,
			Context:    "diagnostic",
			RiskLevel:  "low",
		})
		actions = append(actions, PredictedAction{
			Action:     "kubectl logs",
			Confidence: 0.6,
			Priority:   2,
			Context:    "diagnostic",
			RiskLevel:  "low",
		})

	case strings.Contains(lastAction, "describe"):
		actions = append(actions, PredictedAction{
			Action:     "kubectl logs",
			Confidence: 0.8,
			Priority:   3,
			Context:    "diagnostic",
			RiskLevel:  "low",
		})
		actions = append(actions, PredictedAction{
			Action:     "kubectl get events",
			Confidence: 0.6,
			Priority:   2,
			Context:    "diagnostic",
			RiskLevel:  "low",
		})

	case strings.Contains(lastAction, "logs"):
		actions = append(actions, PredictedAction{
			Action:     "kubectl get events",
			Confidence: 0.7,
			Priority:   2,
			Context:    "diagnostic",
			RiskLevel:  "low",
		})

	case strings.Contains(lastAction, "apply"):
		actions = append(actions, PredictedAction{
			Action:     "kubectl get pods",
			Confidence: 0.8,
			Priority:   3,
			Context:    "verification",
			RiskLevel:  "low",
		})
	}

	return actions
}

// predictFromContext generates predictions based on current context
func (ba *BehaviorAnalyzer) predictFromContext(context *KubeContextSummary, userInput string) []PredictedAction {
	var predictions []PredictedAction

	// Analyze context for common patterns
	if len(context.PodProblemCounts) > 0 {
		// There are pod problems, suggest diagnostic actions
		predictions = append(predictions, PredictedAction{
			Action:     "kubectl get pods --field-selector=status.phase!=Running",
			Confidence: 0.8,
			Priority:   3,
			Context:    "diagnostic",
			RiskLevel:  "low",
		})
	}

	// Check for namespace-specific patterns
	if strings.Contains(strings.ToLower(context.Namespace), "prod") {
		// Production namespace - suggest safer actions
		predictions = append(predictions, PredictedAction{
			Action:     "kubectl get pods --dry-run=client",
			Confidence: 0.7,
			Priority:   2,
			Context:    "safe_diagnostic",
			RiskLevel:  "low",
		})
	}

	// Analyze user input for intent
	inputLower := strings.ToLower(userInput)
	switch {
	case strings.Contains(inputLower, "debug") || strings.Contains(inputLower, "troubleshoot"):
		predictions = append(predictions, PredictedAction{
			Action:     "kubectl describe pods",
			Confidence: 0.8,
			Priority:   3,
			Context:    "diagnostic",
			RiskLevel:  "low",
		})

	case strings.Contains(inputLower, "scale") || strings.Contains(inputLower, "replicas"):
		predictions = append(predictions, PredictedAction{
			Action:     "kubectl get deployments",
			Confidence: 0.7,
			Priority:   2,
			Context:    "scaling",
			RiskLevel:  "medium",
		})

	case strings.Contains(inputLower, "logs") || strings.Contains(inputLower, "error"):
		predictions = append(predictions, PredictedAction{
			Action:     "kubectl logs --tail=100",
			Confidence: 0.8,
			Priority:   3,
			Context:    "diagnostic",
			RiskLevel:  "low",
		})
	}

	return predictions
}

// getRecentEvents returns the most recent behavior events
func (ba *BehaviorAnalyzer) getRecentEvents(count int) []BehaviorEvent {
	if len(ba.behaviorHistory) < count {
		return ba.behaviorHistory
	}
	return ba.behaviorHistory[len(ba.behaviorHistory)-count:]
}

// matchesEventSequence checks if recent events match a pattern sequence
func (ba *BehaviorAnalyzer) matchesEventSequence(events []BehaviorEvent, sequence []string) bool {
	if len(events) < len(sequence) {
		return false
	}

	for i, action := range sequence {
		if events[len(events)-len(sequence)+i].Action != action {
			return false
		}
	}
	return true
}

// CleanupOldEvents removes old behavior events to manage memory
func (ba *BehaviorAnalyzer) CleanupOldEvents() {
	cutoff := time.Now().Add(-7 * 24 * time.Hour) // Keep 7 days of history

	var filteredHistory []BehaviorEvent
	for _, event := range ba.behaviorHistory {
		if event.Timestamp.After(cutoff) {
			filteredHistory = append(filteredHistory, event)
		}
	}

	ba.behaviorHistory = filteredHistory

	// Clean up old patterns
	for id, pattern := range ba.patterns {
		if pattern.LastSeen.Before(cutoff) && pattern.Confidence < 0.3 {
			delete(ba.patterns, id)
		}
	}
}

// GetBehaviorStats returns behavior analysis statistics
func (ba *BehaviorAnalyzer) GetBehaviorStats() map[string]interface{} {
	ba.mu.RLock()
	defer ba.mu.RUnlock()

	return map[string]interface{}{
		"total_events":     len(ba.behaviorHistory),
		"total_patterns":   len(ba.patterns),
		"anomaly_count":    len(ba.anomalyDetector.anomalyHistory),
		"session_count":    len(ba.sessionTracker.sessionHistory),
		"detector_sensitivity": ba.anomalyDetector.sensitivity,
	}
}

// DetectAnomaly detects unusual behavior patterns
func (ad *AnomalyDetector) DetectAnomaly(userInput string, context *KubeContextSummary) bool {
	// Simple anomaly detection based on input length and context
	inputLength := float64(len(userInput))
	
	if baseline, exists := ad.baselineMetrics["input_length"]; exists {
		if math.Abs(inputLength-baseline) > ad.thresholds["input_length"] {
			ad.anomalyHistory = append(ad.anomalyHistory, AnomalyEvent{
				Timestamp:   time.Now(),
				AnomalyType: "unusual_input_length",
				Severity:    "low",
				Description: fmt.Sprintf("Input length %f deviates from baseline %f", inputLength, baseline),
				Metrics:     map[string]float64{"input_length": inputLength, "baseline": baseline},
				Resolved:    false,
			})
			return true
		}
	} else {
		ad.baselineMetrics["input_length"] = inputLength
		ad.thresholds["input_length"] = inputLength * 0.5
	}

	// Check for unusual command patterns
	if ad.detectUnusualCommands(userInput) {
		return true
	}

	// Check for context anomalies
	if ad.detectContextAnomalies(context) {
		return true
	}

	return false
}

// detectUnusualCommands detects unusual command patterns
func (ad *AnomalyDetector) detectUnusualCommands(userInput string) bool {
	// Check for potentially dangerous commands
	dangerousPatterns := []string{
		"delete all",
		"rm -rf",
		"kubectl delete namespace",
		"helm uninstall",
	}

	inputLower := strings.ToLower(userInput)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(inputLower, pattern) {
			ad.anomalyHistory = append(ad.anomalyHistory, AnomalyEvent{
				Timestamp:   time.Now(),
				AnomalyType: "dangerous_command",
				Severity:    "high",
				Description: fmt.Sprintf("Potentially dangerous command detected: %s", pattern),
				Metrics:     map[string]float64{"danger_score": 1.0},
				Resolved:    false,
			})
			return true
		}
	}

	return false
}

// detectContextAnomalies detects unusual context patterns
func (ad *AnomalyDetector) detectContextAnomalies(context *KubeContextSummary) bool {
	// Check for unusual namespace access patterns
	if strings.Contains(strings.ToLower(context.Namespace), "prod") {
		// Production access - check if this is unusual
		if baseline, exists := ad.baselineMetrics["prod_access"]; exists {
			if baseline < 0.1 { // User rarely accesses production
				ad.anomalyHistory = append(ad.anomalyHistory, AnomalyEvent{
					Timestamp:   time.Now(),
					AnomalyType: "unusual_prod_access",
					Severity:    "medium",
					Description: "Unusual production namespace access detected",
					Metrics:     map[string]float64{"prod_access": 1.0, "baseline": baseline},
					Resolved:    false,
				})
				return true
			}
		} else {
			ad.baselineMetrics["prod_access"] = 0.1
		}
	}

	return false
}

// UpdateBaselines updates anomaly detection baselines
func (ad *AnomalyDetector) UpdateBaselines(event BehaviorEvent) {
	// Update input length baseline
	inputLength := float64(len(event.Action))
	if baseline, exists := ad.baselineMetrics["input_length"]; exists {
		ad.baselineMetrics["input_length"] = baseline*0.95 + inputLength*0.05
	} else {
		ad.baselineMetrics["input_length"] = inputLength
	}

	// Update production access baseline
	if strings.Contains(strings.ToLower(event.Context), "prod") {
		if baseline, exists := ad.baselineMetrics["prod_access"]; exists {
			ad.baselineMetrics["prod_access"] = baseline*0.95 + 0.05
		} else {
			ad.baselineMetrics["prod_access"] = 0.05
		}
	}

	// Update success rate baseline
	successValue := 0.0
	if event.Success {
		successValue = 1.0
	}
	if baseline, exists := ad.baselineMetrics["success_rate"]; exists {
		ad.baselineMetrics["success_rate"] = baseline*0.9 + successValue*0.1
	} else {
		ad.baselineMetrics["success_rate"] = successValue
	}
}

// UpdateSession updates the current session with new events
func (st *SessionTracker) UpdateSession(event BehaviorEvent) {
	if st.currentSession == nil {
		st.startNewSession(event)
		return
	}

	// Add event to current session
	st.currentSession.Actions = append(st.currentSession.Actions, event)

	// Check if session should end (e.g., after period of inactivity)
	if time.Since(st.currentSession.StartTime) > 30*time.Minute {
		st.endCurrentSession()
		st.startNewSession(event)
	}
}

// startNewSession starts a new user session
func (st *SessionTracker) startNewSession(event BehaviorEvent) {
	sessionID := fmt.Sprintf("session-%d", time.Now().Unix())
	
	st.currentSession = &UserSession{
		SessionID:    sessionID,
		StartTime:    time.Now(),
		Actions:      []BehaviorEvent{event},
		Context:      []ContextSnapshot{},
		Outcome:      "in_progress",
		Satisfaction: 0.5,
		Efficiency:   0.5,
	}
}

// endCurrentSession ends the current session and analyzes it
func (st *SessionTracker) endCurrentSession() {
	if st.currentSession == nil {
		return
	}

	st.currentSession.EndTime = time.Now()
	st.currentSession.Outcome = st.analyzeSessionOutcome()

	// Add to history
	st.sessionHistory = append(st.sessionHistory, *st.currentSession)

	// Maintain history size
	if len(st.sessionHistory) > st.maxSessions {
		st.sessionHistory = st.sessionHistory[len(st.sessionHistory)-st.maxSessions:]
	}

	// Extract session patterns
	st.extractSessionPatterns(*st.currentSession)

	st.currentSession = nil
}

// analyzeSessionOutcome analyzes the outcome of a session
func (st *SessionTracker) analyzeSessionOutcome() string {
	if st.currentSession == nil {
		return "unknown"
	}

	successCount := 0
	totalActions := len(st.currentSession.Actions)

	for _, action := range st.currentSession.Actions {
		if action.Success {
			successCount++
		}
	}

	if totalActions == 0 {
		return "empty"
	}

	successRate := float64(successCount) / float64(totalActions)
	switch {
	case successRate >= 0.8:
		return "success"
	case successRate >= 0.5:
		return "partial"
	default:
		return "failure"
	}
}

// extractSessionPatterns extracts patterns from completed sessions
func (st *SessionTracker) extractSessionPatterns(session UserSession) {
	if len(session.Actions) < 2 {
		return
	}

	// Create session type based on actions
	sessionType := st.classifySessionType(session.Actions)
	
	if pattern, exists := st.sessionPatterns[sessionType]; exists {
		pattern.Frequency++
		pattern.LastSeen = time.Now()
		
		// Update success rate
		sessionSuccess := 0.0
		if session.Outcome == "success" {
			sessionSuccess = 1.0
		} else if session.Outcome == "partial" {
			sessionSuccess = 0.5
		}
		
		pattern.SuccessRate = pattern.SuccessRate*0.9 + sessionSuccess*0.1
		pattern.Duration = time.Duration((float64(pattern.Duration)*0.9 + float64(session.EndTime.Sub(session.StartTime))*0.1))
	} else {
		st.sessionPatterns[sessionType] = &SessionPattern{
			SessionType:  sessionType,
			CommonFlow:   st.extractCommonFlow(session.Actions),
			Duration:     session.EndTime.Sub(session.StartTime),
			SuccessRate:  0.5,
			LastSeen:     time.Now(),
			Frequency:    1,
		}
	}
}

// classifySessionType classifies the type of session based on actions
func (st *SessionTracker) classifySessionType(actions []BehaviorEvent) string {
	diagnosticCount := 0
	deploymentCount := 0
	scalingCount := 0

	for _, action := range actions {
		actionLower := strings.ToLower(action.Action)
		switch {
		case strings.Contains(actionLower, "get") || strings.Contains(actionLower, "describe") || strings.Contains(actionLower, "logs"):
			diagnosticCount++
		case strings.Contains(actionLower, "apply") || strings.Contains(actionLower, "create"):
			deploymentCount++
		case strings.Contains(actionLower, "scale"):
			scalingCount++
		}
	}

	// Classify based on predominant action type
	if diagnosticCount > deploymentCount && diagnosticCount > scalingCount {
		return "diagnostic"
	} else if deploymentCount > scalingCount {
		return "deployment"
	} else if scalingCount > 0 {
		return "scaling"
	}

	return "general"
}

// extractCommonFlow extracts common action flow from session
func (st *SessionTracker) extractCommonFlow(actions []BehaviorEvent) []string {
	flow := make([]string, len(actions))
	for i, action := range actions {
		flow[i] = action.Action
	}
	return flow
}

