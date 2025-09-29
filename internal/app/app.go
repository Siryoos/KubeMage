package app

import (
	"context"
	"fmt"
	
	"github.com/siryoos/kubemage/internal/config"
	"github.com/siryoos/kubemage/internal/engine"
	"github.com/siryoos/kubemage/internal/execx"
	"github.com/siryoos/kubemage/internal/llm"
	"github.com/siryoos/kubemage/internal/ui"
)

// App represents the main application
type App struct {
	engine *engine.Engine
	ui     *ui.UI
	config *config.Config
	llm    llm.Client
	runner execx.Runner
}

// Options configures the application
type Options struct {
	Config *config.Config
	LLM    llm.Client
	Runner execx.Runner
}

// New creates a new application instance
func New(opts Options) (*App, error) {
	if opts.Config == nil {
		return nil, fmt.Errorf("config is required")
	}
	if opts.LLM == nil {
		return nil, fmt.Errorf("LLM client is required")
	}
	if opts.Runner == nil {
		return nil, fmt.Errorf("command runner is required")
	}
	
	// Create the engine with dependencies
	eng, err := engine.New(engine.Options{
		LLM:    opts.LLM,
		Runner: opts.Runner,
		Config: opts.Config,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create engine: %w", err)
	}
	
	// Create the UI
	uiInstance := ui.New(ui.Options{
		Engine: eng,
		Config: opts.Config,
	})
	
	return &App{
		engine: eng,
		ui:     uiInstance,
		config: opts.Config,
		llm:    opts.LLM,
		runner: opts.Runner,
	}, nil
}

// Run starts the application
func (a *App) Run(ctx context.Context) error {
	// Check if LLM is available
	if !a.llm.IsAvailable(ctx) {
		return fmt.Errorf("LLM service is not available. Please ensure Ollama is running")
	}
	
	// Start the UI
	return a.ui.Run(ctx)
}

// GetEngine returns the engine instance
func (a *App) GetEngine() *engine.Engine {
	return a.engine
}

// GetUI returns the UI instance
func (a *App) GetUI() *ui.UI {
	return a.ui
}

// GetConfig returns the configuration
func (a *App) GetConfig() *config.Config {
	return a.config
}