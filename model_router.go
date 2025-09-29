// model_router.go - Intelligent model routing for optimal LLM selection
package main

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"sync"
	"time"
)

// ModelRouter manages intelligent model selection based on query complexity
type ModelRouter struct {
	fastModel       string  // llama3.1:8b for simple queries
	deepModel       string  // llama3.1:70b for complex analysis
	codeModel       string  // codellama for code generation
	diagnosticModel string  // specialized for diagnostics
	smartCache      *SmartCache
	modelSelector   *ModelSelector
	performanceTracker *ModelPerformanceTracker
	mu              sync.RWMutex
}

// ModelSelector analyzes queries to select optimal models
type ModelSelector struct {
	queryClassifier    *QueryClassifier
	complexityAnalyzer *ComplexityAnalyzer
	contextAnalyzer    *ContextAnalyzer
	performanceTracker *ModelPerformanceTracker
	selectionHistory   []ModelSelection
	mu                 sync.RWMutex
}

// QueryClassifier classifies queries into categories
type QueryClassifier struct {
	patterns        map[QueryType]*QueryPattern
	keywords        map[string]QueryType
	regexPatterns   map[QueryType]*regexp.Regexp
	confidenceThreshold float64
}

// ComplexityAnalyzer analyzes query complexity
type ComplexityAnalyzer struct {
	complexityFactors map[string]float64
	weightings        map[string]float64
	thresholds        map[string]float64
}

// ContextAnalyzer analyzes context for model selection
type ContextAnalyzer struct {
	contextPatterns   map[string]*ContextPattern
	riskAssessment    *RiskAssessment
	performanceHints  map[string]string
}

// ModelPerformanceTracker tracks model performance metrics
type ModelPerformanceTracker struct {
	modelMetrics     map[string]*ModelMetrics
	responseHistory  []ModelResponse
	optimizationRules []OptimizationRule
	mu               sync.RWMutex
}

// QueryType represents different types of queries
type QueryType int

const (
	QueryTypeSimple QueryType = iota
	QueryTypeComplex
	QueryTypeDiagnostic
	QueryTypeCode
	QueryTypeExplanation
	QueryTypeOptimization
	QueryTypeEmergency
)

// QueryPattern represents patterns for query classification
type QueryPattern struct {
	Type        QueryType `json:"type"`
	Keywords    []string  `json:"keywords"`
	Patterns    []string  `json:"patterns"`
	Complexity  float64   `json:"complexity"`
	Confidence  float64   `json:"confidence"`
	ModelHint   string    `json:"model_hint"`
}

// ContextPattern represents context-based patterns
type ContextPattern struct {
	Pattern     string    `json:"pattern"`
	ModelHint   string    `json:"model_hint"`
	Priority    int       `json:"priority"`
	Conditions  []string  `json:"conditions"`
	Performance float64   `json:"performance"`
}

// RiskAssessment represents risk-based model selection
type RiskAssessment struct {
	riskLevels    map[string]float64
	safetyModels  map[string]string
	riskThreshold float64
}

// ModelSelection represents a model selection decision
type ModelSelection struct {
	Query         string        `json:"query"`
	Context       string        `json:"context"`
	SelectedModel string        `json:"selected_model"`
	Confidence    float64       `json:"confidence"`
	Reasoning     []string      `json:"reasoning"`
	Timestamp     time.Time     `json:"timestamp"`
	Performance   *ModelMetrics `json:"performance"`
}

// ModelMetrics represents performance metrics for a model
type ModelMetrics struct {
	ModelName       string        `json:"model_name"`
	AvgResponseTime time.Duration `json:"avg_response_time"`
	SuccessRate     float64       `json:"success_rate"`
	QualityScore    float64       `json:"quality_score"`
	TokenEfficiency float64       `json:"token_efficiency"`
	UsageCount      int           `json:"usage_count"`
	LastUsed        time.Time     `json:"last_used"`
	CostPerQuery    float64       `json:"cost_per_query"`
}

// ModelResponse represents a model response for tracking
type ModelResponse struct {
	ModelName    string        `json:"model_name"`
	Query        string        `json:"query"`
	ResponseTime time.Duration `json:"response_time"`
	Success      bool          `json:"success"`
	Quality      float64       `json:"quality"`
	TokensUsed   int           `json:"tokens_used"`
	Timestamp    time.Time     `json:"timestamp"`
}

// OptimizationRule represents rules for model optimization
type OptimizationRule struct {
	Name        string                    `json:"name"`
	Condition   func(*ModelSelection) bool `json:"-"`
	Action      func(*ModelSelection) string `json:"-"`
	Priority    int                       `json:"priority"`
	Enabled     bool                      `json:"enabled"`
	Description string                    `json:"description"`
}

// NewModelRouter creates a new model router
func NewModelRouter(smartCache *SmartCache) *ModelRouter {
	return &ModelRouter{
		fastModel:       "llama3.1:8b",
		deepModel:       "llama3.1:70b",
		codeModel:       "codellama:13b",
		diagnosticModel: "llama3.1:13b",
		smartCache:      smartCache,
		modelSelector: &ModelSelector{
			queryClassifier: &QueryClassifier{
				patterns:            make(map[QueryType]*QueryPattern),
				keywords:            make(map[string]QueryType),
				regexPatterns:       make(map[QueryType]*regexp.Regexp),
				confidenceThreshold: 0.7,
			},
			complexityAnalyzer: &ComplexityAnalyzer{
				complexityFactors: map[string]float64{
					"word_count":      0.1,
					"technical_terms": 0.3,
					"question_marks":  0.2,
					"yaml_content":    0.4,
					"multi_step":      0.5,
				},
				weightings: map[string]float64{
					"length":     0.2,
					"complexity": 0.4,
					"context":    0.3,
					"urgency":    0.1,
				},
				thresholds: map[string]float64{
					"simple":  0.3,
					"medium":  0.6,
					"complex": 0.8,
				},
			},
			contextAnalyzer: &ContextAnalyzer{
				contextPatterns: make(map[string]*ContextPattern),
				riskAssessment: &RiskAssessment{
					riskLevels:    make(map[string]float64),
					safetyModels:  make(map[string]string),
					riskThreshold: 0.7,
				},
				performanceHints: make(map[string]string),
			},
			selectionHistory: make([]ModelSelection, 0),
		},
		performanceTracker: &ModelPerformanceTracker{
			modelMetrics:      make(map[string]*ModelMetrics),
			responseHistory:   make([]ModelResponse, 0),
			optimizationRules: make([]OptimizationRule, 0),
		},
	}
}

// SelectModel selects the optimal model for a given query and context
func (mr *ModelRouter) SelectModel(query string, context *KubeContextSummary) (string, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	// Check cache first
	cacheKey := fmt.Sprintf("model_selection:%s:%s", query, context.Hash())
	if cached := mr.smartCache.GetL1(cacheKey); cached != nil {
		return cached.(string), nil
	}

	// Analyze query complexity and type
	queryType := mr.modelSelector.queryClassifier.ClassifyQuery(query)
	complexity := mr.modelSelector.complexityAnalyzer.AnalyzeComplexity(query, context)
	contextHints := mr.modelSelector.contextAnalyzer.AnalyzeContext(context)

	// Create selection context
	selection := &ModelSelection{
		Query:     query,
		Context:   context.Hash(),
		Timestamp: time.Now(),
		Reasoning: make([]string, 0),
	}

	// Select model based on analysis
	selectedModel := mr.selectOptimalModel(queryType, complexity, contextHints, selection)
	selection.SelectedModel = selectedModel
	selection.Confidence = mr.calculateSelectionConfidence(selection)

	// Record selection
	mr.modelSelector.RecordSelection(*selection)

	// Cache the selection
	mr.smartCache.SetL1(cacheKey, selectedModel, 5*time.Minute)

	return selectedModel, nil
}

// selectOptimalModel selects the optimal model based on analysis
func (mr *ModelRouter) selectOptimalModel(queryType QueryType, complexity float64, contextHints map[string]string, selection *ModelSelection) string {
	// Check performance history for model recommendations
	bestModel := mr.performanceTracker.GetBestModelForQuery(selection.Query)
	if bestModel != "" {
		selection.Reasoning = append(selection.Reasoning, fmt.Sprintf("performance_history_suggests_%s", bestModel))
		return bestModel
	}

	// Apply query type rules
	switch queryType {
	case QueryTypeSimple:
		if complexity < 0.3 {
			selection.Reasoning = append(selection.Reasoning, "simple_query_low_complexity")
			return mr.fastModel
		}
		selection.Reasoning = append(selection.Reasoning, "simple_query_medium_complexity")
		return mr.fastModel

	case QueryTypeCode:
		selection.Reasoning = append(selection.Reasoning, "code_generation_query")
		return mr.codeModel

	case QueryTypeDiagnostic:
		if complexity > 0.7 {
			selection.Reasoning = append(selection.Reasoning, "complex_diagnostic_query")
			return mr.deepModel
		}
		selection.Reasoning = append(selection.Reasoning, "standard_diagnostic_query")
		return mr.diagnosticModel

	case QueryTypeEmergency:
		selection.Reasoning = append(selection.Reasoning, "emergency_query_fast_response")
		return mr.fastModel

	case QueryTypeComplex:
		selection.Reasoning = append(selection.Reasoning, "complex_query_deep_analysis")
		return mr.deepModel

	case QueryTypeOptimization:
		selection.Reasoning = append(selection.Reasoning, "optimization_query_specialized")
		return mr.diagnosticModel

	default:
		// Apply complexity-based selection
		if complexity < 0.3 {
			selection.Reasoning = append(selection.Reasoning, "low_complexity_fast_model")
			return mr.fastModel
		} else if complexity > 0.7 {
			selection.Reasoning = append(selection.Reasoning, "high_complexity_deep_model")
			return mr.deepModel
		} else {
			selection.Reasoning = append(selection.Reasoning, "medium_complexity_diagnostic_model")
			return mr.diagnosticModel
		}
	}
}

// ClassifyQuery classifies a query into a type
func (qc *QueryClassifier) ClassifyQuery(query string) QueryType {
	queryLower := strings.ToLower(query)

	// Check for code-related queries
	codeKeywords := []string{"yaml", "json", "manifest", "template", "generate", "create"}
	for _, keyword := range codeKeywords {
		if strings.Contains(queryLower, keyword) {
			return QueryTypeCode
		}
	}

	// Check for diagnostic queries
	diagnosticKeywords := []string{"debug", "troubleshoot", "error", "failed", "not working", "issue", "problem"}
	for _, keyword := range diagnosticKeywords {
		if strings.Contains(queryLower, keyword) {
			return QueryTypeDiagnostic
		}
	}

	// Check for emergency queries
	emergencyKeywords := []string{"urgent", "critical", "emergency", "down", "outage", "crash"}
	for _, keyword := range emergencyKeywords {
		if strings.Contains(queryLower, keyword) {
			return QueryTypeEmergency
		}
	}

	// Check for optimization queries
	optimizationKeywords := []string{"optimize", "improve", "performance", "scale", "resource", "efficiency"}
	for _, keyword := range optimizationKeywords {
		if strings.Contains(queryLower, keyword) {
			return QueryTypeOptimization
		}
	}

	// Check for explanation queries
	explanationKeywords := []string{"explain", "what is", "how does", "why", "help", "understand"}
	for _, keyword := range explanationKeywords {
		if strings.Contains(queryLower, keyword) {
			return QueryTypeExplanation
		}
	}

	// Analyze complexity to determine if it's simple or complex
	wordCount := len(strings.Fields(query))
	if wordCount < 5 {
		return QueryTypeSimple
	} else if wordCount > 20 {
		return QueryTypeComplex
	}

	return QueryTypeSimple
}

// AnalyzeComplexity analyzes the complexity of a query
func (ca *ComplexityAnalyzer) AnalyzeComplexity(query string, context *KubeContextSummary) float64 {
	complexity := 0.0

	// Word count factor
	wordCount := len(strings.Fields(query))
	complexity += float64(wordCount) * ca.complexityFactors["word_count"]

	// Technical terms factor
	technicalTerms := ca.countTechnicalTerms(query)
	complexity += float64(technicalTerms) * ca.complexityFactors["technical_terms"]

	// Question marks (multiple questions increase complexity)
	questionMarks := strings.Count(query, "?")
	complexity += float64(questionMarks) * ca.complexityFactors["question_marks"]

	// YAML/JSON content increases complexity
	if strings.Contains(query, "apiVersion") || strings.Contains(query, "{") || strings.Contains(query, "---") {
		complexity += ca.complexityFactors["yaml_content"]
	}

	// Multi-step queries
	if strings.Contains(query, "and then") || strings.Contains(query, "after that") || strings.Contains(query, "next") {
		complexity += ca.complexityFactors["multi_step"]
	}

	// Context complexity
	if context != nil {
		if len(context.PodProblemCounts) > 0 {
			complexity += 0.2 // Problems increase complexity
		}
		if strings.Contains(strings.ToLower(context.Namespace), "prod") {
			complexity += 0.1 // Production context adds complexity
		}
	}

	// Normalize to 0-1 range
	complexity = math.Min(1.0, complexity)
	return complexity
}

// countTechnicalTerms counts technical Kubernetes terms in the query
func (ca *ComplexityAnalyzer) countTechnicalTerms(query string) int {
	technicalTerms := []string{
		"kubectl", "helm", "pod", "deployment", "service", "ingress", "configmap",
		"secret", "namespace", "node", "cluster", "container", "image", "volume",
		"pvc", "pv", "rbac", "serviceaccount", "networkpolicy", "hpa", "vpa",
		"daemonset", "statefulset", "job", "cronjob", "operator", "crd",
	}

	queryLower := strings.ToLower(query)
	count := 0
	for _, term := range technicalTerms {
		if strings.Contains(queryLower, term) {
			count++
		}
	}
	return count
}

// AnalyzeContext analyzes context for model selection hints
func (ca *ContextAnalyzer) AnalyzeContext(context *KubeContextSummary) map[string]string {
	hints := make(map[string]string)

	if context == nil {
		return hints
	}

	// Production context suggests careful model selection
	if strings.Contains(strings.ToLower(context.Namespace), "prod") {
		hints["environment"] = "production"
		hints["model_preference"] = "conservative"
	}

	// Problem context suggests diagnostic model
	if len(context.PodProblemCounts) > 0 {
		hints["situation"] = "problems_detected"
		hints["model_preference"] = "diagnostic"
	}

	// System namespace suggests technical queries
	if context.Namespace == "kube-system" {
		hints["namespace_type"] = "system"
		hints["model_preference"] = "technical"
	}

	return hints
}

// RecordSelection records a model selection for learning
func (ms *ModelSelector) RecordSelection(selection ModelSelection) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.selectionHistory = append(ms.selectionHistory, selection)

	// Maintain history size
	if len(ms.selectionHistory) > 1000 {
		ms.selectionHistory = ms.selectionHistory[len(ms.selectionHistory)-1000:]
	}
}

// RecordResponse records a model response for performance tracking
func (mpt *ModelPerformanceTracker) RecordResponse(response ModelResponse) {
	mpt.mu.Lock()
	defer mpt.mu.Unlock()

	// Add to response history
	mpt.responseHistory = append(mpt.responseHistory, response)

	// Update model metrics
	if metrics, exists := mpt.modelMetrics[response.ModelName]; exists {
		mpt.updateModelMetrics(metrics, response)
	} else {
		mpt.modelMetrics[response.ModelName] = mpt.createModelMetrics(response)
	}

	// Maintain history size
	if len(mpt.responseHistory) > 1000 {
		mpt.responseHistory = mpt.responseHistory[len(mpt.responseHistory)-1000:]
	}
}

// updateModelMetrics updates existing model metrics
func (mpt *ModelPerformanceTracker) updateModelMetrics(metrics *ModelMetrics, response ModelResponse) {
	// Update response time (exponential moving average)
	metrics.AvgResponseTime = time.Duration(float64(metrics.AvgResponseTime)*0.9 + float64(response.ResponseTime)*0.1)

	// Update success rate
	if response.Success {
		metrics.SuccessRate = metrics.SuccessRate*0.9 + 0.1
	} else {
		metrics.SuccessRate = metrics.SuccessRate * 0.9
	}

	// Update quality score
	metrics.QualityScore = metrics.QualityScore*0.9 + response.Quality*0.1

	// Update token efficiency
	if response.TokensUsed > 0 {
		efficiency := response.Quality / float64(response.TokensUsed) * 1000 // Quality per 1000 tokens
		metrics.TokenEfficiency = metrics.TokenEfficiency*0.9 + efficiency*0.1
	}

	metrics.UsageCount++
	metrics.LastUsed = response.Timestamp
}

// createModelMetrics creates new model metrics
func (mpt *ModelPerformanceTracker) createModelMetrics(response ModelResponse) *ModelMetrics {
	successRate := 0.0
	if response.Success {
		successRate = 1.0
	}

	tokenEfficiency := 0.0
	if response.TokensUsed > 0 {
		tokenEfficiency = response.Quality / float64(response.TokensUsed) * 1000
	}

	return &ModelMetrics{
		ModelName:       response.ModelName,
		AvgResponseTime: response.ResponseTime,
		SuccessRate:     successRate,
		QualityScore:    response.Quality,
		TokenEfficiency: tokenEfficiency,
		UsageCount:      1,
		LastUsed:        response.Timestamp,
		CostPerQuery:    0.0, // Would be calculated based on actual costs
	}
}

// GetBestModelForQuery returns the best performing model for similar queries
func (mpt *ModelPerformanceTracker) GetBestModelForQuery(query string) string {
	mpt.mu.RLock()
	defer mpt.mu.RUnlock()

	// Find similar queries in history
	var candidates []string
	queryLower := strings.ToLower(query)

	for _, response := range mpt.responseHistory {
		if mpt.isSimilarQuery(queryLower, strings.ToLower(response.Query)) {
			candidates = append(candidates, response.ModelName)
		}
	}

	if len(candidates) == 0 {
		return ""
	}

	// Find the model with best performance for these queries
	modelScores := make(map[string]float64)
	for _, modelName := range candidates {
		if metrics, exists := mpt.modelMetrics[modelName]; exists {
			// Calculate composite score
			score := metrics.SuccessRate*0.4 + metrics.QualityScore*0.4 + metrics.TokenEfficiency*0.2
			modelScores[modelName] = score
		}
	}

	// Return the best scoring model
	bestModel := ""
	bestScore := 0.0
	for model, score := range modelScores {
		if score > bestScore {
			bestScore = score
			bestModel = model
		}
	}

	return bestModel
}

// isSimilarQuery checks if two queries are similar
func (mpt *ModelPerformanceTracker) isSimilarQuery(query1, query2 string) bool {
	// Simple similarity check based on common words
	words1 := strings.Fields(query1)
	words2 := strings.Fields(query2)

	if len(words1) == 0 || len(words2) == 0 {
		return false
	}

	commonWords := 0
	for _, word1 := range words1 {
		for _, word2 := range words2 {
			if word1 == word2 && len(word1) > 3 { // Only count significant words
				commonWords++
				break
			}
		}
	}

	similarity := float64(commonWords) / float64(len(words1))
	return similarity > 0.3 // 30% similarity threshold
}

// calculateSelectionConfidence calculates confidence in model selection
func (mr *ModelRouter) calculateSelectionConfidence(selection *ModelSelection) float64 {
	confidence := 0.5 // Base confidence

	// Increase confidence based on reasoning strength
	confidence += float64(len(selection.Reasoning)) * 0.1

	// Increase confidence if we have performance history
	if metrics, exists := mr.performanceTracker.modelMetrics[selection.SelectedModel]; exists {
		confidence += metrics.SuccessRate * 0.3
	}

	// Cap at 1.0
	return math.Min(1.0, confidence)
}

// GetModelStats returns model routing statistics
func (mr *ModelRouter) GetModelStats() map[string]interface{} {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	stats := make(map[string]interface{})
	
	// Model usage statistics
	modelUsage := make(map[string]int)
	for _, metrics := range mr.performanceTracker.modelMetrics {
		modelUsage[metrics.ModelName] = metrics.UsageCount
	}
	stats["model_usage"] = modelUsage

	// Performance statistics
	performanceStats := make(map[string]interface{})
	for modelName, metrics := range mr.performanceTracker.modelMetrics {
		performanceStats[modelName] = map[string]interface{}{
			"avg_response_time": metrics.AvgResponseTime.String(),
			"success_rate":      metrics.SuccessRate,
			"quality_score":     metrics.QualityScore,
			"token_efficiency":  metrics.TokenEfficiency,
		}
	}
	stats["performance"] = performanceStats

	// Selection statistics
	stats["total_selections"] = len(mr.modelSelector.selectionHistory)
	stats["total_responses"] = len(mr.performanceTracker.responseHistory)

	return stats
}

// Global model router instance
var ModelRouter *ModelRouter

// InitializeModelRouter initializes the global model router
func InitializeModelRouter(smartCache *SmartCache) {
	ModelRouter = NewModelRouter(smartCache)
}

