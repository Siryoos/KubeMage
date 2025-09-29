package engine

import (
	"context"
	"fmt"
	
	"github.com/siryoos/kubemage/internal/config"
	"github.com/siryoos/kubemage/internal/engine/validator"
	"github.com/siryoos/kubemage/internal/execx"
	"github.com/siryoos/kubemage/internal/llm"
)

// Engine is the main orchestrator for KubeMage
type Engine struct {
	llm                    llm.Client
	runner                 execx.Runner
	config                 *config.Config
	intelligence           *IntelligenceEngine
	facts                  *FactHelper
	knowledge              *PlaybookLibrary
	optimizer              *OptimizationAdvisor
	router                 *IntentRouter
	validator              *validator.ValidationPipeline
	modelRouter            *ModelRouter
	performanceOptimizer   *PerformanceOptimizer
	predictiveEngine       *PredictiveIntelligenceEngine
	commandGenerator       *IntelligentCommandGenerator
	performanceMonitor     *RealTimePerformanceMonitor
	recorder               *FlightRecorder
}

// Options configures the engine
type Options struct {
	LLM    llm.Client
	Runner execx.Runner
	Config *config.Config
}

// New creates a new engine instance with all dependencies
func New(opts Options) (*Engine, error) {
	if opts.LLM == nil {
		return nil, fmt.Errorf("LLM client is required")
	}
	if opts.Runner == nil {
		return nil, fmt.Errorf("command runner is required")
	}
	if opts.Config == nil {
		return nil, fmt.Errorf("config is required")
	}
	
	e := &Engine{
		llm:    opts.LLM,
		runner: opts.Runner,
		config: opts.Config,
	}
	
	// Initialize all components
	e.facts = NewFactHelper()
	e.knowledge = NewPlaybookLibrary()
	e.optimizer = NewOptimizationAdvisor()
	e.router = NewIntentRouter()
	e.intelligence = NewIntelligenceEngine()
	e.validator = validator.NewValidationPipeline()
	e.modelRouter = NewModelRouter()
	e.performanceOptimizer = NewPerformanceOptimizer()
	e.predictiveEngine = NewPredictiveIntelligenceEngine()
	e.commandGenerator = NewIntelligentCommandGenerator()
	e.performanceMonitor = NewRealTimePerformanceMonitor()
	e.recorder = NewFlightRecorder("./kubemage_data")
	
	// Wire up dependencies
	e.intelligence.facts = e.facts
	e.intelligence.knowledge = e.knowledge
	e.intelligence.optimizer = e.optimizer
	e.intelligence.router = e.router
	e.intelligence.predictive = e.predictiveEngine
	
	return e, nil
}

// GenerateCommand generates a kubectl/helm command from natural language
func (e *Engine) GenerateCommand(ctx context.Context, prompt string) (string, error) {
	// Use the LLM to generate a command
	return e.llm.Complete(ctx, prompt)
}

// GenerateCommandWithValidation generates and validates a command
func (e *Engine) GenerateCommandWithValidation(ctx context.Context, prompt string) (string, error) {
	// Generate command
	command, err := e.GenerateCommand(ctx, prompt)
	if err != nil {
		return "", err
	}
	
	// Validate command
	if e.validator != nil {
		if err := e.validator.Validate(command); err != nil {
			return "", fmt.Errorf("validation failed: %w", err)
		}
	}
	
	return command, nil
}

// ExecuteCommand executes a command using the runner
func (e *Engine) ExecuteCommand(ctx context.Context, command string) (string, string, error) {
	return e.runner.RunCommand(ctx, command)
}

// AnalyzeIntelligently performs intelligent analysis
func (e *Engine) AnalyzeIntelligently(ctx context.Context, input string) (*AnalysisSession, error) {
	// Build context summary
	contextSum, err := BuildContextSummary()
	if err != nil {
		return nil, err
	}
	
	// Perform analysis
	return e.intelligence.AnalyzeIntelligently(input, contextSum)
}

// GetFacts returns the fact helper
func (e *Engine) GetFacts() *FactHelper {
	return e.facts
}

// GetKnowledge returns the knowledge library
func (e *Engine) GetKnowledge() *PlaybookLibrary {
	return e.knowledge
}

// GetOptimizer returns the optimization advisor
func (e *Engine) GetOptimizer() *OptimizationAdvisor {
	return e.optimizer
}

// GetRouter returns the intent router
func (e *Engine) GetRouter() *IntentRouter {
	return e.router
}

// GetIntelligence returns the intelligence engine
func (e *Engine) GetIntelligence() *IntelligenceEngine {
	return e.intelligence
}

// GetValidator returns the validation pipeline
func (e *Engine) GetValidator() *validator.ValidationPipeline {
	return e.validator
}

// GetRecorder returns the flight recorder
func (e *Engine) GetRecorder() *FlightRecorder {
	return e.recorder
}