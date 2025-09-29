// predictive_intelligence.go - Predictive intelligence engine with pattern caching
package engine

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// PredictiveEngine provides intelligent prediction and proactive caching
type PredictiveEngine struct {
	patterns       map[string]*PredictionPattern
	userBehavior   *UserBehaviorModel
	contextPatterns map[string]*ContextPattern
	predictionCache *PredictionCache
	mutex          sync.RWMutex
	enabled        bool
	minConfidence  float64
}

// PredictionPattern represents a learned user interaction pattern
type PredictionPattern struct {
	InputPattern     string             `json:"input_pattern"`
	ContextHash      string             `json:"context_hash"`
	PredictedActions []PredictedAction  `json:"predicted_actions"`
	Confidence       float64            `json:"confidence"`
	Frequency        int                `json:"frequency"`
	LastUsed         time.Time          `json:"last_used"`
	SuccessRate      float64            `json:"success_rate"`
	Metadata         map[string]string  `json:"metadata"`
}

// PredictedAction represents an action the user is likely to take
type PredictedAction struct {
	Type        string    `json:"type"`        // "command", "explanation", "edit", "diagnostic"
	Command     string    `json:"command"`     // kubectl/helm command
	Description string    `json:"description"` // Human-readable description
	Confidence  float64   `json:"confidence"`  // 0.0 - 1.0
	Priority    int       `json:"priority"`    // 1-10
	Conditions  []string  `json:"conditions"`  // Preconditions for this action
	NextActions []string  `json:"next_actions"` // Likely next actions after this one
}

// UserBehaviorModel tracks user behavior patterns over time
type UserBehaviorModel struct {
	CommandFrequency    map[string]int            `json:"command_frequency"`
	SequencePatterns    map[string]CommandSequence `json:"sequence_patterns"`
	TimePatterns        map[int][]string          `json:"time_patterns"` // hour -> common commands
	ContextPreferences  map[string][]string       `json:"context_preferences"`
	TypingSpeed         float64                   `json:"typing_speed"`
	SessionDuration     time.Duration             `json:"session_duration"`
	ErrorPatterns       map[string]int            `json:"error_patterns"`
	LastUpdated         time.Time                 `json:"last_updated"`
}

// CommandSequence represents a sequence of commands users often execute together
type CommandSequence struct {
	Commands    []string  `json:"commands"`
	Frequency   int       `json:"frequency"`
	AvgDelay    time.Duration `json:"avg_delay"`
	Confidence  float64   `json:"confidence"`
	LastSeen    time.Time `json:"last_seen"`
}

// ContextPattern represents patterns in cluster context that predict user needs
// ContextPattern is now defined in types.go

// PredictionCache manages cached predictions and preloaded results
type PredictionCache struct {
	entries        map[string]*CachedPrediction
	preloadQueue   chan PredictionWork
	maxSize        int
	ttl            time.Duration
	hitRate        float64
	totalRequests  int
	cacheHits      int
	mutex          sync.RWMutex
}

// CachedPrediction represents a cached prediction result
type CachedPrediction struct {
	Key         string           `json:"key"`
	Predictions []PredictedAction `json:"predictions"`
	Confidence  float64          `json:"confidence"`
	Timestamp   time.Time        `json:"timestamp"`
	AccessCount int              `json:"access_count"`
	LastAccess  time.Time        `json:"last_access"`
}

// PredictionWork represents work to preload predictions
type PredictionWork struct {
	InputPattern string
	Context      *KubeContextSummary
	Priority     int
}

// NewPredictiveEngine creates a new predictive intelligence engine
func NewPredictiveEngine() *PredictiveEngine {
	return &PredictiveEngine{
		patterns:        make(map[string]*PredictionPattern),
		userBehavior:    NewUserBehaviorModel(),
		contextPatterns: make(map[string]*ContextPattern),
		predictionCache: NewPredictionCache(1000, 5*time.Minute), // 1000 entries, 5min TTL
		enabled:         true,
		minConfidence:   0.6,
	}
}

// NewUserBehaviorModel creates a new user behavior model
func NewUserBehaviorModel() *UserBehaviorModel {
	return &UserBehaviorModel{
		CommandFrequency:   make(map[string]int),
		SequencePatterns:   make(map[string]CommandSequence),
		TimePatterns:       make(map[int][]string),
		ContextPreferences: make(map[string][]string),
		ErrorPatterns:      make(map[string]int),
		LastUpdated:        time.Now(),
	}
}

// NewPredictionCache creates a new prediction cache
func NewPredictionCache(maxSize int, ttl time.Duration) *PredictionCache {
	cache := &PredictionCache{
		entries:     make(map[string]*CachedPrediction),
		preloadQueue: make(chan PredictionWork, 100),
		maxSize:     maxSize,
		ttl:         ttl,
	}

	// Start preload worker
	go cache.preloadWorker()

	return cache
}

// PredictActions predicts likely next actions based on input and context
func (pe *PredictiveEngine) PredictActions(input string, context *KubeContextSummary) []PredictedAction {
	if !pe.enabled {
		return nil
	}

	pe.mutex.RLock()
	defer pe.mutex.RUnlock()

	// Create cache key
	cacheKey := pe.createCacheKey(input, context)

	// Check cache first
	if cached := pe.predictionCache.Get(cacheKey); cached != nil {
		return cached.Predictions
	}

	// Generate predictions
	predictions := pe.generatePredictions(input, context)

	// Cache the results
	pe.predictionCache.Set(cacheKey, predictions, pe.calculateConfidence(predictions))

	// Trigger preloading of related patterns
	pe.triggerPreload(input, context)

	return predictions
}

// generatePredictions creates predictions based on patterns and behavior
func (pe *PredictiveEngine) generatePredictions(input string, context *KubeContextSummary) []PredictedAction {
	var predictions []PredictedAction

	// Pattern-based predictions
	patternPredictions := pe.getPredictionsFromPatterns(input, context)
	predictions = append(predictions, patternPredictions...)

	// Context-based predictions
	contextPredictions := pe.getPredictionsFromContext(context)
	predictions = append(predictions, contextPredictions...)

	// Behavior-based predictions
	behaviorPredictions := pe.getPredictionsFromBehavior(input)
	predictions = append(predictions, behaviorPredictions...)

	// Command sequence predictions
	sequencePredictions := pe.getPredictionsFromSequences(input)
	predictions = append(predictions, sequencePredictions...)

	// Deduplicate and sort by confidence
	predictions = pe.deduplicateAndSort(predictions)

	// Filter by minimum confidence
	var filteredPredictions []PredictedAction
	for _, pred := range predictions {
		if pred.Confidence >= pe.minConfidence {
			filteredPredictions = append(filteredPredictions, pred)
		}
	}

	return filteredPredictions
}

// getPredictionsFromPatterns gets predictions based on learned patterns
func (pe *PredictiveEngine) getPredictionsFromPatterns(input string, context *KubeContextSummary) []PredictedAction {
	var predictions []PredictedAction

	inputPattern := pe.extractInputPattern(input)
	contextHash := pe.createContextHash(context)

	for _, pattern := range pe.patterns {
		if pe.matchesPattern(inputPattern, pattern.InputPattern) &&
		   pe.matchesContext(contextHash, pattern.ContextHash) {

			// Calculate confidence based on pattern success rate and recency
			confidence := pattern.Confidence * pattern.SuccessRate
			if time.Since(pattern.LastUsed) > 24*time.Hour {
				confidence *= 0.8 // Reduce confidence for old patterns
			}

			for _, action := range pattern.PredictedActions {
				action.Confidence = confidence
				action.Priority = pe.calculatePriority(action, pattern)
				predictions = append(predictions, action)
			}
		}
	}

	return predictions
}

// getPredictionsFromContext gets predictions based on current cluster context
func (pe *PredictiveEngine) getPredictionsFromContext(context *KubeContextSummary) []PredictedAction {
	var predictions []PredictedAction

	if context == nil {
		return predictions
	}

	// Predict actions based on pod problems
	for problem, count := range context.PodProblemCounts {
		if count > 0 {
			actions := pe.getActionsForProblem(problem, count)
			predictions = append(predictions, actions...)
		}
	}

	// Predict actions based on pod phases
	if context.PodPhaseCounts != nil {
		pending := context.PodPhaseCounts["Pending"]
		failed := context.PodPhaseCounts["Failed"]
		total := 0
		for _, count := range context.PodPhaseCounts {
			total += count
		}

		if total > 0 {
			problemRatio := float64(pending+failed) / float64(total)
			if problemRatio > 0.2 {
				// High problem ratio suggests diagnostic needs
				predictions = append(predictions, PredictedAction{
					Type:        "diagnostic",
					Command:     "kubectl get events --sort-by=.lastTimestamp",
					Description: "Check recent cluster events for issues",
					Confidence:  0.8,
					Priority:    8,
				})
			}
		}
	}

	return predictions
}

// getPredictionsFromBehavior gets predictions based on user behavior patterns
func (pe *PredictiveEngine) getPredictionsFromBehavior(input string) []PredictedAction {
	var predictions []PredictedAction

	// Get frequent commands
	var frequentCommands []struct {
		command string
		freq    int
	}

	for cmd, freq := range pe.userBehavior.CommandFrequency {
		frequentCommands = append(frequentCommands, struct {
			command string
			freq    int
		}{cmd, freq})
	}

	// Sort by frequency
	sort.Slice(frequentCommands, func(i, j int) bool {
		return frequentCommands[i].freq > frequentCommands[j].freq
	})

	// Create predictions from top frequent commands
	for i, cmd := range frequentCommands {
		if i >= 5 { // Limit to top 5
			break
		}

		confidence := 0.5 + (float64(cmd.freq)/float64(pe.userBehavior.CommandFrequency["total"])*0.4)
		predictions = append(predictions, PredictedAction{
			Type:        "command",
			Command:     cmd.command,
			Description: fmt.Sprintf("Frequently used command (%d times)", cmd.freq),
			Confidence:  confidence,
			Priority:    5 - i, // Higher priority for more frequent
		})
	}

	return predictions
}

// getPredictionsFromSequences gets predictions based on command sequences
func (pe *PredictiveEngine) getPredictionsFromSequences(input string) []PredictedAction {
	var predictions []PredictedAction

	// Look for matching sequence patterns
	for _, sequence := range pe.userBehavior.SequencePatterns {
		if len(sequence.Commands) > 1 {
			// Check if current input matches the beginning of a sequence
			firstCmd := sequence.Commands[0]
			if pe.commandSimilarity(input, firstCmd) > 0.7 {
				// Predict the next command in the sequence
				if len(sequence.Commands) > 1 {
					nextCmd := sequence.Commands[1]
					predictions = append(predictions, PredictedAction{
						Type:        "command",
						Command:     nextCmd,
						Description: fmt.Sprintf("Next in sequence after '%s'", firstCmd),
						Confidence:  sequence.Confidence,
						Priority:    7,
						NextActions: sequence.Commands[2:], // Remaining commands
					})
				}
			}
		}
	}

	return predictions
}

// LearnFromInteraction learns from user interactions to improve predictions
func (pe *PredictiveEngine) LearnFromInteraction(input string, context *KubeContextSummary, action string, success bool) {
	pe.mutex.Lock()
	defer pe.mutex.Unlock()

	// Update user behavior
	pe.userBehavior.CommandFrequency[action]++
	pe.userBehavior.CommandFrequency["total"]++
	pe.userBehavior.LastUpdated = time.Now()

	// Update or create pattern
	patternKey := pe.createPatternKey(input, context)
	pattern, exists := pe.patterns[patternKey]

	if !exists {
		pattern = &PredictionPattern{
			InputPattern:     pe.extractInputPattern(input),
			ContextHash:      pe.createContextHash(context),
			PredictedActions: []PredictedAction{},
			Frequency:        1,
			LastUsed:         time.Now(),
			SuccessRate:      0.0,
			Metadata:         make(map[string]string),
		}
		pe.patterns[patternKey] = pattern
	}

	// Update pattern
	pattern.Frequency++
	pattern.LastUsed = time.Now()

	// Update success rate
	if success {
		pattern.SuccessRate = (pattern.SuccessRate*float64(pattern.Frequency-1) + 1.0) / float64(pattern.Frequency)
	} else {
		pattern.SuccessRate = (pattern.SuccessRate*float64(pattern.Frequency-1) + 0.0) / float64(pattern.Frequency)
	}

	// Add or update predicted action
	found := false
	for i, predAction := range pattern.PredictedActions {
		if predAction.Command == action {
			pattern.PredictedActions[i].Confidence = pe.calculateUpdatedConfidence(
				predAction.Confidence, success, pattern.Frequency)
			found = true
			break
		}
	}

	if !found {
		newAction := PredictedAction{
			Type:        pe.classifyActionType(action),
			Command:     action,
			Description: pe.generateActionDescription(action),
			Confidence:  0.7,
			Priority:    5,
		}
		pattern.PredictedActions = append(pattern.PredictedActions, newAction)
	}

	// Update context patterns
	pe.updateContextPatterns(context, action, success)
}

// Helper methods

func (pe *PredictiveEngine) createCacheKey(input string, context *KubeContextSummary) string {
	contextStr := ""
	if context != nil {
		contextStr = fmt.Sprintf("%s:%s:%d", context.Context, context.Namespace, len(context.PodProblemCounts))
	}
	hash := md5.Sum([]byte(input + ":" + contextStr))
	return fmt.Sprintf("%x", hash)
}

func (pe *PredictiveEngine) createPatternKey(input string, context *KubeContextSummary) string {
	inputPattern := pe.extractInputPattern(input)
	contextHash := pe.createContextHash(context)
	return fmt.Sprintf("%s:%s", inputPattern, contextHash)
}

func (pe *PredictiveEngine) extractInputPattern(input string) string {
	// Extract key terms and patterns from user input
	input = strings.ToLower(strings.TrimSpace(input))

	// Common patterns
	patterns := []struct {
		regex   string
		pattern string
	}{
		{`pods?.*not.*running`, "pods_not_running"},
		{`service.*not.*working`, "service_not_working"},
		{`deploy.*fail`, "deployment_failing"},
		{`get.*pods?`, "get_pods"},
		{`describe.*pod`, "describe_pod"},
		{`logs?.*pod`, "pod_logs"},
		{`scale.*deployment`, "scale_deployment"},
	}

	for _, p := range patterns {
		if matched, _ := regexp.MatchString(p.regex, input); matched {
			return p.pattern
		}
	}

	// Fallback to first few words
	words := strings.Fields(input)
	if len(words) > 3 {
		words = words[:3]
	}
	return strings.Join(words, "_")
}

func (pe *PredictiveEngine) createContextHash(context *KubeContextSummary) string {
	if context == nil {
		return "no_context"
	}

	signature := fmt.Sprintf("%s:%s:%d:%d",
		context.Context,
		context.Namespace,
		len(context.PodPhaseCounts),
		len(context.PodProblemCounts))

	hash := md5.Sum([]byte(signature))
	return fmt.Sprintf("%x", hash)[:8] // Short hash
}

func (pe *PredictiveEngine) calculateConfidence(predictions []PredictedAction) float64 {
	if len(predictions) == 0 {
		return 0.0
	}

	total := 0.0
	for _, pred := range predictions {
		total += pred.Confidence
	}
	return total / float64(len(predictions))
}

func (pe *PredictiveEngine) deduplicateAndSort(predictions []PredictedAction) []PredictedAction {
	seen := make(map[string]bool)
	var unique []PredictedAction

	for _, pred := range predictions {
		key := pred.Command + ":" + pred.Type
		if !seen[key] {
			seen[key] = true
			unique = append(unique, pred)
		}
	}

	// Sort by confidence and priority
	sort.Slice(unique, func(i, j int) bool {
		if unique[i].Confidence != unique[j].Confidence {
			return unique[i].Confidence > unique[j].Confidence
		}
		return unique[i].Priority > unique[j].Priority
	})

	return unique
}

// Cache methods

func (pc *PredictionCache) Get(key string) *CachedPrediction {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()

	pc.totalRequests++

	if entry, exists := pc.entries[key]; exists {
		if time.Since(entry.Timestamp) < pc.ttl {
			entry.AccessCount++
			entry.LastAccess = time.Now()
			pc.cacheHits++
			pc.hitRate = float64(pc.cacheHits) / float64(pc.totalRequests)
			return entry
		} else {
			// Entry expired
			delete(pc.entries, key)
		}
	}

	return nil
}

func (pc *PredictionCache) Set(key string, predictions []PredictedAction, confidence float64) {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()

	// Check cache size and evict if necessary
	if len(pc.entries) >= pc.maxSize {
		pc.evictLeastUsed()
	}

	pc.entries[key] = &CachedPrediction{
		Key:         key,
		Predictions: predictions,
		Confidence:  confidence,
		Timestamp:   time.Now(),
		AccessCount: 1,
		LastAccess:  time.Now(),
	}
}

func (pc *PredictionCache) evictLeastUsed() {
	oldestKey := ""
	oldestTime := time.Now()

	for key, entry := range pc.entries {
		if entry.LastAccess.Before(oldestTime) {
			oldestTime = entry.LastAccess
			oldestKey = key
		}
	}

	if oldestKey != "" {
		delete(pc.entries, oldestKey)
	}
}

func (pc *PredictionCache) preloadWorker() {
	for work := range pc.preloadQueue {
		// This would preload predictions in the background
		// For now, we'll just consume the work
		_ = work
	}
}

// Implement remaining helper methods with simple logic for now
func (pe *PredictiveEngine) matchesPattern(input, pattern string) bool {
	return strings.Contains(input, pattern) || strings.Contains(pattern, input)
}

func (pe *PredictiveEngine) matchesContext(hash1, hash2 string) bool {
	return hash1 == hash2
}

func (pe *PredictiveEngine) calculatePriority(action PredictedAction, pattern *PredictionPattern) int {
	priority := action.Priority
	if pattern.Frequency > 10 {
		priority += 2
	}
	return min(priority, 10)
}

func (pe *PredictiveEngine) getActionsForProblem(problem string, count int) []PredictedAction {
	actions := []PredictedAction{}

	switch problem {
	case "ImagePullBackOff":
		actions = append(actions, PredictedAction{
			Type:        "diagnostic",
			Command:     "kubectl describe pods",
			Description: "Check image pull issues",
			Confidence:  0.9,
			Priority:    9,
		})
	case "CrashLoopBackOff":
		actions = append(actions, PredictedAction{
			Type:        "diagnostic",
			Command:     "kubectl logs",
			Description: "Check container logs for crash details",
			Confidence:  0.9,
			Priority:    9,
		})
	}

	return actions
}

func (pe *PredictiveEngine) commandSimilarity(cmd1, cmd2 string) float64 {
	// Simple similarity check
	if cmd1 == cmd2 {
		return 1.0
	}
	if strings.Contains(cmd1, cmd2) || strings.Contains(cmd2, cmd1) {
		return 0.7
	}
	return 0.0
}

func (pe *PredictiveEngine) calculateUpdatedConfidence(oldConf float64, success bool, frequency int) float64 {
	newConf := oldConf
	if success {
		newConf += 0.1
	} else {
		newConf -= 0.1
	}
	return max(0.0, min(1.0, newConf))
}

func (pe *PredictiveEngine) classifyActionType(action string) string {
	action = strings.ToLower(action)
	if strings.Contains(action, "get") || strings.Contains(action, "describe") {
		return "diagnostic"
	} else if strings.Contains(action, "apply") || strings.Contains(action, "create") {
		return "command"
	}
	return "command"
}

func (pe *PredictiveEngine) generateActionDescription(action string) string {
	return fmt.Sprintf("Execute: %s", action)
}

func (pe *PredictiveEngine) updateContextPatterns(context *KubeContextSummary, action string, success bool) {
	// Simple implementation for context pattern updates
	if context != nil {
		contextKey := pe.createContextHash(context)
		if pattern, exists := pe.contextPatterns[contextKey]; exists {
			pattern.LastSeen = time.Now()
		}
	}
}

func (pe *PredictiveEngine) triggerPreload(input string, context *KubeContextSummary) {
	// Trigger preloading of related patterns
	work := PredictionWork{
		InputPattern: pe.extractInputPattern(input),
		Context:      context,
		Priority:     5,
	}

	select {
	case pe.predictionCache.preloadQueue <- work:
	default:
		// Queue full, skip preload
	}
}

// min and max functions are now defined in types.go

// Global predictive engine instance
var GlobalPredictiveEngine = NewPredictiveEngine()