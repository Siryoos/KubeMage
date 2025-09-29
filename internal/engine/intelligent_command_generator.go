// intelligent_command_generator.go - Advanced command generation with optimization and caching
package engine

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// IntelligentCommandGenerator provides advanced command generation capabilities
type IntelligentCommandGenerator struct {
	modelRouter        *ModelRouter
	smartCache         *SmartCacheSystem
	contextAnalyzer    *ContextAnalyzer
	commandOptimizer   *CommandOptimizer
	safetyValidator    *SafetyValidator
	templateEngine     *TemplateEngine
	performanceTracker *CommandPerformanceTracker
	mu                 sync.RWMutex
}

// CommandOptimizer optimizes generated commands for performance and safety
type CommandOptimizer struct {
	commandPatterns    map[string]*CommandPattern
	optimizationRules  []OptimizationRule
	performanceTracker *CommandPerformanceTracker
	safetyRules        []SafetyRule
	mu                 sync.RWMutex
}

// SafetyValidator validates commands for safety and compliance
type SafetyValidator struct {
	riskPatterns       map[string]*RiskPattern
	complianceRules    []ComplianceRule
	auditLogger        *AuditLogger
	riskThreshold      float64
	mu                 sync.RWMutex
}

// TemplateEngine manages command templates and generation
type TemplateEngine struct {
	templates          map[string]*CommandTemplate
	templateCache      map[string]*CachedTemplate
	variableResolver   *VariableResolver
	contextInjector    *ContextInjector
	mu                 sync.RWMutex
}

// CommandPerformanceTracker tracks command performance metrics
type CommandPerformanceTracker struct {
	commandMetrics     map[string]*CommandMetrics
	executionHistory   []CommandExecution
	optimizationHints  map[string]*OptimizationHint
	mu                 sync.RWMutex
}

// GeneratedCommand represents a generated command with metadata
type GeneratedCommand struct {
	Command            string                 `json:"command"`
	Explanation        string                 `json:"explanation"`
	RiskLevel          string                 `json:"risk_level"`
	SafetyChecks       []SafetyCheck          `json:"safety_checks"`
	OptimizationHints  []string               `json:"optimization_hints"`
	AlternativeCommands []string              `json:"alternative_commands"`
	Prerequisites      []string               `json:"prerequisites"`
	ExpectedOutput     string                 `json:"expected_output"`
	EstimatedDuration  time.Duration          `json:"estimated_duration"`
	Confidence         float64                `json:"confidence"`
	Context            *KubeContextSummary    `json:"context"`
	Metadata           map[string]interface{} `json:"metadata"`
	GeneratedAt        time.Time              `json:"generated_at"`
}

// CommandPattern is now defined in types.go

// OptimizationRule is now defined in types.go

// SafetyRule represents a safety validation rule
type SafetyRule struct {
	Name        string                    `json:"name"`
	Pattern     string                    `json:"pattern"`
	RiskLevel   string                    `json:"risk_level"`
	Condition   func(string) bool         `json:"-"`
	Mitigation  string                    `json:"mitigation"`
	Enabled     bool                      `json:"enabled"`
	Description string                    `json:"description"`
}

// RiskPattern represents a risk pattern in commands
type RiskPattern struct {
	Pattern     string   `json:"pattern"`
	RiskLevel   string   `json:"risk_level"`
	Indicators  []string `json:"indicators"`
	Mitigations []string `json:"mitigations"`
	Examples    []string `json:"examples"`
	Frequency   int      `json:"frequency"`
}

// ComplianceRule represents a compliance validation rule
type ComplianceRule struct {
	Name        string                    `json:"name"`
	Standard    string                    `json:"standard"`
	Requirement string                    `json:"requirement"`
	Validator   func(string) bool         `json:"-"`
	Severity    string                    `json:"severity"`
	Enabled     bool                      `json:"enabled"`
}

// AuditLogger logs command generation and execution for audit purposes
type AuditLogger struct {
	entries     []AuditEntry
	maxEntries  int
	mu          sync.RWMutex
}

// AuditEntry represents an audit log entry
type AuditEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Action      string                 `json:"action"`
	Command     string                 `json:"command"`
	User        string                 `json:"user"`
	Context     *KubeContextSummary    `json:"context"`
	RiskLevel   string                 `json:"risk_level"`
	Approved    bool                   `json:"approved"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// CommandTemplate represents a command template
type CommandTemplate struct {
	Name        string            `json:"name"`
	Template    string            `json:"template"`
	Variables   []TemplateVariable `json:"variables"`
	Category    string            `json:"category"`
	RiskLevel   string            `json:"risk_level"`
	Description string            `json:"description"`
	Examples    []string          `json:"examples"`
	LastUsed    time.Time         `json:"last_used"`
	UsageCount  int               `json:"usage_count"`
}

// TemplateVariable represents a template variable
type TemplateVariable struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Required     bool     `json:"required"`
	DefaultValue string   `json:"default_value"`
	Validation   string   `json:"validation"`
	Options      []string `json:"options"`
	Description  string   `json:"description"`
}

// CachedTemplate represents a cached template
type CachedTemplate struct {
	Template    *CommandTemplate `json:"template"`
	RenderedAt  time.Time        `json:"rendered_at"`
	TTL         time.Duration    `json:"ttl"`
	Variables   map[string]string `json:"variables"`
}

// VariableResolver resolves template variables
type VariableResolver struct {
	resolvers map[string]func(*KubeContextSummary) string
	cache     map[string]string
	mu        sync.RWMutex
}

// ContextInjector injects context into templates
type ContextInjector struct {
	injectors map[string]func(*KubeContextSummary, string) string
	mu        sync.RWMutex
}

// CommandMetrics represents performance metrics for a command
type CommandMetrics struct {
	Command         string        `json:"command"`
	ExecutionCount  int           `json:"execution_count"`
	SuccessRate     float64       `json:"success_rate"`
	AverageTime     time.Duration `json:"average_time"`
	ErrorPatterns   []string      `json:"error_patterns"`
	OptimizedTime   time.Duration `json:"optimized_time"`
	LastExecuted    time.Time     `json:"last_executed"`
	Confidence      float64       `json:"confidence"`
}

// CommandExecution represents a command execution record
type CommandExecution struct {
	Command     string        `json:"command"`
	Success     bool          `json:"success"`
	Duration    time.Duration `json:"duration"`
	Error       string        `json:"error"`
	Context     *KubeContextSummary `json:"context"`
	Timestamp   time.Time     `json:"timestamp"`
	Optimized   bool          `json:"optimized"`
}

// SafetyCheck represents a safety validation check
type SafetyCheck struct {
	Name        string `json:"name"`
	Passed      bool   `json:"passed"`
	RiskLevel   string `json:"risk_level"`
	Message     string `json:"message"`
	Mitigation  string `json:"mitigation"`
}

// NewIntelligentCommandGenerator creates a new intelligent command generator
func NewIntelligentCommandGenerator(modelRouter *ModelRouter, smartCache *SmartCacheSystem) *IntelligentCommandGenerator {
	return &IntelligentCommandGenerator{
		modelRouter:     modelRouter,
		smartCache:      smartCache,
		contextAnalyzer: &ContextAnalyzer{
			contextPatterns: make(map[string]*ContextPattern),
			riskAssessment: &RiskAssessment{
				riskLevels:    make(map[string]float64),
				safetyModels:  make(map[string]string),
				riskThreshold: 0.7,
			},
			performanceHints: make(map[string]string),
		},
		commandOptimizer: &CommandOptimizer{
			commandPatterns:   make(map[string]*CommandPattern),
			optimizationRules: make([]OptimizationRule, 0),
			performanceTracker: &CommandPerformanceTracker{
				commandMetrics:    make(map[string]*CommandMetrics),
				executionHistory:  make([]CommandExecution, 0),
				optimizationHints: make(map[string]*OptimizationHint),
			},
			safetyRules: make([]SafetyRule, 0),
		},
		safetyValidator: &SafetyValidator{
			riskPatterns:    make(map[string]*RiskPattern),
			complianceRules: make([]ComplianceRule, 0),
			auditLogger: &AuditLogger{
				entries:    make([]AuditEntry, 0),
				maxEntries: 10000,
			},
			riskThreshold: 0.7,
		},
		templateEngine: &TemplateEngine{
			templates:     make(map[string]*CommandTemplate),
			templateCache: make(map[string]*CachedTemplate),
			variableResolver: &VariableResolver{
				resolvers: make(map[string]func(*KubeContextSummary) string),
				cache:     make(map[string]string),
			},
			contextInjector: &ContextInjector{
				injectors: make(map[string]func(*KubeContextSummary, string) string),
			},
		},
		performanceTracker: &CommandPerformanceTracker{
			commandMetrics:    make(map[string]*CommandMetrics),
			executionHistory:  make([]CommandExecution, 0),
			optimizationHints: make(map[string]*OptimizationHint),
		},
	}
}

// GenerateOptimizedCommand generates an optimized command with caching
func (icg *IntelligentCommandGenerator) GenerateOptimizedCommand(query string, context *KubeContextSummary) (*GeneratedCommand, error) {
	icg.mu.RLock()
	defer icg.mu.RUnlock()

	// Check cache first
	cacheKey := fmt.Sprintf("command:%s:%s", query, context.Hash())
	if cached := icg.smartCache.GetL1(cacheKey); cached != nil {
		return cached.(*GeneratedCommand), nil
	}

	// Select appropriate model
	var model string
	var err error
	if icg.modelRouter != nil {
		model, err = icg.modelRouter.SelectModel(query, context)
		if err != nil {
			return nil, fmt.Errorf("model selection failed: %w", err)
		}
	} else {
		model = "llama3.1:8b" // Default model
	}

	// Generate base command
	baseCommand, err := icg.generateBaseCommand(query, context, model)
	if err != nil {
		return nil, fmt.Errorf("base command generation failed: %w", err)
	}

	// Optimize the command
	optimizedCommand := icg.commandOptimizer.OptimizeCommand(baseCommand, context)

	// Validate safety
	safetyChecks := icg.safetyValidator.ValidateCommand(optimizedCommand, context)

	// Calculate risk level
	riskLevel := icg.calculateRiskLevel(optimizedCommand, safetyChecks)

	// Generate alternatives
	alternatives := icg.generateAlternatives(optimizedCommand, context)

	// Create generated command
	generatedCmd := &GeneratedCommand{
		Command:             optimizedCommand,
		Explanation:         icg.generateExplanation(optimizedCommand, context),
		RiskLevel:           riskLevel,
		SafetyChecks:        safetyChecks,
		OptimizationHints:   icg.generateOptimizationHints(optimizedCommand),
		AlternativeCommands: alternatives,
		Prerequisites:       icg.generatePrerequisites(optimizedCommand, context),
		ExpectedOutput:      icg.generateExpectedOutput(optimizedCommand, context),
		EstimatedDuration:   icg.estimateDuration(optimizedCommand),
		Confidence:          icg.calculateConfidence(optimizedCommand, context),
		Context:             context,
		Metadata:            make(map[string]interface{}),
		GeneratedAt:         time.Now(),
	}

	// Add metadata
	generatedCmd.Metadata["model_used"] = model
	generatedCmd.Metadata["optimization_applied"] = optimizedCommand != baseCommand
	generatedCmd.Metadata["cache_key"] = cacheKey

	// Cache the result
	icg.smartCache.SetL1(cacheKey, generatedCmd, 10*time.Minute)

	// Log for audit
	icg.safetyValidator.auditLogger.LogGeneration(generatedCmd)

	return generatedCmd, nil
}

// generateBaseCommand generates the base command using LLM
func (icg *IntelligentCommandGenerator) generateBaseCommand(query string, context *KubeContextSummary, model string) (string, error) {
	// This would integrate with the existing Ollama generation
	// For now, return a placeholder
	return fmt.Sprintf("kubectl get pods -n %s", context.Namespace), nil
}

// OptimizeCommand optimizes a command for performance and safety
func (co *CommandOptimizer) OptimizeCommand(command string, context *KubeContextSummary) string {
	co.mu.RLock()
	defer co.mu.RUnlock()

	optimized := command

	// Apply optimization rules
	for _, rule := range co.optimizationRules {
		if rule.Enabled && rule.Condition(optimized) {
			optimized = rule.Transform(optimized)
		}
	}

	// Apply context-specific optimizations
	optimized = co.applyContextOptimizations(optimized, context)

	// Apply performance optimizations
	optimized = co.applyPerformanceOptimizations(optimized)

	return optimized
}

// applyContextOptimizations applies context-specific optimizations
func (co *CommandOptimizer) applyContextOptimizations(command string, context *KubeContextSummary) string {
	// Add namespace if not specified
	if !strings.Contains(command, "-n ") && !strings.Contains(command, "--namespace") {
		if context.Namespace != "default" {
			command = strings.Replace(command, "kubectl ", fmt.Sprintf("kubectl -n %s ", context.Namespace), 1)
		}
	}

	// Add output format for better parsing
	if strings.Contains(command, "kubectl get") && !strings.Contains(command, "-o ") {
		command += " -o wide"
	}

	return command
}

// applyPerformanceOptimizations applies performance optimizations
func (co *CommandOptimizer) applyPerformanceOptimizations(command string) string {
	// Add timeout for long-running commands
	if strings.Contains(command, "kubectl logs") && !strings.Contains(command, "--tail") {
		command += " --tail=100"
	}

	// Add field selectors for efficiency
	if strings.Contains(command, "kubectl get pods") && !strings.Contains(command, "--field-selector") {
		// Only get running pods by default
		command += " --field-selector=status.phase=Running"
	}

	return command
}

// ValidateCommand validates a command for safety and compliance
func (sv *SafetyValidator) ValidateCommand(command string, context *KubeContextSummary) []SafetyCheck {
	sv.mu.RLock()
	defer sv.mu.RUnlock()

	var checks []SafetyCheck

	// Check against risk patterns
	for _, pattern := range sv.riskPatterns {
		if strings.Contains(strings.ToLower(command), strings.ToLower(pattern.Pattern)) {
			checks = append(checks, SafetyCheck{
				Name:       fmt.Sprintf("Risk Pattern: %s", pattern.Pattern),
				Passed:     false,
				RiskLevel:  pattern.RiskLevel,
				Message:    fmt.Sprintf("Command matches high-risk pattern: %s", pattern.Pattern),
				Mitigation: strings.Join(pattern.Mitigations, "; "),
			})
		}
	}

	// Check against safety rules
	for _, rule := range sv.safetyRules {
		if rule.Enabled && rule.Condition(command) {
			checks = append(checks, SafetyCheck{
				Name:       rule.Name,
				Passed:     false,
				RiskLevel:  rule.RiskLevel,
				Message:    rule.Description,
				Mitigation: rule.Mitigation,
			})
		}
	}

	// Production environment checks
	if strings.Contains(strings.ToLower(context.Namespace), "prod") {
		checks = append(checks, SafetyCheck{
			Name:       "Production Environment",
			Passed:     !strings.Contains(command, "delete"),
			RiskLevel:  "high",
			Message:    "Command executed in production environment",
			Mitigation: "Use --dry-run flag first",
		})
	}

	return checks
}

// calculateRiskLevel calculates overall risk level
func (icg *IntelligentCommandGenerator) calculateRiskLevel(command string, checks []SafetyCheck) string {
	highRiskCount := 0
	mediumRiskCount := 0

	for _, check := range checks {
		if !check.Passed {
			switch check.RiskLevel {
			case "high", "critical":
				highRiskCount++
			case "medium":
				mediumRiskCount++
			}
		}
	}

	if highRiskCount > 0 {
		return "high"
	} else if mediumRiskCount > 0 {
		return "medium"
	}

	return "low"
}

// generateAlternatives generates alternative commands
func (icg *IntelligentCommandGenerator) generateAlternatives(command string, context *KubeContextSummary) []string {
	var alternatives []string

	// Generate safer alternatives
	if strings.Contains(command, "delete") {
		saferCmd := strings.Replace(command, "delete", "get", 1)
		alternatives = append(alternatives, saferCmd+" # Safer: view before delete")
	}

	// Generate more specific alternatives
	if strings.Contains(command, "kubectl get pods") {
		alternatives = append(alternatives, command+" --show-labels")
		alternatives = append(alternatives, command+" -o json")
	}

	return alternatives
}

// generateExplanation generates command explanation
func (icg *IntelligentCommandGenerator) generateExplanation(command string, context *KubeContextSummary) string {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "No command to explain"
	}

	explanation := fmt.Sprintf("This command uses %s to", parts[0])

	if len(parts) > 1 {
		switch parts[1] {
		case "get":
			explanation += " retrieve and display resources"
		case "describe":
			explanation += " show detailed information about resources"
		case "delete":
			explanation += " remove resources (DESTRUCTIVE)"
		case "apply":
			explanation += " create or update resources"
		case "logs":
			explanation += " display container logs"
		default:
			explanation += fmt.Sprintf(" execute the '%s' operation", parts[1])
		}
	}

	if context.Namespace != "default" {
		explanation += fmt.Sprintf(" in the '%s' namespace", context.Namespace)
	}

	return explanation
}

// generateOptimizationHints generates optimization hints
func (icg *IntelligentCommandGenerator) generateOptimizationHints(command string) []string {
	var hints []string

	if strings.Contains(command, "kubectl get") && !strings.Contains(command, "-o ") {
		hints = append(hints, "Add -o json for machine-readable output")
	}

	if strings.Contains(command, "kubectl logs") && !strings.Contains(command, "--tail") {
		hints = append(hints, "Add --tail=N to limit log output")
	}

	if !strings.Contains(command, "--dry-run") && (strings.Contains(command, "apply") || strings.Contains(command, "delete")) {
		hints = append(hints, "Consider using --dry-run=client first")
	}

	return hints
}

// generatePrerequisites generates command prerequisites
func (icg *IntelligentCommandGenerator) generatePrerequisites(command string, context *KubeContextSummary) []string {
	var prerequisites []string

	prerequisites = append(prerequisites, "kubectl must be installed and configured")
	prerequisites = append(prerequisites, fmt.Sprintf("Access to cluster '%s'", context.Context))

	if context.Namespace != "default" {
		prerequisites = append(prerequisites, fmt.Sprintf("Access to namespace '%s'", context.Namespace))
	}

	if strings.Contains(command, "delete") || strings.Contains(command, "apply") {
		prerequisites = append(prerequisites, "Appropriate RBAC permissions for write operations")
	}

	return prerequisites
}

// generateExpectedOutput generates expected output description
func (icg *IntelligentCommandGenerator) generateExpectedOutput(command string, context *KubeContextSummary) string {
	if strings.Contains(command, "get pods") {
		return "List of pods with their status, age, and other details"
	} else if strings.Contains(command, "describe") {
		return "Detailed information about the specified resource"
	} else if strings.Contains(command, "logs") {
		return "Container log output"
	} else if strings.Contains(command, "delete") {
		return "Confirmation of resource deletion"
	}

	return "Command output will vary based on the operation"
}

// estimateDuration estimates command execution duration
func (icg *IntelligentCommandGenerator) estimateDuration(command string) time.Duration {
	// Simple estimation based on command type
	if strings.Contains(command, "get") {
		return 2 * time.Second
	} else if strings.Contains(command, "describe") {
		return 3 * time.Second
	} else if strings.Contains(command, "logs") {
		return 5 * time.Second
	} else if strings.Contains(command, "apply") {
		return 10 * time.Second
	} else if strings.Contains(command, "delete") {
		return 5 * time.Second
	}

	return 3 * time.Second
}

// calculateConfidence calculates generation confidence
func (icg *IntelligentCommandGenerator) calculateConfidence(command string, context *KubeContextSummary) float64 {
	confidence := 0.7 // Base confidence

	// Increase confidence for simple commands
	if strings.Contains(command, "get") || strings.Contains(command, "describe") {
		confidence += 0.2
	}

	// Decrease confidence for complex commands
	if strings.Contains(command, "delete") || strings.Contains(command, "patch") {
		confidence -= 0.1
	}

	// Adjust based on context completeness
	if context.Namespace != "" && context.Context != "" {
		confidence += 0.1
	}

	// Cap at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// LogGeneration logs command generation for audit
func (al *AuditLogger) LogGeneration(cmd *GeneratedCommand) {
	al.mu.Lock()
	defer al.mu.Unlock()

	entry := AuditEntry{
		Timestamp: time.Now(),
		Action:    "command_generated",
		Command:   cmd.Command,
		User:      "system", // Would be actual user in real implementation
		Context:   cmd.Context,
		RiskLevel: cmd.RiskLevel,
		Approved:  false, // Will be updated when executed
		Metadata: map[string]interface{}{
			"confidence":    cmd.Confidence,
			"model_used":    cmd.Metadata["model_used"],
			"optimized":     cmd.Metadata["optimization_applied"],
		},
	}

	al.entries = append(al.entries, entry)

	// Maintain max entries
	if len(al.entries) > al.maxEntries {
		al.entries = al.entries[len(al.entries)-al.maxEntries:]
	}
}

// GetGenerationStats returns command generation statistics
func (icg *IntelligentCommandGenerator) GetGenerationStats() map[string]interface{} {
	icg.mu.RLock()
	defer icg.mu.RUnlock()

	return map[string]interface{}{
		"command_patterns":    len(icg.commandOptimizer.commandPatterns),
		"optimization_rules":  len(icg.commandOptimizer.optimizationRules),
		"safety_rules":        len(icg.commandOptimizer.safetyRules),
		"risk_patterns":       len(icg.safetyValidator.riskPatterns),
		"compliance_rules":    len(icg.safetyValidator.complianceRules),
		"templates":           len(icg.templateEngine.templates),
		"cached_templates":    len(icg.templateEngine.templateCache),
		"audit_entries":       len(icg.safetyValidator.auditLogger.entries),
		"execution_history":   len(icg.performanceTracker.executionHistory),
	}
}

// Removed global instance - now created via dependency injection

// InitializeIntelligentCommandGenerator creates a new command generator instance
func InitializeIntelligentCommandGenerator(modelRouter *ModelRouter, smartCache *SmartCacheSystem) *IntelligentCommandGenerator {
	return NewIntelligentCommandGenerator(modelRouter, smartCache)
}

