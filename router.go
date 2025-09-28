// router.go - Intelligent intent routing for precise mode selection
package main

import (
	"math"
	"regexp"
	"strings"
	"time"
)

// AgentMode represents different operational modes
type AgentMode string

const (
	ModeDiagnose AgentMode = "diagnose" // ReAct-lite diagnostic loops
	ModeGenerate AgentMode = "generate" // Schema-constrained YAML/values generation
	ModeEdit     AgentMode = "edit"     // Unified diff editing
	ModeExplain  AgentMode = "explain"  // Short explanations with context
	ModeCommand  AgentMode = "command"  // Direct command synthesis
)

// IntentRouter provides intelligent mode selection with confidence scoring
type IntentRouter struct {
	Mode         AgentMode     `json:"mode"`
	Confidence   float64       `json:"confidence"`
	Alternatives []AgentMode   `json:"alternatives"`
	Reasoning    string        `json:"reasoning"`
	Policy       *PromptPolicy `json:"policy"`
}

// PromptPolicy defines mode-specific prompting strategies
type PromptPolicy struct {
	Mode            AgentMode `json:"mode"`
	SystemPrompt    string    `json:"system_prompt"`
	Temperature     float64   `json:"temperature"`
	MaxTokens       int       `json:"max_tokens"`
	ContextStrategy string    `json:"context_strategy"`
	SafetyLevel     string    `json:"safety_level"`
}

// IntentPattern defines patterns for intent classification
type IntentPattern struct {
	Mode       AgentMode
	Patterns   []string
	Weight     float64
	Examples   []string
	Confidence float64
}

// Global intent patterns for classification
var intentPatterns = []IntentPattern{
	{
		Mode:   ModeDiagnose,
		Weight: 1.0,
		Patterns: []string{
			`(?i)(not working|failing|broken|error|issue|problem|debug|troubleshoot|investigate)`,
			`(?i)(why.*not|what.*wrong|check.*status|diagnose)`,
			`(?i)(crashloop|imagepull|oomkilled|pending|failed)`,
			`(?i)(logs|events|describe|status)`,
		},
		Examples: []string{
			"Why is my pod not running?",
			"My service isn't working",
			"Debug this CrashLoopBackOff",
			"Investigate the failing deployment",
		},
		Confidence: 0.9,
	},
	{
		Mode:   ModeGenerate,
		Weight: 1.0,
		Patterns: []string{
			`(?i)(create|generate|new|scaffold|build|make)`,
			`(?i)(deployment|service|ingress|configmap|secret).*(?i)(for|with)`,
			`(?i)(helm.*chart|values.*file)`,
			`(?i)(template|manifest|yaml)`,
		},
		Examples: []string{
			"Create a deployment for nginx",
			"Generate a service for my app",
			"Build a Helm chart for microservice",
			"Make a ConfigMap with these values",
		},
		Confidence: 0.8,
	},
	{
		Mode:   ModeEdit,
		Weight: 1.0,
		Patterns: []string{
			`(?i)(edit|modify|change|update|fix|patch)`,
			`(?i)(increase|decrease|set.*to|bump|scale)`,
			`(?i)(add.*to|remove.*from|replace.*with)`,
			`(?i)(values\.yaml|config|settings)`,
		},
		Examples: []string{
			"Edit the deployment to use 3 replicas",
			"Update values.yaml to increase memory",
			"Change the image tag to latest",
			"Modify the service port",
		},
		Confidence: 0.85,
	},
	{
		Mode:   ModeExplain,
		Weight: 1.0,
		Patterns: []string{
			`(?i)(what.*is|how.*does|explain|describe|tell.*about)`,
			`(?i)(difference.*between|compare|versus|vs)`,
			`(?i)(meaning.*of|purpose.*of|why.*use)`,
			`(?i)(best.*practice|recommended|should.*use)`,
		},
		Examples: []string{
			"What is a StatefulSet?",
			"Explain how ingress works",
			"What's the difference between Service types?",
			"How does pod scheduling work?",
		},
		Confidence: 0.75,
	},
	{
		Mode:   ModeCommand,
		Weight: 1.0,
		Patterns: []string{
			`(?i)(get|list|show|display)`,
			`(?i)(kubectl|helm)`,
			`(?i)(apply|delete|scale|restart)`,
			`(?i)(run.*command|execute)`,
		},
		Examples: []string{
			"List all pods in default namespace",
			"Show deployment status",
			"Scale deployment to 5 replicas",
			"Get service endpoints",
		},
		Confidence: 0.85,
	},
}

// RouteIntent performs intelligent intent classification
func RouteIntent(input string, context *KubeContextSummary) IntentRouter {
	scores := make(map[AgentMode]float64)

	// Score each mode based on pattern matching
	for _, pattern := range intentPatterns {
		score := calculatePatternScore(input, pattern)
		scores[pattern.Mode] += score * pattern.Weight
	}

	// Add contextual scoring
	contextScore := calculateContextualScore(input, context, scores)
	for mode, bonus := range contextScore {
		scores[mode] += bonus
	}

	// Find best mode and confidence
	var bestMode AgentMode
	var bestScore float64
	for mode, score := range scores {
		if score > bestScore {
			bestScore = score
			bestMode = mode
		}
	}

	// Calculate confidence (normalized to 0-1)
	confidence := math.Min(bestScore/2.0, 1.0)

	// Find alternatives for low confidence
	var alternatives []AgentMode
	if confidence < 0.7 {
		for mode, score := range scores {
			if mode != bestMode && score > bestScore*0.6 {
				alternatives = append(alternatives, mode)
			}
		}
	}

	// Generate reasoning
	reasoning := generateReasoning(input, bestMode, confidence, alternatives)

	return IntentRouter{
		Mode:         bestMode,
		Confidence:   confidence,
		Alternatives: alternatives,
		Reasoning:    reasoning,
		Policy:       getPromptPolicy(bestMode),
	}
}

// calculatePatternScore scores input against intent patterns
func calculatePatternScore(input string, pattern IntentPattern) float64 {
	score := 0.0

	for _, patternStr := range pattern.Patterns {
		if matched, _ := regexp.MatchString(patternStr, input); matched {
			score += pattern.Confidence
		}
	}

	// Bonus for exact example matches
	lowerInput := strings.ToLower(input)
	for _, example := range pattern.Examples {
		lowerExample := strings.ToLower(example)
		if strings.Contains(lowerInput, lowerExample) ||
		   calculateSimilarity(lowerInput, lowerExample) > 0.7 {
			score += 0.3
		}
	}

	return score
}

// calculateContextualScore adds context-aware scoring
func calculateContextualScore(input string, context *KubeContextSummary, scores map[AgentMode]float64) map[AgentMode]float64 {
	contextBonus := make(map[AgentMode]float64)

	if context == nil {
		return contextBonus
	}

	// If there are pod problems, boost diagnose mode
	if len(context.PodProblemCounts) > 0 {
		for problem, count := range context.PodProblemCounts {
			if count > 0 && strings.Contains(strings.ToLower(input), strings.ToLower(problem)) {
				contextBonus[ModeDiagnose] += 0.5
			}
		}
	}

	// If pods are mostly pending/failed, boost diagnose
	if context.PodPhaseCounts != nil {
		pending := context.PodPhaseCounts["Pending"]
		failed := context.PodPhaseCounts["Failed"]
		total := 0
		for _, count := range context.PodPhaseCounts {
			total += count
		}
		if total > 0 && float64(pending+failed)/float64(total) > 0.3 {
			contextBonus[ModeDiagnose] += 0.3
		}
	}

	// Production namespace detection
	if context.Namespace != "" {
		prodPatterns := []string{"prod", "production", "live", "staging"}
		for _, pattern := range prodPatterns {
			if strings.Contains(strings.ToLower(context.Namespace), pattern) {
				// Boost explain mode for production (encourage caution)
				contextBonus[ModeExplain] += 0.2
				break
			}
		}
	}

	return contextBonus
}

// calculateSimilarity provides basic string similarity scoring
func calculateSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	words1 := strings.Fields(s1)
	words2 := strings.Fields(s2)

	matches := 0
	for _, w1 := range words1 {
		for _, w2 := range words2 {
			if w1 == w2 {
				matches++
				break
			}
		}
	}

	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}

	return float64(matches) / float64(max(len(words1), len(words2)))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// generateReasoning creates human-readable reasoning for the routing decision
func generateReasoning(input string, mode AgentMode, confidence float64, alternatives []AgentMode) string {
	var reasoning strings.Builder

	switch mode {
	case ModeDiagnose:
		reasoning.WriteString("Detected troubleshooting intent - will run diagnostic checks")
	case ModeGenerate:
		reasoning.WriteString("Detected creation intent - will generate new resources")
	case ModeEdit:
		reasoning.WriteString("Detected modification intent - will create targeted diffs")
	case ModeExplain:
		reasoning.WriteString("Detected explanation request - will provide contextual information")
	case ModeCommand:
		reasoning.WriteString("Detected command request - will synthesize kubectl/helm commands")
	}

	if confidence < 0.7 && len(alternatives) > 0 {
		reasoning.WriteString(" (Low confidence - consider: ")
		for i, alt := range alternatives {
			if i > 0 {
				reasoning.WriteString(", ")
			}
			reasoning.WriteString(string(alt))
		}
		reasoning.WriteString(")")
	}

	return reasoning.String()
}

// getPromptPolicy returns mode-specific prompting policies
func getPromptPolicy(mode AgentMode) *PromptPolicy {
	policies := map[AgentMode]*PromptPolicy{
		ModeDiagnose: {
			Mode:         ModeDiagnose,
			SystemPrompt: "You are a Kubernetes diagnostic specialist. Use the ReAct protocol to systematically investigate issues. Always start with the most relevant describe/events/logs commands. Limit to 5 steps maximum. End with 'Final:' and actionable next steps.",
			Temperature:  0.2,
			MaxTokens:    1024,
			ContextStrategy: "surgical", // Use minimal, targeted context
			SafetyLevel:    "read-only",
		},
		ModeGenerate: {
			Mode:         ModeGenerate,
			SystemPrompt: "You are a Kubernetes resource generator. Create valid, production-ready YAML manifests. Always include resource limits, labels, and annotations. Propose kubectl apply --dry-run commands for validation.",
			Temperature:  0.3,
			MaxTokens:    2048,
			ContextStrategy: "schema-aware", // Include relevant schemas
			SafetyLevel:    "dry-run-first",
		},
		ModeEdit: {
			Mode:         ModeEdit,
			SystemPrompt: "You are a precision Kubernetes editor. Generate minimal, targeted unified diffs only. Preserve existing structure and comments. Always follow with validation commands.",
			Temperature:  0.25,
			MaxTokens:    1024,
			ContextStrategy: "diff-focused", // Show current vs desired state
			SafetyLevel:    "validate-first",
		},
		ModeExplain: {
			Mode:         ModeExplain,
			SystemPrompt: "You are a Kubernetes educator. Provide concise, accurate explanations with practical examples. Reference current cluster state when relevant. Keep responses under 200 words.",
			Temperature:  0.4,
			MaxTokens:    512,
			ContextStrategy: "educational", // Include relevant docs/examples
			SafetyLevel:    "informational",
		},
		ModeCommand: {
			Mode:         ModeCommand,
			SystemPrompt: "You are a kubectl/helm command synthesizer. Generate single, executable commands only. Prefer read-only operations. For mutations, always include --dry-run first.",
			Temperature:  0.2,
			MaxTokens:    256,
			ContextStrategy: "command-focused", // Minimal context, high precision
			SafetyLevel:    "prefer-read-only",
		},
	}

	if policy, exists := policies[mode]; exists {
		return policy
	}

	// Default policy
	return &PromptPolicy{
		Mode:         mode,
		SystemPrompt: "You are a helpful Kubernetes assistant.",
		Temperature:  0.3,
		MaxTokens:    1024,
		ContextStrategy: "balanced",
		SafetyLevel:    "standard",
	}
}

// QuickAction represents UI actions for low-confidence scenarios
type QuickAction struct {
	Label       string    `json:"label"`
	Mode        AgentMode `json:"mode"`
	Command     string    `json:"command"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
}

// GenerateQuickActions creates UI quick actions for ambiguous intents
func GenerateQuickActions(input string, alternatives []AgentMode, context *KubeContextSummary) []QuickAction {
	var actions []QuickAction

	for _, mode := range alternatives {
		switch mode {
		case ModeDiagnose:
			actions = append(actions, QuickAction{
				Label:       "🔍 Diagnose Issues",
				Mode:        ModeDiagnose,
				Command:     input,
				Description: "Run systematic diagnostic checks",
				Icon:        "🔍",
			})
		case ModeGenerate:
			actions = append(actions, QuickAction{
				Label:       "🏗️ Generate Resources",
				Mode:        ModeGenerate,
				Command:     input,
				Description: "Create new manifests or charts",
				Icon:        "🏗️",
			})
		case ModeEdit:
			actions = append(actions, QuickAction{
				Label:       "✏️ Edit Configuration",
				Mode:        ModeEdit,
				Command:     input,
				Description: "Modify existing resources with diffs",
				Icon:        "✏️",
			})
		case ModeExplain:
			actions = append(actions, QuickAction{
				Label:       "📚 Explain Concepts",
				Mode:        ModeExplain,
				Command:     input,
				Description: "Get detailed explanations",
				Icon:        "📚",
			})
		case ModeCommand:
			actions = append(actions, QuickAction{
				Label:       "⚡ Generate Commands",
				Mode:        ModeCommand,
				Command:     input,
				Description: "Create kubectl/helm commands",
				Icon:        "⚡",
			})
		}
	}

	// Add contextual quick actions based on cluster state
	if context != nil {
		if len(context.PodProblemCounts) > 0 {
			actions = append(actions, QuickAction{
				Label:       "🚨 Check Pod Issues",
				Mode:        ModeDiagnose,
				Command:     "investigate pod problems",
				Description: "Examine detected pod issues",
				Icon:        "🚨",
			})
		}
	}

	return actions
}

// RouterMetrics tracks routing performance for learning
type RouterMetrics struct {
	TotalRoutes       int                    `json:"total_routes"`
	ModeDistribution  map[AgentMode]int      `json:"mode_distribution"`
	ConfidenceAverage float64               `json:"confidence_average"`
	UserCorrections   map[AgentMode]int      `json:"user_corrections"`
	LastUpdated       time.Time              `json:"last_updated"`
}

// Global router metrics instance
var routerMetrics = &RouterMetrics{
	ModeDistribution: make(map[AgentMode]int),
	UserCorrections:  make(map[AgentMode]int),
	LastUpdated:      time.Now(),
}

// RecordRouting logs routing decisions for learning
func RecordRouting(input string, router IntentRouter, userAccepted bool) {
	routerMetrics.TotalRoutes++
	routerMetrics.ModeDistribution[router.Mode]++

	// Update confidence average
	if routerMetrics.TotalRoutes == 1 {
		routerMetrics.ConfidenceAverage = router.Confidence
	} else {
		// Running average
		routerMetrics.ConfidenceAverage = (routerMetrics.ConfidenceAverage*float64(routerMetrics.TotalRoutes-1) + router.Confidence) / float64(routerMetrics.TotalRoutes)
	}

	if !userAccepted {
		routerMetrics.UserCorrections[router.Mode]++
	}

	routerMetrics.LastUpdated = time.Now()
}

// GetRoutingAccuracy calculates current routing accuracy
func GetRoutingAccuracy() float64 {
	if routerMetrics.TotalRoutes == 0 {
		return 0.0
	}

	totalCorrections := 0
	for _, corrections := range routerMetrics.UserCorrections {
		totalCorrections += corrections
	}

	return 1.0 - (float64(totalCorrections) / float64(routerMetrics.TotalRoutes))
}