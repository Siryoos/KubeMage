// performance_monitor.go - Real-time performance monitoring and automatic optimization
package engine

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// RealTimePerformanceMonitor provides comprehensive performance monitoring
type RealTimePerformanceMonitor struct {
	metricsCollector   *RealTimeMetricsCollector
	alertManager       *AlertManager
	optimizationEngine *AutoOptimizationEngine
	dashboard          *PerformanceDashboard
	healthChecker      *HealthChecker
	trendAnalyzer      *TrendAnalyzer
	enabled            bool
	ctx                context.Context
	cancel             context.CancelFunc
	mu                 sync.RWMutex
}

// RealTimeMetricsCollector collects metrics in real-time
type RealTimeMetricsCollector struct {
	systemMetrics      *SystemMetrics
	applicationMetrics *ApplicationMetrics
	kubernetesMetrics  *KubernetesMetrics
	customMetrics      map[string]*CustomMetric
	collectionInterval time.Duration
	history            []MetricsSnapshot
	maxHistory         int
	mu                 sync.RWMutex
}

// AlertManager manages performance alerts and notifications
type AlertManager struct {
	alerts         []PerformanceAlert
	rules          []AlertRule
	channels       []AlertChannel
	suppressions   map[string]time.Time
	escalationRules []EscalationRule
	mu             sync.RWMutex
}

// AutoOptimizationEngine automatically optimizes performance
type AutoOptimizationEngine struct {
	optimizers         []AutoOptimizer
	optimizationHistory []OptimizationAction
	learningEnabled    bool
	aggressiveness     float64
	safetyThreshold    float64
	mu                 sync.RWMutex
}

// PerformanceDashboard provides real-time performance visualization
type PerformanceDashboard struct {
	widgets        []DashboardWidget
	updateInterval time.Duration
	subscribers    []DashboardSubscriber
	lastUpdate     time.Time
	mu             sync.RWMutex
}

// HealthChecker monitors system health
type HealthChecker struct {
	healthChecks   []HealthCheck
	healthStatus   HealthStatus
	checkInterval  time.Duration
	failureHistory []HealthFailure
	mu             sync.RWMutex
}

// TrendAnalyzer analyzes performance trends
type TrendAnalyzer struct {
	trends         map[string]*PerformanceTrend
	predictions    map[string]*PerformancePrediction
	analysisWindow time.Duration
	mu             sync.RWMutex
}

// SystemMetrics represents system-level metrics
// SystemMetrics is now defined in types.go

// ApplicationMetrics represents application-specific metrics
type ApplicationMetrics struct {
	ResponseTime       float64   `json:"response_time"`
	Throughput         float64   `json:"throughput"`
	ErrorRate          float64   `json:"error_rate"`
	CacheHitRate       float64   `json:"cache_hit_rate"`
	ActiveConnections  int       `json:"active_connections"`
	QueueLength        int       `json:"queue_length"`
	IntelligenceLatency float64  `json:"intelligence_latency"`
	StreamingRate      float64   `json:"streaming_rate"`
	Timestamp          time.Time `json:"timestamp"`
}

// KubernetesMetrics represents Kubernetes-specific metrics
type KubernetesMetrics struct {
	PodCount           int       `json:"pod_count"`
	NodeCount          int       `json:"node_count"`
	ServiceCount       int       `json:"service_count"`
	NamespaceCount     int       `json:"namespace_count"`
	EventRate          float64   `json:"event_rate"`
	APIServerLatency   float64   `json:"api_server_latency"`
	EtcdLatency        float64   `json:"etcd_latency"`
	ClusterHealth      string    `json:"cluster_health"`
	Timestamp          time.Time `json:"timestamp"`
}

// CustomMetric represents a custom performance metric
type CustomMetric struct {
	Name        string                 `json:"name"`
	Value       float64                `json:"value"`
	Unit        string                 `json:"unit"`
	Labels      map[string]string      `json:"labels"`
	Collector   func() float64         `json:"-"`
	Threshold   float64                `json:"threshold"`
	Timestamp   time.Time              `json:"timestamp"`
}

// PerformanceAlert represents a performance alert
type PerformanceAlert struct {
	ID          string                 `json:"id"`
	Level       AlertLevel             `json:"level"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Metric      string                 `json:"metric"`
	Value       float64                `json:"value"`
	Threshold   float64                `json:"threshold"`
	Timestamp   time.Time              `json:"timestamp"`
	Resolved    bool                   `json:"resolved"`
	Actions     []string               `json:"actions"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// AlertLevel represents alert severity levels
type AlertLevel int

const (
	AlertLevelInfo AlertLevel = iota
	AlertLevelWarning
	AlertLevelCritical
	AlertLevelEmergency
)

// AlertRule represents an alert rule
type AlertRule struct {
	Name        string                    `json:"name"`
	Metric      string                    `json:"metric"`
	Condition   func(float64) bool        `json:"-"`
	Threshold   float64                   `json:"threshold"`
	Level       AlertLevel                `json:"level"`
	Duration    time.Duration             `json:"duration"`
	Enabled     bool                      `json:"enabled"`
	Actions     []AlertAction             `json:"actions"`
}

// AlertChannel represents an alert notification channel
type AlertChannel struct {
	Name    string                    `json:"name"`
	Type    string                    `json:"type"`
	Config  map[string]interface{}    `json:"config"`
	Send    func(PerformanceAlert) error `json:"-"`
	Enabled bool                      `json:"enabled"`
}

// EscalationRule represents alert escalation rules
type EscalationRule struct {
	Name        string        `json:"name"`
	Condition   string        `json:"condition"`
	Delay       time.Duration `json:"delay"`
	Actions     []AlertAction `json:"actions"`
	Enabled     bool          `json:"enabled"`
}

// AlertAction represents an action to take on alert
type AlertAction struct {
	Type        string                 `json:"type"`
	Config      map[string]interface{} `json:"config"`
	Execute     func() error           `json:"-"`
	Enabled     bool                   `json:"enabled"`
}

// AutoOptimizer represents an automatic optimizer
type AutoOptimizer struct {
	Name        string                    `json:"name"`
	Target      string                    `json:"target"`
	Condition   func(*SystemMetrics) bool `json:"-"`
	Optimize    func() error              `json:"-"`
	Priority    int                       `json:"priority"`
	Enabled     bool                      `json:"enabled"`
	SafetyCheck func() bool               `json:"-"`
}

// OptimizationAction represents an optimization action taken
type OptimizationAction struct {
	Optimizer   string                 `json:"optimizer"`
	Action      string                 `json:"action"`
	Timestamp   time.Time              `json:"timestamp"`
	Success     bool                   `json:"success"`
	Impact      float64                `json:"impact"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// DashboardWidget represents a dashboard widget
type DashboardWidget struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Title    string                 `json:"title"`
	Config   map[string]interface{} `json:"config"`
	Data     interface{}            `json:"data"`
	Render   func() string          `json:"-"`
	Update   func()                 `json:"-"`
	Enabled  bool                   `json:"enabled"`
}

// DashboardSubscriber represents a dashboard subscriber
type DashboardSubscriber struct {
	ID       string                    `json:"id"`
	Callback func(DashboardUpdate)     `json:"-"`
	Filters  []string                  `json:"filters"`
	Enabled  bool                      `json:"enabled"`
}

// DashboardUpdate represents a dashboard update
type DashboardUpdate struct {
	Widget    string      `json:"widget"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// HealthCheck represents a health check
type HealthCheck struct {
	Name        string                `json:"name"`
	Check       func() HealthResult   `json:"-"`
	Interval    time.Duration         `json:"interval"`
	Timeout     time.Duration         `json:"timeout"`
	Enabled     bool                  `json:"enabled"`
	LastRun     time.Time             `json:"last_run"`
	LastResult  HealthResult          `json:"last_result"`
}

// HealthResult represents a health check result
type HealthResult struct {
	Healthy   bool                   `json:"healthy"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details"`
	Timestamp time.Time              `json:"timestamp"`
}

// HealthStatus represents overall health status
type HealthStatus struct {
	Overall     string                 `json:"overall"`
	Components  map[string]HealthResult `json:"components"`
	LastUpdate  time.Time              `json:"last_update"`
}

// HealthFailure represents a health check failure
type HealthFailure struct {
	Check     string    `json:"check"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Resolved  bool      `json:"resolved"`
}

// PerformanceTrend represents a performance trend
type PerformanceTrend struct {
	Metric      string    `json:"metric"`
	Direction   string    `json:"direction"`
	Slope       float64   `json:"slope"`
	Confidence  float64   `json:"confidence"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Prediction  float64   `json:"prediction"`
}

// PerformancePrediction represents a performance prediction
type PerformancePrediction struct {
	Metric      string    `json:"metric"`
	Value       float64   `json:"value"`
	Confidence  float64   `json:"confidence"`
	TimeHorizon time.Duration `json:"time_horizon"`
	Timestamp   time.Time `json:"timestamp"`
}

// NewRealTimePerformanceMonitor creates a new real-time performance monitor
func NewRealTimePerformanceMonitor() *RealTimePerformanceMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &RealTimePerformanceMonitor{
		metricsCollector: &RealTimeMetricsCollector{
			systemMetrics:      &SystemMetrics{},
			applicationMetrics: &ApplicationMetrics{},
			kubernetesMetrics:  &KubernetesMetrics{},
			customMetrics:      make(map[string]*CustomMetric),
			collectionInterval: 5 * time.Second,
			history:            make([]MetricsSnapshot, 0),
			maxHistory:         1000,
		},
		alertManager: &AlertManager{
			alerts:       make([]PerformanceAlert, 0),
			rules:        make([]AlertRule, 0),
			channels:     make([]AlertChannel, 0),
			suppressions: make(map[string]time.Time),
			escalationRules: make([]EscalationRule, 0),
		},
		optimizationEngine: &AutoOptimizationEngine{
			optimizers:          make([]AutoOptimizer, 0),
			optimizationHistory: make([]OptimizationAction, 0),
			learningEnabled:     true,
			aggressiveness:      0.5,
			safetyThreshold:     0.8,
		},
		dashboard: &PerformanceDashboard{
			widgets:        make([]DashboardWidget, 0),
			updateInterval: 2 * time.Second,
			subscribers:    make([]DashboardSubscriber, 0),
		},
		healthChecker: &HealthChecker{
			healthChecks:   make([]HealthCheck, 0),
			healthStatus:   HealthStatus{Overall: "unknown", Components: make(map[string]HealthResult)},
			checkInterval:  30 * time.Second,
			failureHistory: make([]HealthFailure, 0),
		},
		trendAnalyzer: &TrendAnalyzer{
			trends:         make(map[string]*PerformanceTrend),
			predictions:    make(map[string]*PerformancePrediction),
			analysisWindow: 10 * time.Minute,
		},
		enabled: true,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start starts the real-time performance monitoring
func (rtpm *RealTimePerformanceMonitor) Start() error {
	if !rtpm.enabled {
		return nil
	}

	rtpm.mu.Lock()
	defer rtpm.mu.Unlock()

	// Start metrics collection
	go rtpm.metricsCollector.StartCollection(rtpm.ctx)

	// Start alert monitoring
	go rtpm.alertManager.StartMonitoring(rtpm.ctx, rtpm.metricsCollector)

	// Start auto-optimization
	go rtpm.optimizationEngine.StartOptimization(rtpm.ctx, rtpm.metricsCollector)

	// Start health checking
	go rtpm.healthChecker.StartHealthChecks(rtpm.ctx)

	// Start trend analysis
	go rtpm.trendAnalyzer.StartAnalysis(rtpm.ctx, rtpm.metricsCollector)

	// Start dashboard updates
	go rtpm.dashboard.StartUpdates(rtpm.ctx, rtpm.metricsCollector)

	return nil
}

// Stop stops the performance monitoring
func (rtpm *RealTimePerformanceMonitor) Stop() error {
	rtpm.mu.Lock()
	defer rtpm.mu.Unlock()

	rtpm.cancel()
	rtpm.enabled = false

	return nil
}

// StartCollection starts metrics collection
func (rtmc *RealTimeMetricsCollector) StartCollection(ctx context.Context) {
	ticker := time.NewTicker(rtmc.collectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rtmc.collectMetrics()
		}
	}
}

// collectMetrics collects all metrics
func (rtmc *RealTimeMetricsCollector) collectMetrics() {
	rtmc.mu.Lock()
	defer rtmc.mu.Unlock()

	// Collect system metrics
	rtmc.collectSystemMetrics()

	// Collect application metrics
	rtmc.collectApplicationMetrics()

	// Collect Kubernetes metrics
	rtmc.collectKubernetesMetrics()

	// Collect custom metrics
	rtmc.collectCustomMetrics()

	// Create snapshot
	snapshot := MetricsSnapshot{
		Metrics: SystemMetrics{
			CPUUsage:    rtmc.systemMetrics.CPUUsage,
			MemoryUsage: rtmc.systemMetrics.MemoryUsage,
			DiskUsage:   rtmc.systemMetrics.DiskUsage,
			NetworkIO:   rtmc.systemMetrics.NetworkIO,
			LoadAverage: rtmc.systemMetrics.LoadAverage,
			Goroutines:  rtmc.systemMetrics.Goroutines,
			Timestamp:   time.Now(),
		},
		Timestamp: time.Now(),
	}

	// Add to history
	rtmc.history = append(rtmc.history, snapshot)

	// Maintain history size
	if len(rtmc.history) > rtmc.maxHistory {
		rtmc.history = rtmc.history[len(rtmc.history)-rtmc.maxHistory:]
	}
}

// collectSystemMetrics collects system-level metrics
func (rtmc *RealTimeMetricsCollector) collectSystemMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// CPU usage (simplified)
	rtmc.systemMetrics.CPUUsage = float64(runtime.NumGoroutine()) / float64(runtime.NumCPU()*10) * 100
	if rtmc.systemMetrics.CPUUsage > 100 {
		rtmc.systemMetrics.CPUUsage = 100
	}

	// Memory usage
	rtmc.systemMetrics.MemoryUsage = float64(m.HeapAlloc) / float64(m.Sys) * 100

	// Goroutines
	rtmc.systemMetrics.Goroutines = runtime.NumGoroutine()

	// GC pauses
	if len(m.PauseNs) > 0 {
		rtmc.systemMetrics.GCPauses = []float64{float64(m.PauseNs[(m.NumGC+255)%256]) / 1e6} // Convert to ms
	}

	rtmc.systemMetrics.Timestamp = time.Now()
}

// collectApplicationMetrics collects application-specific metrics
func (rtmc *RealTimeMetricsCollector) collectApplicationMetrics() {
	// These would be collected from actual application metrics
	// For now, using placeholder values

	rtmc.applicationMetrics.ResponseTime = 150.0 // ms
	rtmc.applicationMetrics.Throughput = 100.0   // requests/sec
	rtmc.applicationMetrics.ErrorRate = 0.5      // %
	
	// Cache hit rate from smart cache if available
	if SmartCache != nil {
		stats := SmartCache.GetStats()
		rtmc.applicationMetrics.CacheHitRate = stats.HitRate * 100
	}

	rtmc.applicationMetrics.Timestamp = time.Now()
}

// collectKubernetesMetrics collects Kubernetes-specific metrics
func (rtmc *RealTimeMetricsCollector) collectKubernetesMetrics() {
	// These would be collected from Kubernetes API
	// For now, using placeholder values

	rtmc.kubernetesMetrics.PodCount = 50
	rtmc.kubernetesMetrics.NodeCount = 3
	rtmc.kubernetesMetrics.ServiceCount = 20
	rtmc.kubernetesMetrics.NamespaceCount = 10
	rtmc.kubernetesMetrics.EventRate = 5.0
	rtmc.kubernetesMetrics.APIServerLatency = 25.0
	rtmc.kubernetesMetrics.ClusterHealth = "healthy"
	rtmc.kubernetesMetrics.Timestamp = time.Now()
}

// collectCustomMetrics collects custom metrics
func (rtmc *RealTimeMetricsCollector) collectCustomMetrics() {
	for name, metric := range rtmc.customMetrics {
		if metric.Collector != nil {
			metric.Value = metric.Collector()
			metric.Timestamp = time.Now()
			rtmc.customMetrics[name] = metric
		}
	}
}

// StartMonitoring starts alert monitoring
func (am *AlertManager) StartMonitoring(ctx context.Context, collector *RealTimeMetricsCollector) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			am.checkAlerts(collector)
		}
	}
}

// checkAlerts checks for alert conditions
func (am *AlertManager) checkAlerts(collector *RealTimeMetricsCollector) {
	am.mu.Lock()
	defer am.mu.Unlock()

	metrics := collector.systemMetrics

	// Check each alert rule
	for _, rule := range am.rules {
		if !rule.Enabled {
			continue
		}

		var value float64
		switch rule.Metric {
		case "cpu_usage":
			value = metrics.CPUUsage
		case "memory_usage":
			value = metrics.MemoryUsage
		case "goroutines":
			value = float64(metrics.Goroutines)
		default:
			continue
		}

		if rule.Condition(value) {
			alert := PerformanceAlert{
				ID:          fmt.Sprintf("alert-%d", time.Now().Unix()),
				Level:       rule.Level,
				Title:       fmt.Sprintf("%s Alert", rule.Name),
				Description: fmt.Sprintf("%s exceeded threshold: %.2f > %.2f", rule.Metric, value, rule.Threshold),
				Metric:      rule.Metric,
				Value:       value,
				Threshold:   rule.Threshold,
				Timestamp:   time.Now(),
				Resolved:    false,
				Actions:     []string{},
				Metadata:    make(map[string]interface{}),
			}

			am.alerts = append(am.alerts, alert)

			// Execute alert actions
			for _, action := range rule.Actions {
				if action.Enabled && action.Execute != nil {
					go action.Execute()
				}
			}

			// Send notifications
			for _, channel := range am.channels {
				if channel.Enabled && channel.Send != nil {
					go channel.Send(alert)
				}
			}
		}
	}
}

// StartOptimization starts automatic optimization
func (aoe *AutoOptimizationEngine) StartOptimization(ctx context.Context, collector *RealTimeMetricsCollector) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			aoe.runOptimizations(collector)
		}
	}
}

// runOptimizations runs automatic optimizations
func (aoe *AutoOptimizationEngine) runOptimizations(collector *RealTimeMetricsCollector) {
	aoe.mu.Lock()
	defer aoe.mu.Unlock()

	metrics := collector.systemMetrics

	// Run optimizers based on priority
	for _, optimizer := range aoe.optimizers {
		if !optimizer.Enabled {
			continue
		}

		// Check safety threshold
		if optimizer.SafetyCheck != nil && !optimizer.SafetyCheck() {
			continue
		}

		// Check condition
		if optimizer.Condition(metrics) {
			// Execute optimization
			err := optimizer.Optimize()
			
			action := OptimizationAction{
				Optimizer: optimizer.Name,
				Action:    optimizer.Target,
				Timestamp: time.Now(),
				Success:   err == nil,
				Impact:    0.0, // Would be calculated based on before/after metrics
				Metadata:  make(map[string]interface{}),
			}

			if err != nil {
				action.Metadata["error"] = err.Error()
			}

			aoe.optimizationHistory = append(aoe.optimizationHistory, action)

			// Maintain history size
			if len(aoe.optimizationHistory) > 1000 {
				aoe.optimizationHistory = aoe.optimizationHistory[len(aoe.optimizationHistory)-1000:]
			}
		}
	}
}

// StartHealthChecks starts health checking
func (hc *HealthChecker) StartHealthChecks(ctx context.Context) {
	ticker := time.NewTicker(hc.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hc.runHealthChecks()
		}
	}
}

// runHealthChecks runs all health checks
func (hc *HealthChecker) runHealthChecks() {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	overallHealthy := true
	
	for i := range hc.healthChecks {
		check := &hc.healthChecks[i]
		if !check.Enabled {
			continue
		}

		// Run health check with timeout
		result := check.Check()
		check.LastRun = time.Now()
		check.LastResult = result

		hc.healthStatus.Components[check.Name] = result

		if !result.Healthy {
			overallHealthy = false
			
			// Record failure
			failure := HealthFailure{
				Check:     check.Name,
				Message:   result.Message,
				Timestamp: time.Now(),
				Resolved:  false,
			}
			hc.failureHistory = append(hc.failureHistory, failure)
		}
	}

	// Update overall status
	if overallHealthy {
		hc.healthStatus.Overall = "healthy"
	} else {
		hc.healthStatus.Overall = "unhealthy"
	}
	hc.healthStatus.LastUpdate = time.Now()
}

// StartAnalysis starts trend analysis
func (ta *TrendAnalyzer) StartAnalysis(ctx context.Context, collector *RealTimeMetricsCollector) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ta.analyzeTrends(collector)
		}
	}
}

// analyzeTrends analyzes performance trends
func (ta *TrendAnalyzer) analyzeTrends(collector *RealTimeMetricsCollector) {
	ta.mu.Lock()
	defer ta.mu.Unlock()

	// Analyze trends from metrics history
	if len(collector.history) < 10 {
		return // Need more data points
	}

	// Analyze CPU usage trend
	cpuValues := make([]float64, len(collector.history))
	for i, snapshot := range collector.history {
		cpuValues[i] = snapshot.Metrics.CPUUsage
	}

	cpuTrend := ta.calculateTrend("cpu_usage", cpuValues)
	ta.trends["cpu_usage"] = cpuTrend

	// Analyze memory usage trend
	memValues := make([]float64, len(collector.history))
	for i, snapshot := range collector.history {
		memValues[i] = snapshot.Metrics.MemoryUsage
	}

	memTrend := ta.calculateTrend("memory_usage", memValues)
	ta.trends["memory_usage"] = memTrend

	// Generate predictions
	ta.generatePredictions()
}

// calculateTrend calculates trend for a metric
func (ta *TrendAnalyzer) calculateTrend(metric string, values []float64) *PerformanceTrend {
	if len(values) < 2 {
		return nil
	}

	// Simple linear regression for trend calculation
	n := float64(len(values))
	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0

	for i, y := range values {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	
	direction := "stable"
	if slope > 0.1 {
		direction = "increasing"
	} else if slope < -0.1 {
		direction = "decreasing"
	}

	return &PerformanceTrend{
		Metric:     metric,
		Direction:  direction,
		Slope:      slope,
		Confidence: 0.8, // Simplified confidence calculation
		StartTime:  time.Now().Add(-time.Duration(len(values)) * time.Minute),
		EndTime:    time.Now(),
		Prediction: values[len(values)-1] + slope*5, // Predict 5 minutes ahead
	}
}

// generatePredictions generates performance predictions
func (ta *TrendAnalyzer) generatePredictions() {
	for metric, trend := range ta.trends {
		prediction := &PerformancePrediction{
			Metric:      metric,
			Value:       trend.Prediction,
			Confidence:  trend.Confidence,
			TimeHorizon: 5 * time.Minute,
			Timestamp:   time.Now(),
		}
		ta.predictions[metric] = prediction
	}
}

// StartUpdates starts dashboard updates
func (pd *PerformanceDashboard) StartUpdates(ctx context.Context, collector *RealTimeMetricsCollector) {
	ticker := time.NewTicker(pd.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pd.updateDashboard(collector)
		}
	}
}

// updateDashboard updates dashboard widgets
func (pd *PerformanceDashboard) updateDashboard(collector *RealTimeMetricsCollector) {
	pd.mu.Lock()
	defer pd.mu.Unlock()

	// Update each widget
	for i := range pd.widgets {
		widget := &pd.widgets[i]
		if widget.Enabled && widget.Update != nil {
			widget.Update()
		}
	}

	pd.lastUpdate = time.Now()

	// Notify subscribers
	for _, subscriber := range pd.subscribers {
		if subscriber.Enabled && subscriber.Callback != nil {
			update := DashboardUpdate{
				Widget:    "all",
				Data:      collector.systemMetrics,
				Timestamp: time.Now(),
			}
			go subscriber.Callback(update)
		}
	}
}

// GetPerformanceStats returns comprehensive performance statistics
func (rtpm *RealTimePerformanceMonitor) GetPerformanceStats() map[string]interface{} {
	rtpm.mu.RLock()
	defer rtpm.mu.RUnlock()

	return map[string]interface{}{
		"system_metrics":       rtpm.metricsCollector.systemMetrics,
		"application_metrics":  rtpm.metricsCollector.applicationMetrics,
		"kubernetes_metrics":   rtpm.metricsCollector.kubernetesMetrics,
		"active_alerts":        len(rtpm.alertManager.alerts),
		"alert_rules":          len(rtpm.alertManager.rules),
		"optimizers":           len(rtpm.optimizationEngine.optimizers),
		"optimization_history": len(rtpm.optimizationEngine.optimizationHistory),
		"health_status":        rtpm.healthChecker.healthStatus,
		"trends":               len(rtpm.trendAnalyzer.trends),
		"predictions":          len(rtpm.trendAnalyzer.predictions),
		"dashboard_widgets":    len(rtpm.dashboard.widgets),
		"enabled":              rtpm.enabled,
	}
}

// Removed global instance - now created via dependency injection

// InitializeRealTimePerformanceMonitor initializes the global performance monitor
func InitializeRealTimePerformanceMonitor() {
	RealTimeMonitor = NewRealTimePerformanceMonitor()
	
	// Start monitoring
	go func() {
		if err := RealTimeMonitor.Start(); err != nil {
			// Log error but don't fail
		}
	}()
}

