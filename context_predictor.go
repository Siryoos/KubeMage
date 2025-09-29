// context_predictor.go - Context prediction and transition analysis
package main

import (
	"fmt"
	"time"
)

// PredictContextChanges predicts likely context changes based on patterns
func (cp *ContextPredictor) PredictContextChanges(context *KubeContextSummary) []PredictedAction {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	var predictions []PredictedAction

	// Analyze context transition patterns
	currentState := cp.contextToState(context)
	if transitions, exists := cp.transitionMatrix[currentState]; exists {
		for nextState, probability := range transitions {
			if probability > 0.3 {
				predictions = append(predictions, PredictedAction{
					Action:     fmt.Sprintf("switch to %s", nextState),
					Confidence: probability,
					Priority:   1,
					Context:    nextState,
					RiskLevel:  "low",
				})
			}
		}
	}

	// Add time-based predictions
	timePredictions := cp.predictFromTimePatterns(context)
	predictions = append(predictions, timePredictions...)

	// Add health-based predictions
	healthPredictions := cp.predictFromHealthPatterns(context)
	predictions = append(predictions, healthPredictions...)

	return predictions
}

// UpdateTransitions updates context transition patterns
func (cp *ContextPredictor) UpdateTransitions(context *KubeContextSummary) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	currentState := cp.contextToState(context)
	
	// Add to history
	snapshot := ContextSnapshot{
		Timestamp:    time.Now(),
		Context:      context,
		UserActivity: cp.determineUserActivity(context),
		SystemLoad:   cp.calculateSystemLoad(context),
		PredictedNext: cp.generatePredictions(context),
	}
	
	cp.contextHistory = append(cp.contextHistory, snapshot)
	
	// Maintain history size
	if len(cp.contextHistory) > 100 {
		cp.contextHistory = cp.contextHistory[len(cp.contextHistory)-100:]
	}

	// Update transition matrix
	if len(cp.contextHistory) >= 2 {
		prevState := cp.contextToState(cp.contextHistory[len(cp.contextHistory)-2].Context)
		
		if cp.transitionMatrix[prevState] == nil {
			cp.transitionMatrix[prevState] = make(map[string]float64)
		}
		
		cp.transitionMatrix[prevState][currentState] += 0.1
		
		// Normalize probabilities
		total := 0.0
		for _, prob := range cp.transitionMatrix[prevState] {
			total += prob
		}
		
		if total > 0 {
			for state := range cp.transitionMatrix[prevState] {
				cp.transitionMatrix[prevState][state] /= total
			}
		}
	}

	// Update accuracy based on prediction success
	cp.updateAccuracy(context)
}

// predictFromTimePatterns generates predictions based on time patterns
func (cp *ContextPredictor) predictFromTimePatterns(context *KubeContextSummary) []PredictedAction {
	var predictions []PredictedAction

	currentHour := time.Now().Hour()
	
	// Business hours patterns
	if currentHour >= 9 && currentHour <= 17 {
		// During business hours, predict more deployment activities
		predictions = append(predictions, PredictedAction{
			Action:     "kubectl get deployments",
			Confidence: 0.6,
			Priority:   2,
			Context:    "business_hours",
			RiskLevel:  "low",
		})
	} else {
		// Outside business hours, predict more monitoring activities
		predictions = append(predictions, PredictedAction{
			Action:     "kubectl get pods --all-namespaces",
			Confidence: 0.7,
			Priority:   2,
			Context:    "monitoring",
			RiskLevel:  "low",
		})
	}

	// Weekend patterns
	if time.Now().Weekday() == time.Saturday || time.Now().Weekday() == time.Sunday {
		predictions = append(predictions, PredictedAction{
			Action:     "kubectl get events --sort-by=.metadata.creationTimestamp",
			Confidence: 0.5,
			Priority:   1,
			Context:    "weekend_monitoring",
			RiskLevel:  "low",
		})
	}

	return predictions
}

// predictFromHealthPatterns generates predictions based on cluster health
func (cp *ContextPredictor) predictFromHealthPatterns(context *KubeContextSummary) []PredictedAction {
	var predictions []PredictedAction

	// Analyze pod problems
	if len(context.PodProblemCounts) > 0 {
		predictions = append(predictions, PredictedAction{
			Action:     "kubectl describe pods --field-selector=status.phase!=Running",
			Confidence: 0.8,
			Priority:   3,
			Context:    "health_diagnostic",
			RiskLevel:  "low",
		})

		// If there are many problems, suggest broader investigation
		totalProblems := 0
		for _, count := range context.PodProblemCounts {
			totalProblems += count
		}

		if totalProblems > 5 {
			predictions = append(predictions, PredictedAction{
				Action:     "kubectl get events --sort-by=.metadata.creationTimestamp --field-selector type=Warning",
				Confidence: 0.9,
				Priority:   3,
				Context:    "cluster_investigation",
				RiskLevel:  "low",
			})
		}
	}

	// Predict based on namespace health
	if context.Namespace != "default" && context.Namespace != "kube-system" {
		predictions = append(predictions, PredictedAction{
			Action:     fmt.Sprintf("kubectl get all -n %s", context.Namespace),
			Confidence: 0.6,
			Priority:   2,
			Context:    "namespace_overview",
			RiskLevel:  "low",
		})
	}

	return predictions
}

// determineUserActivity determines current user activity level
func (cp *ContextPredictor) determineUserActivity(context *KubeContextSummary) string {
	// Simple activity determination based on context
	if len(context.PodProblemCounts) > 0 {
		return "troubleshooting"
	}

	// Check recent history for activity patterns
	if len(cp.contextHistory) > 0 {
		lastSnapshot := cp.contextHistory[len(cp.contextHistory)-1]
		timeSinceLastActivity := time.Since(lastSnapshot.Timestamp)
		
		if timeSinceLastActivity < 1*time.Minute {
			return "active"
		} else if timeSinceLastActivity < 5*time.Minute {
			return "moderate"
		}
	}

	return "idle"
}

// calculateSystemLoad calculates current system load indicator
func (cp *ContextPredictor) calculateSystemLoad(context *KubeContextSummary) float64 {
	// Simple system load calculation based on context
	load := 0.0

	// Base load from pod problems
	for _, count := range context.PodProblemCounts {
		load += float64(count) * 0.1
	}

	// Adjust based on namespace
	if context.Namespace == "kube-system" {
		load += 0.2 // System namespace adds to load
	}

	// Cap at 1.0
	if load > 1.0 {
		load = 1.0
	}

	return load
}

// generatePredictions generates next predictions for context snapshot
func (cp *ContextPredictor) generatePredictions(context *KubeContextSummary) []string {
	var predictions []string

	// Generate simple predictions based on current context
	if len(context.PodProblemCounts) > 0 {
		predictions = append(predictions, "diagnostic_commands")
	}

	if context.Namespace != "default" {
		predictions = append(predictions, "namespace_exploration")
	}

	predictions = append(predictions, "status_check")

	return predictions
}

// updateAccuracy updates prediction accuracy based on actual outcomes
func (cp *ContextPredictor) updateAccuracy(context *KubeContextSummary) {
	if len(cp.contextHistory) < 2 {
		return
	}

	// Get previous prediction
	prevSnapshot := cp.contextHistory[len(cp.contextHistory)-2]
	currentState := cp.contextToState(context)

	// Check if any predictions were correct
	correct := false
	for _, prediction := range prevSnapshot.PredictedNext {
		if prediction == currentState || cp.isRelatedPrediction(prediction, currentState) {
			correct = true
			break
		}
	}

	// Update accuracy using exponential moving average
	if correct {
		cp.accuracy = cp.accuracy*0.9 + 0.1
	} else {
		cp.accuracy = cp.accuracy * 0.95
	}

	// Ensure accuracy stays within bounds
	if cp.accuracy > 1.0 {
		cp.accuracy = 1.0
	} else if cp.accuracy < 0.0 {
		cp.accuracy = 0.0
	}
}

// isRelatedPrediction checks if a prediction is related to the actual outcome
func (cp *ContextPredictor) isRelatedPrediction(prediction, actual string) bool {
	// Simple relationship checking
	predictionParts := strings.Split(prediction, ":")
	actualParts := strings.Split(actual, ":")

	if len(predictionParts) >= 1 && len(actualParts) >= 1 {
		return predictionParts[0] == actualParts[0] // Same cluster
	}

	return false
}

// contextToState converts context to a state string for transition matrix
func (cp *ContextPredictor) contextToState(context *KubeContextSummary) string {
	return fmt.Sprintf("%s:%s", context.Context, context.Namespace)
}

// CleanupOldHistory removes old context history to manage memory
func (cp *ContextPredictor) CleanupOldHistory() {
	cutoff := time.Now().Add(-24 * time.Hour) // Keep 24 hours of history

	var filteredHistory []ContextSnapshot
	for _, snapshot := range cp.contextHistory {
		if snapshot.Timestamp.After(cutoff) {
			filteredHistory = append(filteredHistory, snapshot)
		}
	}

	cp.contextHistory = filteredHistory

	// Clean up old transition matrix entries with low probabilities
	for fromState, transitions := range cp.transitionMatrix {
		for toState, probability := range transitions {
			if probability < 0.05 { // Remove very low probability transitions
				delete(transitions, toState)
			}
		}
		
		// Remove empty transition maps
		if len(transitions) == 0 {
			delete(cp.transitionMatrix, fromState)
		}
	}
}

// GetPredictionStats returns context prediction statistics
func (cp *ContextPredictor) GetPredictionStats() map[string]interface{} {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	transitionCount := 0
	for _, transitions := range cp.transitionMatrix {
		transitionCount += len(transitions)
	}

	return map[string]interface{}{
		"accuracy":         cp.accuracy,
		"history_length":   len(cp.contextHistory),
		"transition_states": len(cp.transitionMatrix),
		"transition_count": transitionCount,
		"prediction_window": cp.predictionWindow.String(),
	}
}

// PredictNextContext predicts the most likely next context
func (cp *ContextPredictor) PredictNextContext(currentContext *KubeContextSummary) *ContextSnapshot {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	currentState := cp.contextToState(currentContext)
	
	// Find most likely next state
	var mostLikelyState string
	var highestProbability float64

	if transitions, exists := cp.transitionMatrix[currentState]; exists {
		for state, probability := range transitions {
			if probability > highestProbability {
				highestProbability = probability
				mostLikelyState = state
			}
		}
	}

	if mostLikelyState == "" {
		return nil
	}

	// Create predicted context snapshot
	return &ContextSnapshot{
		Timestamp:    time.Now().Add(cp.predictionWindow),
		Context:      currentContext, // In a real implementation, this would be modified based on prediction
		UserActivity: cp.determineUserActivity(currentContext),
		SystemLoad:   cp.calculateSystemLoad(currentContext),
		PredictedNext: []string{mostLikelyState},
	}
}

