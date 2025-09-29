// streaming_intelligence.go - Real-time intelligence streaming with throttling
package engine

import (
	"context"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// StreamingIntelligenceManager handles real-time intelligence updates
type StreamingIntelligenceManager struct {
	updateStream    chan IntelligenceUpdate
	subscribers     map[string]*StreamSubscriber
	throttleManager *StreamThrottleManager
	program         *tea.Program
	ctx             context.Context
	cancel          context.CancelFunc
	running         bool
	mutex           sync.RWMutex
	updateBuffer    *UpdateBuffer
	analytics       *StreamAnalytics
}

// IntelligenceUpdate represents a streaming intelligence update
type IntelligenceUpdate struct {
	ID          string                 `json:"id"`
	Type        IntelligenceUpdateType `json:"type"`
	Priority    int                    `json:"priority"`
	Timestamp   time.Time              `json:"timestamp"`
	Source      string                 `json:"source"`
	Data        interface{}            `json:"data"`
	Context     *KubeContextSummary    `json:"context"`
	Confidence  float64                `json:"confidence"`
	Invalidates []string               `json:"invalidates"` // Cache keys to invalidate
	TTL         time.Duration          `json:"ttl"`
}

// IntelligenceUpdateType defines types of intelligence updates
type IntelligenceUpdateType string

const (
	UpdateTypeRiskChange      IntelligenceUpdateType = "risk_change"
	UpdateTypeNewSuggestion   IntelligenceUpdateType = "new_suggestion"
	UpdateTypeContextChange   IntelligenceUpdateType = "context_change"
	UpdateTypeHealthChange    IntelligenceUpdateType = "health_change"
	UpdateTypePrediction      IntelligenceUpdateType = "prediction"
	UpdateTypeOptimization    IntelligenceUpdateType = "optimization"
	UpdateTypeDiagnostic      IntelligenceUpdateType = "diagnostic"
	UpdateTypeUserBehavior    IntelligenceUpdateType = "user_behavior"
	UpdateTypeCacheEviction   IntelligenceUpdateType = "cache_eviction"
)

// StreamSubscriber manages subscription to intelligence updates
type StreamSubscriber struct {
	ID          string
	Channel     chan IntelligenceUpdate
	Filters     []UpdateFilter
	LastUpdate  time.Time
	Active      bool
	BufferSize  int
	Dropped     int64
	Received    int64
}

// UpdateFilter defines filters for intelligence updates
type UpdateFilter struct {
	Type      IntelligenceUpdateType
	MinPriority int
	ContextFilter func(*KubeContextSummary) bool
	CustomFilter  func(IntelligenceUpdate) bool
}

// StreamThrottleManager manages intelligent throttling of updates
type StreamThrottleManager struct {
	rules         map[IntelligenceUpdateType]*ThrottleRule
	userActivity  *UserActivityTracker
	bandwidth     *BandwidthManager
	adaptiveRates map[string]float64
	mutex         sync.RWMutex
}

// ThrottleRule defines throttling behavior for update types
type ThrottleRule struct {
	MaxRate       float64       // Updates per second
	BurstSize     int           // Max burst updates
	CooldownTime  time.Duration // Cooldown after burst
	AdaptiveRate  bool          // Adjust rate based on user activity
	PriorityBoost float64       // Rate multiplier for high priority updates
}

// UserActivityTracker tracks user interaction patterns
type UserActivityTracker struct {
	lastActivity    time.Time
	activityLevel   ActivityLevel
	typingSpeed     float64 // Characters per minute
	focusTime       time.Duration
	sessionStart    time.Time
	interactionRate float64
	mutex           sync.RWMutex
}

// ActivityLevel represents user activity intensity
type ActivityLevel string

const (
	ActivityIdle     ActivityLevel = "idle"
	ActivityLow      ActivityLevel = "low"
	ActivityModerate ActivityLevel = "moderate"
	ActivityHigh     ActivityLevel = "high"
	ActivityIntense  ActivityLevel = "intense"
)

// BandwidthManager manages update bandwidth based on system performance
type BandwidthManager struct {
	availableBandwidth float64 // Updates per second
	currentUsage       float64
	qualitySettings    QualitySettings
	performanceMetrics *PerformanceMetrics
	mutex              sync.RWMutex
}

// QualitySettings defines update quality vs performance tradeoffs
type QualitySettings struct {
	MaxUpdatesPerSecond int
	PriorityThreshold   int
	CompressionLevel    int
	BatchingEnabled     bool
	PredictiveFiltering bool
}

// PerformanceMetrics tracks streaming performance
// PerformanceMetrics is now defined in types.go

// UpdateBuffer manages buffering and batching of updates
type UpdateBuffer struct {
	updates      []IntelligenceUpdate
	maxSize      int
	flushTicker  *time.Ticker
	batchSize    int
	compression  bool
	deduplication bool
	mutex        sync.RWMutex
}

// StreamAnalytics tracks streaming intelligence performance
type StreamAnalytics struct {
	totalUpdates      int64
	updatesByType     map[IntelligenceUpdateType]int64
	averageLatency    time.Duration
	peakRate          float64
	throttleEvents    int64
	userSatisfaction  float64
	predictionAccuracy float64
	cacheHitRate      float64
	mutex             sync.RWMutex
}

// NewStreamingIntelligenceManager creates a new streaming manager
func NewStreamingIntelligenceManager(program *tea.Program) *StreamingIntelligenceManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &StreamingIntelligenceManager{
		updateStream:    make(chan IntelligenceUpdate, 1000),
		subscribers:     make(map[string]*StreamSubscriber),
		throttleManager: NewStreamThrottleManager(),
		program:         program,
		ctx:             ctx,
		cancel:          cancel,
		running:         false,
		updateBuffer:    NewUpdateBuffer(100, 500*time.Millisecond),
		analytics:       NewStreamAnalytics(),
	}
}

// Start begins streaming intelligence updates
func (sim *StreamingIntelligenceManager) Start() error {
	sim.mutex.Lock()
	defer sim.mutex.Unlock()

	if sim.running {
		return nil
	}

	sim.running = true

	// Start main streaming processor
	go sim.processUpdates()

	// Start buffer flusher
	go sim.updateBuffer.startFlushing(sim.ctx, sim.flushUpdates)

	// Start analytics collector
	go sim.analytics.collectMetrics(sim.ctx)

	// Start throttle manager
	go sim.throttleManager.adaptiveAdjustment(sim.ctx)

	return nil
}

// Stop gracefully stops the streaming manager
func (sim *StreamingIntelligenceManager) Stop() error {
	sim.mutex.Lock()
	defer sim.mutex.Unlock()

	if !sim.running {
		return nil
	}

	sim.running = false
	sim.cancel()
	close(sim.updateStream)
	sim.updateBuffer.stop()

	return nil
}

// Subscribe creates a new subscriber for intelligence updates
func (sim *StreamingIntelligenceManager) Subscribe(id string, filters []UpdateFilter, bufferSize int) *StreamSubscriber {
	sim.mutex.Lock()
	defer sim.mutex.Unlock()

	subscriber := &StreamSubscriber{
		ID:         id,
		Channel:    make(chan IntelligenceUpdate, bufferSize),
		Filters:    filters,
		LastUpdate: time.Now(),
		Active:     true,
		BufferSize: bufferSize,
	}

	sim.subscribers[id] = subscriber
	return subscriber
}

// Unsubscribe removes a subscriber
func (sim *StreamingIntelligenceManager) Unsubscribe(id string) {
	sim.mutex.Lock()
	defer sim.mutex.Unlock()

	if subscriber, exists := sim.subscribers[id]; exists {
		subscriber.Active = false
		close(subscriber.Channel)
		delete(sim.subscribers, id)
	}
}

// PublishUpdate publishes an intelligence update to the stream
func (sim *StreamingIntelligenceManager) PublishUpdate(update IntelligenceUpdate) {
	if !sim.running {
		return
	}

	update.Timestamp = time.Now()

	// Add predictive enhancements if available
	if PredictiveIntelligence != nil {
		update = sim.enhanceWithPredictions(update)
	}

	// Check throttling
	if sim.throttleManager.ShouldThrottle(update) {
		sim.analytics.recordThrottle()
		return
	}

	select {
	case sim.updateStream <- update:
		sim.analytics.recordUpdate(update)
	case <-sim.ctx.Done():
		return
	default:
		// Channel full, buffer or drop based on priority
		if update.Priority >= 8 {
			sim.updateBuffer.add(update)
		}
		sim.analytics.recordDrop()
	}
}

// processUpdates main update processing loop
func (sim *StreamingIntelligenceManager) processUpdates() {
	for {
		select {
		case <-sim.ctx.Done():
			return
		case update := <-sim.updateStream:
			sim.distributeUpdate(update)
		}
	}
}

// distributeUpdate distributes an update to relevant subscribers
func (sim *StreamingIntelligenceManager) distributeUpdate(update IntelligenceUpdate) {
	sim.mutex.RLock()
	defer sim.mutex.RUnlock()

	for _, subscriber := range sim.subscribers {
		if !subscriber.Active {
			continue
		}

		if sim.matchesFilters(update, subscriber.Filters) {
			select {
			case subscriber.Channel <- update:
				subscriber.Received++
				subscriber.LastUpdate = time.Now()
			default:
				// Subscriber channel full, drop update
				subscriber.Dropped++
			}
		}
	}

	// Send to main program if needed
	if sim.program != nil && sim.shouldNotifyProgram(update) {
		sim.program.Send(streamingIntelligenceUpdateMsg{update: update})
	}
}

// flushUpdates flushes buffered updates
func (sim *StreamingIntelligenceManager) flushUpdates(updates []IntelligenceUpdate) {
	for _, update := range updates {
		sim.distributeUpdate(update)
	}
}

// Helper functions and implementations

func NewStreamThrottleManager() *StreamThrottleManager {
	rules := make(map[IntelligenceUpdateType]*ThrottleRule)

	// Define default throttle rules
	rules[UpdateTypeRiskChange] = &ThrottleRule{
		MaxRate:       2.0,
		BurstSize:     3,
		CooldownTime:  1 * time.Second,
		AdaptiveRate:  true,
		PriorityBoost: 2.0,
	}

	rules[UpdateTypeNewSuggestion] = &ThrottleRule{
		MaxRate:       5.0,
		BurstSize:     10,
		CooldownTime:  500 * time.Millisecond,
		AdaptiveRate:  true,
		PriorityBoost: 1.5,
	}

	rules[UpdateTypeContextChange] = &ThrottleRule{
		MaxRate:       1.0,
		BurstSize:     2,
		CooldownTime:  2 * time.Second,
		AdaptiveRate:  false,
		PriorityBoost: 3.0,
	}

	return &StreamThrottleManager{
		rules:         rules,
		userActivity:  NewUserActivityTracker(),
		bandwidth:     NewBandwidthManager(),
		adaptiveRates: make(map[string]float64),
	}
}

func NewUserActivityTracker() *UserActivityTracker {
	return &UserActivityTracker{
		lastActivity:    time.Now(),
		activityLevel:   ActivityIdle,
		sessionStart:    time.Now(),
		interactionRate: 0.0,
	}
}

func NewBandwidthManager() *BandwidthManager {
	return &BandwidthManager{
		availableBandwidth: 50.0, // 50 updates per second default
		currentUsage:       0.0,
		qualitySettings: QualitySettings{
			MaxUpdatesPerSecond: 100,
			PriorityThreshold:   5,
			CompressionLevel:    3,
			BatchingEnabled:     true,
			PredictiveFiltering: true,
		},
		performanceMetrics: &PerformanceMetrics{
			LastMeasurement: time.Now(),
		},
	}
}

func NewUpdateBuffer(maxSize int, flushInterval time.Duration) *UpdateBuffer {
	return &UpdateBuffer{
		updates:       make([]IntelligenceUpdate, 0, maxSize),
		maxSize:       maxSize,
		flushTicker:   time.NewTicker(flushInterval),
		batchSize:     10,
		compression:   true,
		deduplication: true,
	}
}

func NewStreamAnalytics() *StreamAnalytics {
	return &StreamAnalytics{
		updatesByType:      make(map[IntelligenceUpdateType]int64),
		userSatisfaction:   0.8,
		predictionAccuracy: 0.75,
		cacheHitRate:       0.65,
	}
}

// ShouldThrottle determines if an update should be throttled
func (stm *StreamThrottleManager) ShouldThrottle(update IntelligenceUpdate) bool {
	stm.mutex.RLock()
	defer stm.mutex.RUnlock()

	rule, exists := stm.rules[update.Type]
	if !exists {
		return false
	}

	// High priority updates get special treatment
	if update.Priority >= 9 {
		return false
	}

	// Apply adaptive rate adjustment based on user activity
	effectiveRate := rule.MaxRate
	if rule.AdaptiveRate {
		activityMultiplier := stm.getActivityMultiplier()
		effectiveRate *= activityMultiplier
	}

	// Apply priority boost
	if update.Priority >= 7 {
		effectiveRate *= rule.PriorityBoost
	}

	// Check bandwidth constraints
	if !stm.bandwidth.hasCapacity(effectiveRate) {
		return true
	}

	return false
}

func (stm *StreamThrottleManager) getActivityMultiplier() float64 {
	stm.userActivity.mutex.RLock()
	defer stm.userActivity.mutex.RUnlock()

	switch stm.userActivity.activityLevel {
	case ActivityIdle:
		return 0.2
	case ActivityLow:
		return 0.5
	case ActivityModerate:
		return 1.0
	case ActivityHigh:
		return 1.5
	case ActivityIntense:
		return 2.0
	default:
		return 1.0
	}
}

func (bm *BandwidthManager) hasCapacity(requestedRate float64) bool {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	return bm.currentUsage+requestedRate <= bm.availableBandwidth
}

func (ub *UpdateBuffer) add(update IntelligenceUpdate) {
	ub.mutex.Lock()
	defer ub.mutex.Unlock()

	if len(ub.updates) >= ub.maxSize {
		// Remove oldest update
		ub.updates = ub.updates[1:]
	}

	ub.updates = append(ub.updates, update)
}

func (ub *UpdateBuffer) startFlushing(ctx context.Context, flushFunc func([]IntelligenceUpdate)) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-ub.flushTicker.C:
			ub.flush(flushFunc)
		}
	}
}

func (ub *UpdateBuffer) flush(flushFunc func([]IntelligenceUpdate)) {
	ub.mutex.Lock()
	defer ub.mutex.Unlock()

	if len(ub.updates) == 0 {
		return
	}

	updates := make([]IntelligenceUpdate, len(ub.updates))
	copy(updates, ub.updates)
	ub.updates = ub.updates[:0]

	go flushFunc(updates)
}

func (ub *UpdateBuffer) stop() {
	ub.flushTicker.Stop()
}

// Analytics methods
func (sa *StreamAnalytics) recordUpdate(update IntelligenceUpdate) {
	sa.mutex.Lock()
	defer sa.mutex.Unlock()

	sa.totalUpdates++
	sa.updatesByType[update.Type]++
}

func (sa *StreamAnalytics) recordThrottle() {
	sa.mutex.Lock()
	defer sa.mutex.Unlock()

	sa.throttleEvents++
}

func (sa *StreamAnalytics) recordDrop() {
	// Implementation for recording dropped updates
}

func (sa *StreamAnalytics) collectMetrics(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sa.updateMetrics()
		}
	}
}

func (sa *StreamAnalytics) updateMetrics() {
	// Implementation for updating performance metrics
}

func (stm *StreamThrottleManager) adaptiveAdjustment(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stm.adjustRates()
		}
	}
}

func (stm *StreamThrottleManager) adjustRates() {
	// Implementation for adaptive rate adjustment
}

// Filter matching
func (sim *StreamingIntelligenceManager) matchesFilters(update IntelligenceUpdate, filters []UpdateFilter) bool {
	if len(filters) == 0 {
		return true
	}

	for _, filter := range filters {
		if filter.Type != "" && filter.Type != update.Type {
			continue
		}
		if filter.MinPriority > 0 && update.Priority < filter.MinPriority {
			continue
		}
		if filter.ContextFilter != nil && !filter.ContextFilter(update.Context) {
			continue
		}
		if filter.CustomFilter != nil && !filter.CustomFilter(update) {
			continue
		}
		return true
	}
	return false
}

func (sim *StreamingIntelligenceManager) shouldNotifyProgram(update IntelligenceUpdate) bool {
	// Only notify program for high-priority updates or UI-relevant changes
	return update.Priority >= 7 ||
		   update.Type == UpdateTypeRiskChange ||
		   update.Type == UpdateTypeContextChange ||
		   update.Type == UpdateTypeHealthChange
}

// Message type for Bubble Tea
type streamingIntelligenceUpdateMsg struct {
	update IntelligenceUpdate
}

// UpdateUserActivity updates user activity tracking
func (sim *StreamingIntelligenceManager) UpdateUserActivity(activity ActivityLevel, typingSpeed float64) {
	sim.throttleManager.userActivity.mutex.Lock()
	defer sim.throttleManager.userActivity.mutex.Unlock()

	sim.throttleManager.userActivity.lastActivity = time.Now()
	sim.throttleManager.userActivity.activityLevel = activity
	sim.throttleManager.userActivity.typingSpeed = typingSpeed
	sim.throttleManager.userActivity.interactionRate = sim.calculateInteractionRate()
}

func (sim *StreamingIntelligenceManager) calculateInteractionRate() float64 {
	// Calculate interactions per minute based on recent activity
	return 10.0 // Placeholder implementation
}

// GetAnalytics returns current streaming analytics
func (sim *StreamingIntelligenceManager) GetAnalytics() *StreamAnalytics {
	return sim.analytics
}

// AdjustQuality adjusts streaming quality based on performance
func (sim *StreamingIntelligenceManager) AdjustQuality(settings QualitySettings) {
	sim.throttleManager.bandwidth.mutex.Lock()
	defer sim.throttleManager.bandwidth.mutex.Unlock()

	sim.throttleManager.bandwidth.qualitySettings = settings
}

// enhanceWithPredictions enhances updates with predictive intelligence
func (sim *StreamingIntelligenceManager) enhanceWithPredictions(update IntelligenceUpdate) IntelligenceUpdate {
	if update.Context == nil {
		return update
	}

	// Get predictions for current context
	predictions := PredictiveIntelligence.PredictNextActions(update.Context, "")
	
	// Add predictions to update data if relevant
	if len(predictions) > 0 {
		if updateData, ok := update.Data.(map[string]interface{}); ok {
			updateData["predictions"] = predictions
			updateData["prediction_count"] = len(predictions)
			
			// Calculate average prediction confidence
			totalConfidence := 0.0
			for _, pred := range predictions {
				totalConfidence += pred.Confidence
			}
			avgConfidence := totalConfidence / float64(len(predictions))
			updateData["prediction_confidence"] = avgConfidence
			
			update.Data = updateData
		} else {
			// Create new data structure with predictions
			update.Data = map[string]interface{}{
				"original_data": update.Data,
				"predictions": predictions,
				"prediction_count": len(predictions),
			}
		}
		
		// Increase priority if high-confidence predictions are available
		highConfidencePredictions := 0
		for _, pred := range predictions {
			if pred.Confidence > 0.8 {
				highConfidencePredictions++
			}
		}
		
		if highConfidencePredictions > 0 {
			update.Priority = minInt(10, update.Priority+2)
		}
	}

	return update
}

// StreamPredictiveUpdate streams a predictive intelligence update
func (sim *StreamingIntelligenceManager) StreamPredictiveUpdate(context *KubeContextSummary, userInput string) {
	if PredictiveIntelligence == nil {
		return
	}

	predictions := PredictiveIntelligence.PredictNextActions(context, userInput)
	if len(predictions) == 0 {
		return
	}

	update := IntelligenceUpdate{
		Type:      UpdateTypePrediction,
		Priority:  5,
		Source:    "predictive_intelligence",
		Data: map[string]interface{}{
			"predictions": predictions,
			"user_input": userInput,
			"context_hash": context.Hash(),
		},
		Context:    context,
		Confidence: sim.calculatePredictionConfidence(predictions),
		TTL:        2 * time.Minute,
	}

	sim.PublishUpdate(update)
}

// calculatePredictionConfidence calculates overall confidence for predictions
func (sim *StreamingIntelligenceManager) calculatePredictionConfidence(predictions []PredictedAction) float64 {
	if len(predictions) == 0 {
		return 0.0
	}

	totalConfidence := 0.0
	for _, pred := range predictions {
		totalConfidence += pred.Confidence
	}
	
	return totalConfidence / float64(len(predictions))
}

// StreamAdaptiveThrottling adjusts throttling based on user activity and system performance
func (sim *StreamingIntelligenceManager) StreamAdaptiveThrottling() {
	if PerformanceOptimizer == nil {
		return
	}

	// Get performance metrics
	stats := PerformanceOptimizer.GetPerformanceStats()
	
	// Adjust throttling based on system performance
	if cpuStats, ok := stats["cpu"].(map[string]interface{}); ok {
		if currentLoad, ok := cpuStats["current_load"].(float64); ok {
			if currentLoad > 0.8 {
				// High CPU load - increase throttling
				sim.throttleManager.baseRateLimit = maxInt(1, sim.throttleManager.baseRateLimit-1)
			} else if currentLoad < 0.3 {
				// Low CPU load - decrease throttling
				sim.throttleManager.baseRateLimit = minInt(20, sim.throttleManager.baseRateLimit+1)
			}
		}
	}

	// Adjust based on memory pressure
	if memStats, ok := stats["memory"].(map[string]interface{}); ok {
		if pressure, ok := memStats["pressure"].(float64); ok {
			if pressure > 0.8 {
				// High memory pressure - reduce update frequency
				sim.throttleManager.baseRateLimit = maxInt(1, sim.throttleManager.baseRateLimit-2)
			}
		}
	}
}

// Helper functions
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}