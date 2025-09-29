// intelligence.go - Intelligent analysis coordination and smart decision making
package engine

import (
	"fmt"
	"sort"
	"strings"
	"time"
	
	tea "github.com/charmbracelet/bubbletea"
)

// IntelligenceEngine coordinates all smart analysis components
type IntelligenceEngine struct {
	facts       *FactHelper
	knowledge   *PlaybookLibrary
	optimizer   *OptimizationAdvisor
	router      *IntentRouter
	history     []AnalysisSession
	processor   *AsyncIntelligenceProcessor
	predictive  *PredictiveEngine
	cache       *SmartCacheSystem
}

// AnalysisSession represents a complete intelligence analysis session
type AnalysisSession struct {
	ID           string              `json:"id"`
	Timestamp    time.Time           `json:"timestamp"`
	Context      *KubeContextSummary `json:"context"`
	Intent       *IntentRouter       `json:"intent"`
	RootCause    *RootCauseAnalysis  `json:"root_cause"`
	Optimization []Recommendation    `json:"optimization"`
	Confidence   float64             `json:"confidence"`
	Actions      []IntelligentAction `json:"actions"`
	Outcome      string              `json:"outcome"`
}

// IntelligentAction represents a smart, context-aware action
type IntelligentAction struct {
	Type        string    `json:"type"`     // "diagnostic", "fix", "optimize", "explain"
	Priority    int       `json:"priority"` // 1-10 (10 = highest)
	Description string    `json:"description"`
	Command     string    `json:"command"`
	Risk        RiskLevel `json:"risk"`
	Expected    string    `json:"expected"`
	Automated   bool      `json:"automated"`
	PreChecks   []string  `json:"pre_checks"`
}

// RiskLevel represents intelligent risk assessment
type RiskLevel struct {
	Level       string   `json:"level"`       // "low", "medium", "high", "critical"
	Factors     []string `json:"factors"`     // Why this risk level
	Mitigations []string `json:"mitigations"` // How to reduce risk
	Reversible  bool     `json:"reversible"`  // Can be undone
}

// IntelligentInsight represents a smart observation or recommendation
type IntelligentInsight struct {
	Type        string   `json:"type"` // "pattern", "anomaly", "optimization", "prediction"
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Confidence  float64  `json:"confidence"`
	Impact      string   `json:"impact"`
	Evidence    []string `json:"evidence"`
	NextSteps   []string `json:"next_steps"`
}

// NewIntelligenceEngine creates a new intelligence coordination engine
func NewIntelligenceEngine() *IntelligenceEngine {
	return &IntelligenceEngine{
		facts:     Facts,
		knowledge: Knowledge,
		optimizer: Optimizer,
		history:   make([]AnalysisSession, 0),
		processor: nil, // Will be initialized when TUI starts
		predictive: PredictiveIntelligence,
		cache:     NewSmartCacheSystem(100, 500, 2*time.Hour), // L1: 100 items, L2: 500 items, L3: 2h max age
	}
}

// InitializeAsyncProcessor initializes the async processing system
func (ie *IntelligenceEngine) InitializeAsyncProcessor(program *tea.Program) {
	ie.processor = NewAsyncIntelligenceProcessor(3, program) // 3 workers
	ie.processor.Start()
}

// StopAsyncProcessor stops the async processing system
func (ie *IntelligenceEngine) StopAsyncProcessor() {
	if ie.processor != nil {
		ie.processor.Stop()
	}
}

// AnalyzeIntelligentlyAsync submits intelligence analysis for async processing
func (ie *IntelligenceEngine) AnalyzeIntelligentlyAsync(input string, context *KubeContextSummary, callback func(IntelligenceResult)) string {
	if ie.processor == nil {
		// Fallback to sync processing
		session, err := ie.AnalyzeIntelligently(input, context)
		if callback != nil {
			result := IntelligenceResult{
				ID:      fmt.Sprintf("sync-%d", time.Now().Unix()),
				Type:    WorkTypeAnalysis,
				Success: err == nil,
				Data:    session,
				Error:   err,
			}
			callback(result)
		}
		return result.ID
	}

	workID := fmt.Sprintf("analysis-%d", time.Now().Unix())
	work := IntelligenceWork{
		ID:         workID,
		Type:       WorkTypeAnalysis,
		Priority:   5, // Medium priority
		Input:      input,
		Context:    context,
		Callback:   callback,
		Timeout:    10 * time.Second,
		MaxRetries: 2,
	}

	ie.processor.SubmitWork(work)
	return workID
}

// AnalyzeIntelligentlyHighPriority submits high-priority intelligence analysis
func (ie *IntelligenceEngine) AnalyzeIntelligentlyHighPriority(input string, context *KubeContextSummary, callback func(IntelligenceResult)) string {
	if ie.processor == nil {
		// Fallback to sync processing
		session, err := ie.AnalyzeIntelligently(input, context)
		if callback != nil {
			result := IntelligenceResult{
				ID:      fmt.Sprintf("sync-hp-%d", time.Now().Unix()),
				Type:    WorkTypeAnalysis,
				Success: err == nil,
				Data:    session,
				Error:   err,
			}
			callback(result)
		}
		return result.ID
	}

	workID := fmt.Sprintf("analysis-hp-%d", time.Now().Unix())
	work := IntelligenceWork{
		ID:         workID,
		Type:       WorkTypeAnalysis,
		Priority:   10, // Highest priority
		Input:      input,
		Context:    context,
		Callback:   callback,
		Timeout:    5 * time.Second,
		MaxRetries: 1,
	}

	ie.processor.SubmitHighPriorityWork(work)
	return workID
}

// AnalyzeIntelligently performs comprehensive intelligent analysis
func (ie *IntelligenceEngine) AnalyzeIntelligently(input string, context *KubeContextSummary) (*AnalysisSession, error) {
	sessionID := fmt.Sprintf("analysis-%d", time.Now().Unix())
	session := &AnalysisSession{
		ID:        sessionID,
		Timestamp: time.Now(),
		Context:   context,
	}

	// Step 1: Route intent intelligently
	router := RouteIntent(input, context)
	session.Intent = &router

	// Step 2: Gather relevant observations based on intent
	observations, err := ie.gatherIntelligentObservations(router.Mode, context)
	if err != nil {
		return nil, fmt.Errorf("failed to gather observations: %w", err)
	}

	// Step 3: Perform root cause analysis if diagnostic mode
	if router.Mode == ModeDiagnose && len(observations) > 0 {
		rootCause := ie.knowledge.DetectRootCause(observations)
		session.RootCause = rootCause
		session.Confidence = rootCause.Confidence
	}

	// Step 4: Generate optimization recommendations
	if context != nil && context.Namespace != "" {
		optimizations, err := ie.optimizer.AnalyzeResourceUtilization(context.Namespace)
		if err == nil {
			session.Optimization = optimizations
		}
	}

	// Step 5: Generate intelligent actions
	actions := ie.generateIntelligentActions(session)
	session.Actions = actions

	// Step 6: Calculate overall confidence
	session.Confidence = ie.calculateOverallConfidence(session)

	// Store session for learning
	ie.history = append(ie.history, *session)

	return session, nil
}

// gatherIntelligentObservations collects relevant data based on analysis mode
func (ie *IntelligenceEngine) gatherIntelligentObservations(mode AgentMode, context *KubeContextSummary) ([]string, error) {
	var observations []string

	if context == nil || context.Namespace == "" {
		return observations, nil
	}

	switch mode {
	case ModeDiagnose:
		// Gather diagnostic observations
		if pods, err := ie.facts.PodsSummary(context.Namespace); err == nil {
			for _, problem := range pods.TopProblems {
				observations = append(observations, fmt.Sprintf("Pod problem: %s (count: %d, example: %s)",
					problem.Reason, problem.Count, problem.Sample))
			}
		}

		// Check for deployment issues
		for i := 0; i < 5; i++ { // Check up to 5 deployments
			if deploy, err := ie.facts.DeployProgress(context.Namespace, fmt.Sprintf("deployment-%d", i)); err == nil && deploy.Name != "" {
				if deploy.ReadyRatio < 1.0 {
					observations = append(observations, fmt.Sprintf("Deployment %s not fully ready: %.1f%% (last event: %s)",
						deploy.Name, deploy.ReadyRatio*100, deploy.LastEvent))
				}
			}
		}

	case ModeGenerate:
		// Gather context for generation
		if quota, err := ie.facts.QuotaSnapshot(context.Namespace); err == nil {
			observations = append(observations, fmt.Sprintf("Namespace health: %s", quota.OverallHealth))
		}

	case ModeEdit:
		// Gather current configuration state
		if workspace, err := ie.facts.IndexWorkspace(); err == nil {
			observations = append(observations, fmt.Sprintf("Workspace contains %d charts, %d templates, %d manifests",
				len(workspace.Charts), len(workspace.Templates), len(workspace.Manifests)))
		}
	}

	return observations, nil
}

// generateIntelligentActions creates smart, prioritized actions
func (ie *IntelligenceEngine) generateIntelligentActions(session *AnalysisSession) []IntelligentAction {
	var actions []IntelligentAction

	// Generate actions based on intent
	switch session.Intent.Mode {
	case ModeDiagnose:
		actions = append(actions, ie.generateDiagnosticActions(session)...)
	case ModeGenerate:
		actions = append(actions, ie.generateCreationActions(session)...)
	case ModeEdit:
		actions = append(actions, ie.generateEditActions(session)...)
	case ModeExplain:
		actions = append(actions, ie.generateExplanationActions(session)...)
	case ModeCommand:
		actions = append(actions, ie.generateCommandActions(session)...)
	}

	// Add optimization actions if recommendations exist
	if len(session.Optimization) > 0 {
		actions = append(actions, ie.generateOptimizationActions(session)...)
	}

	// Sort actions by priority
	sort.Slice(actions, func(i, j int) bool {
		return actions[i].Priority > actions[j].Priority
	})

	return actions
}

// generateDiagnosticActions creates intelligent diagnostic actions
func (ie *IntelligenceEngine) generateDiagnosticActions(session *AnalysisSession) []IntelligentAction {
	var actions []IntelligentAction

	if session.RootCause != nil && session.RootCause.RootCause != "Unknown" {
		// High-priority targeted diagnostic based on root cause
		actions = append(actions, IntelligentAction{
			Type:        "diagnostic",
			Priority:    9,
			Description: fmt.Sprintf("Investigate %s (confidence: %.1f%%)", session.RootCause.RootCause, session.RootCause.Confidence*100),
			Command:     ie.generateTargetedDiagnosticCommand(session.RootCause),
			Risk:        RiskLevel{Level: "low", Factors: []string{"read-only operation"}, Reversible: true},
			Expected:    fmt.Sprintf("Evidence of %s", session.RootCause.RootCause),
			Automated:   true,
			PreChecks:   []string{"verify namespace access", "check kubectl connectivity"},
		})

		// Add specific actions from root cause analysis
		for i, step := range session.RootCause.NextSteps {
			priority := 8 - i // Decreasing priority
			if priority < 1 {
				priority = 1
			}

			risk := ie.assessRisk(step.Command, step.Risk)

			actions = append(actions, IntelligentAction{
				Type:        "fix",
				Priority:    priority,
				Description: step.Action,
				Command:     step.Command,
				Risk:        risk,
				Expected:    "Resolution of identified issue",
				Automated:   risk.Level == "low" && step.Category == "investigate",
			})
		}
	} else {
		// Generic diagnostic actions when root cause is unknown
		actions = append(actions, IntelligentAction{
			Type:        "diagnostic",
			Priority:    7,
			Description: "Run comprehensive cluster health check",
			Command:     ie.generateHealthCheckCommand(session.Context),
			Risk:        RiskLevel{Level: "low", Factors: []string{"read-only operations"}, Reversible: true},
			Expected:    "Overview of cluster health and potential issues",
			Automated:   true,
		})
	}

	return actions
}

// generateCreationActions creates intelligent resource generation actions
func (ie *IntelligenceEngine) generateCreationActions(session *AnalysisSession) []IntelligentAction {
	var actions []IntelligentAction

	actions = append(actions, IntelligentAction{
		Type:        "generate",
		Priority:    8,
		Description: "Generate production-ready manifest with best practices",
		Command:     "# AI will generate manifest based on requirements",
		Risk:        RiskLevel{Level: "medium", Factors: []string{"creates new resources"}, Mitigations: []string{"dry-run first"}, Reversible: true},
		Expected:    "YAML manifest with proper labels, resources, and security settings",
		Automated:   false,
		PreChecks:   []string{"validate namespace exists", "check resource quotas"},
	})

	return actions
}

// generateEditActions creates intelligent editing actions
func (ie *IntelligenceEngine) generateEditActions(session *AnalysisSession) []IntelligentAction {
	var actions []IntelligentAction

	actions = append(actions, IntelligentAction{
		Type:        "edit",
		Priority:    8,
		Description: "Generate minimal, targeted diff for requested changes",
		Command:     "# AI will generate unified diff",
		Risk:        ie.assessEditRisk(session.Context),
		Expected:    "Unified diff showing only necessary changes",
		Automated:   false,
		PreChecks:   []string{"backup current configuration", "validate syntax"},
	})

	return actions
}

// generateExplanationActions creates intelligent explanation actions
func (ie *IntelligenceEngine) generateExplanationActions(session *AnalysisSession) []IntelligentAction {
	var actions []IntelligentAction

	actions = append(actions, IntelligentAction{
		Type:        "explain",
		Priority:    6,
		Description: "Provide contextual explanation with current cluster state",
		Command:     "# AI will generate explanation with examples",
		Risk:        RiskLevel{Level: "low", Factors: []string{"informational only"}, Reversible: true},
		Expected:    "Clear explanation with practical examples from current cluster",
		Automated:   true,
	})

	return actions
}

// generateCommandActions creates intelligent command synthesis actions
func (ie *IntelligenceEngine) generateCommandActions(session *AnalysisSession) []IntelligentAction {
	var actions []IntelligentAction

	actions = append(actions, IntelligentAction{
		Type:        "command",
		Priority:    7,
		Description: "Synthesize precise kubectl/helm command",
		Command:     "# AI will generate appropriate command",
		Risk:        ie.assessCommandRisk(session.Context),
		Expected:    "Single, executable command addressing the request",
		Automated:   true,
		PreChecks:   []string{"validate command syntax", "check for mutations"},
	})

	return actions
}

// generateOptimizationActions creates intelligent optimization actions
func (ie *IntelligenceEngine) generateOptimizationActions(session *AnalysisSession) []IntelligentAction {
	var actions []IntelligentAction

	for _, rec := range session.Optimization {
		if rec.Severity == "critical" || rec.Severity == "warning" {
			priority := 6
			if rec.Severity == "critical" {
				priority = 8
			}

			risk := RiskLevel{
				Level:       rec.RiskLevel,
				Factors:     []string{fmt.Sprintf("%s optimization", rec.Type)},
				Mitigations: []string{"preview changes first", "apply during maintenance window"},
				Reversible:  rec.Type == "resource", // Resource changes are usually reversible
			}

			actions = append(actions, IntelligentAction{
				Type:        "optimize",
				Priority:    priority,
				Description: rec.Title,
				Command:     rec.PreviewCmd,
				Risk:        risk,
				Expected:    rec.Impact,
				Automated:   rec.Automated,
			})
		}
	}

	return actions
}

// Helper methods for intelligent decision making
func (ie *IntelligenceEngine) generateTargetedDiagnosticCommand(rootCause *RootCauseAnalysis) string {
	switch rootCause.Category {
	case "image-issues":
		return "kubectl describe pods -l app={app} | grep -A 10 -B 5 'Events:'"
	case "resource-limits":
		return "kubectl top pods --sort-by=memory && kubectl describe nodes | grep -A 10 'Allocated resources'"
	case "scheduling":
		return "kubectl get events --sort-by=.lastTimestamp | tail -20"
	case "networking":
		return "kubectl get endpoints && kubectl describe services"
	default:
		return "kubectl get events --sort-by=.lastTimestamp | tail -10"
	}
}

func (ie *IntelligenceEngine) generateHealthCheckCommand(context *KubeContextSummary) string {
	if context != nil && context.Namespace != "" {
		return fmt.Sprintf("kubectl get pods,services,deployments -n %s && kubectl get events -n %s --sort-by=.lastTimestamp | tail -10", context.Namespace, context.Namespace)
	}
	return "kubectl get nodes && kubectl get pods --all-namespaces | grep -v Running"
}

func (ie *IntelligenceEngine) assessRisk(command, riskHint string) RiskLevel {
	risk := RiskLevel{
		Reversible: true,
	}

	// Check command for danger patterns
	cmdLower := strings.ToLower(command)

	if strings.Contains(cmdLower, "delete") || strings.Contains(cmdLower, "remove") {
		risk.Level = "high"
		risk.Factors = []string{"destructive operation", "data loss risk"}
		risk.Reversible = false
	} else if strings.Contains(cmdLower, "patch") || strings.Contains(cmdLower, "apply") {
		risk.Level = "medium"
		risk.Factors = []string{"configuration change", "service impact"}
		risk.Mitigations = []string{"dry-run first", "backup current config"}
	} else if strings.Contains(cmdLower, "get") || strings.Contains(cmdLower, "describe") {
		risk.Level = "low"
		risk.Factors = []string{"read-only operation"}
	} else {
		risk.Level = riskHint
		if risk.Level == "" {
			risk.Level = "medium"
		}
	}

	return risk
}

func (ie *IntelligenceEngine) assessEditRisk(context *KubeContextSummary) RiskLevel {
	risk := RiskLevel{
		Level:       "medium",
		Factors:     []string{"configuration modification"},
		Mitigations: []string{"diff preview", "dry-run validation", "backup"},
		Reversible:  true,
	}

	// Increase risk for production-like namespaces
	if context != nil {
		prodPatterns := []string{"prod", "production", "live", "staging"}
		for _, pattern := range prodPatterns {
			if strings.Contains(strings.ToLower(context.Namespace), pattern) {
				risk.Level = "high"
				risk.Factors = append(risk.Factors, "production environment")
				break
			}
		}
	}

	return risk
}

func (ie *IntelligenceEngine) assessCommandRisk(context *KubeContextSummary) RiskLevel {
	risk := RiskLevel{
		Level:       "low",
		Factors:     []string{"command synthesis"},
		Mitigations: []string{"prefer read-only operations", "dry-run for mutations"},
		Reversible:  true,
	}

	// Adjust risk based on context
	if context != nil {
		prodPatterns := []string{"prod", "production", "live"}
		for _, pattern := range prodPatterns {
			if strings.Contains(strings.ToLower(context.Namespace), pattern) {
				risk.Level = "medium"
				risk.Factors = append(risk.Factors, "production environment")
				break
			}
		}
	}

	return risk
}

func (ie *IntelligenceEngine) calculateOverallConfidence(session *AnalysisSession) float64 {
	confidence := session.Intent.Confidence

	// Boost confidence if we have root cause analysis
	if session.RootCause != nil && session.RootCause.Confidence > 0 {
		confidence = (confidence + session.RootCause.Confidence) / 2
	}

	// Boost confidence if we have optimization recommendations
	if len(session.Optimization) > 0 {
		confidence = min(confidence+0.1, 1.0)
	}

	return confidence
}

// GetIntelligentInsights generates smart insights from analysis
func (ie *IntelligenceEngine) GetIntelligentInsights(session *AnalysisSession) []IntelligentInsight {
	var insights []IntelligentInsight

	// Pattern-based insights
	if session.RootCause != nil && session.RootCause.Confidence > 0.7 {
		insights = append(insights, IntelligentInsight{
			Type:        "pattern",
			Title:       fmt.Sprintf("Detected: %s", session.RootCause.RootCause),
			Description: fmt.Sprintf("High confidence (%.1f%%) pattern match for common Kubernetes issue", session.RootCause.Confidence*100),
			Confidence:  session.RootCause.Confidence,
			Impact:      session.RootCause.Timeline,
			Evidence:    session.RootCause.Indicators,
			NextSteps:   session.RootCause.PreventionTips,
		})
	}

	// Optimization insights
	criticalOptimizations := 0
	for _, opt := range session.Optimization {
		if opt.Severity == "critical" {
			criticalOptimizations++
		}
	}

	if criticalOptimizations > 0 {
		insights = append(insights, IntelligentInsight{
			Type:        "optimization",
			Title:       fmt.Sprintf("%d Critical Optimization Opportunities", criticalOptimizations),
			Description: "Found opportunities to improve performance, reduce costs, or enhance reliability",
			Confidence:  0.9,
			Impact:      "Significant improvement potential",
			Evidence:    []string{"resource utilization analysis", "configuration review"},
			NextSteps:   []string{"Review recommendations", "Apply in non-production first", "Monitor impact"},
		})
	}

	return insights
}

// GetPredictiveActions returns predictive action suggestions
func (ie *IntelligenceEngine) GetPredictiveActions(input string, context *KubeContextSummary) []PredictedAction {
	if ie.predictive == nil {
		return nil
	}
	return ie.predictive.PredictActions(input, context)
}

// LearnFromUserAction learns from user actions to improve predictions
func (ie *IntelligenceEngine) LearnFromUserAction(input string, context *KubeContextSummary, action string, success bool) {
	if ie.predictive != nil {
		ie.predictive.LearnFromInteraction(input, context, action, success)
	}
}

// GetPredictiveInsights returns insights about prediction accuracy and patterns
func (ie *IntelligenceEngine) GetPredictiveInsights() map[string]interface{} {
	if ie.predictive == nil {
		return nil
	}

	insights := make(map[string]interface{})

	// Cache statistics
	cacheStats := map[string]interface{}{
		"hit_rate":       ie.predictive.predictionCache.hitRate,
		"total_requests": ie.predictive.predictionCache.totalRequests,
		"cache_hits":     ie.predictive.predictionCache.cacheHits,
		"cache_size":     len(ie.predictive.predictionCache.entries),
	}
	insights["cache_stats"] = cacheStats

	// Pattern statistics
	patternStats := map[string]interface{}{
		"total_patterns":    len(ie.predictive.patterns),
		"context_patterns":  len(ie.predictive.contextPatterns),
		"min_confidence":    ie.predictive.minConfidence,
		"enabled":          ie.predictive.enabled,
	}
	insights["pattern_stats"] = patternStats

	// User behavior insights
	if ie.predictive.userBehavior != nil {
		behaviorStats := map[string]interface{}{
			"total_commands":    ie.predictive.userBehavior.CommandFrequency["total"],
			"unique_commands":   len(ie.predictive.userBehavior.CommandFrequency) - 1, // Excluding "total"
			"sequence_patterns": len(ie.predictive.userBehavior.SequencePatterns),
			"last_updated":      ie.predictive.userBehavior.LastUpdated,
		}
		insights["behavior_stats"] = behaviorStats
	}

	return insights
}

// GetCacheStats returns smart cache system statistics
func (ie *IntelligenceEngine) GetCacheStats() *CacheStats {
	if ie.cache == nil {
		return nil
	}
	stats := ie.cache.GetStats()
	return &stats
}

// GetCacheHitRatio returns overall cache hit ratio
func (ie *IntelligenceEngine) GetCacheHitRatio() float64 {
	if ie.cache == nil {
		return 0.0
	}
	return ie.cache.GetHitRatio()
}

// InvalidateCache invalidates cache entries matching a pattern
func (ie *IntelligenceEngine) InvalidateCache(pattern string) {
	if ie.cache != nil {
		ie.cache.Invalidate(pattern)
	}
}

// AnalyzeIntelligentyWithCache performs analysis with smart caching
func (ie *IntelligenceEngine) AnalyzeIntelligentyWithCache(input string, context *KubeContextSummary) (*AnalysisSession, error) {
	// Generate cache key based on input and context
	cacheKey := fmt.Sprintf("analysis_%s_%s_%s",
		input,
		context.Context,
		context.Namespace)

	// Try cache first
	if ie.cache != nil {
		if cached, found := ie.cache.Get(cacheKey, CacheTypeAnalysis); found {
			if session, ok := cached.(*AnalysisSession); ok {
				return session, nil
			}
		}
	}

	// Cache miss - perform analysis
	session, err := ie.AnalyzeIntelligently(input, context)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if ie.cache != nil {
		ie.cache.Set(cacheKey, session, CacheTypeAnalysis, 10*time.Minute)
	}

	return session, nil
}

// GetOptimizationWithCache retrieves optimization recommendations with caching
func (ie *IntelligenceEngine) GetOptimizationWithCache(context *KubeContextSummary) []Recommendation {
	if context == nil {
		return nil
	}

	cacheKey := fmt.Sprintf("optimization_%s_%s", context.Context, context.Namespace)

	// Try cache first
	if ie.cache != nil {
		if cached, found := ie.cache.Get(cacheKey, CacheTypeOptimization); found {
			if recommendations, ok := cached.([]Recommendation); ok {
				return recommendations
			}
		}
	}

	// Cache miss - generate recommendations
	recommendations := ie.optimizer.GetOptimizationRecommendations(context)

	// Cache the result
	if ie.cache != nil {
		ie.cache.Set(cacheKey, recommendations, CacheTypeOptimization, 15*time.Minute)
	}

	return recommendations
}

// WarmupCache preloads frequently used data
func (ie *IntelligenceEngine) WarmupCache(context *KubeContextSummary) {
	if ie.cache == nil {
		return
	}

	// Preload common analysis patterns
	commonInputs := []string{
		"show pod status",
		"list failing pods",
		"check resource usage",
		"diagnose issues",
		"optimize performance",
	}

	for _, input := range commonInputs {
		go func(inp string) {
			_, _ = ie.AnalyzeIntelligentyWithCache(inp, context)
		}(input)
	}
}

// Removed global instance - now created via dependency injection

// Enhanced intelligence analysis with predictive capabilities
func (ie *IntelligenceEngine) AnalyzeIntelligentlyWithPrediction(input string, context *KubeContextSummary) (*AnalysisSession, error) {
	// First run standard analysis
	session, err := ie.AnalyzeIntelligently(input, context)
	if err != nil {
		return nil, err
	}

	// Add predictive intelligence if available
	if PredictiveIntelligence != nil {
		predictions := PredictiveIntelligence.PredictNextActions(context, input)
		
		// Convert predictions to intelligent actions
		for _, pred := range predictions {
			action := IntelligentAction{
				Type:        "predictive",
				Priority:    pred.Priority,
				Description: fmt.Sprintf("Predicted: %s (%.0f%% confidence)", pred.Action, pred.Confidence*100),
				Command:     pred.Action,
				Risk: RiskLevel{
					Level:       pred.RiskLevel,
					Factors:     []string{"predictive analysis"},
					Mitigations: []string{"verify before execution"},
					Reversible:  true,
				},
				Expected:  fmt.Sprintf("Estimated completion: %v", pred.EstimatedTime),
				Automated: pred.Confidence > 0.8,
				PreChecks: pred.Prerequisites,
			}
			session.Actions = append(session.Actions, action)
		}

		// Update confidence based on predictions
		if len(predictions) > 0 {
			avgPredictionConfidence := 0.0
			for _, pred := range predictions {
				avgPredictionConfidence += pred.Confidence
			}
			avgPredictionConfidence /= float64(len(predictions))
			
			// Blend original confidence with prediction confidence
			session.Confidence = (session.Confidence + avgPredictionConfidence) / 2.0
		}
	}

	return session, nil
}
