package ui

import (
	"context"
	"fmt"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/siryoos/kubemage/internal/config"
	"github.com/siryoos/kubemage/internal/engine"
)

// UI represents the terminal user interface
type UI struct {
	engine  *engine.Engine
	config  *config.Config
	program *tea.Program
}

// Options configures the UI
type Options struct {
	Engine *engine.Engine
	Config *config.Config
}

// New creates a new UI instance
func New(opts Options) *UI {
	if opts.Engine == nil {
		panic("engine is required")
	}
	if opts.Config == nil {
		panic("config is required")
	}
	
	return &UI{
		engine: opts.Engine,
		config: opts.Config,
	}
}

// Run starts the UI
func (ui *UI) Run(ctx context.Context) error {
	// Create the tea program with the model
	m := InitialModel(ui.config.GetModel(), ui.config, false)
	ui.program = tea.NewProgram(m, tea.WithAltScreen())
	
	// Run the program
	if _, err := ui.program.Run(); err != nil {
		return fmt.Errorf("error running UI: %w", err)
	}
	
	return nil
}

// Stop gracefully stops the UI
func (ui *UI) Stop() {
	if ui.program != nil {
		ui.program.Quit()
	}
}