package engine

import (
	"time"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/siryoos/kubemage/internal/engine/validator"
	"github.com/siryoos/kubemage/internal/metrics"
)

// OllamaClient interface for LLM operations
type OllamaClient interface {
	GenerateCommand(prompt string) (string, error)
	GenerateChat(prompt string, ch chan<- string, model string, systemPrompt string)
}

// CommandExecutor interface for command execution
type CommandExecutor interface {
	ExecuteCommand(command string, timeout time.Duration, p *tea.Program) tea.Cmd
	ExecutePreviewCheck(check validator.PreviewCheck, p *tea.Program) tea.Cmd
	ParseCommand(command string) interface{}
}

// ContextProvider interface for Kubernetes context
type ContextProvider interface {
	GetCurrentContext() (string, error)
	GetNamespace() (string, error)
	BuildContextSummary() (*KubeContextSummary, error)
}

// ConfigManager interface for configuration management
type ConfigManager interface {
	GetConfig() interface{}
	UpdateConfig(key string, value interface{}) error
	SaveConfig() error
}

// Validator interface for validation operations
type Validator interface {
	ValidateCommand(command string) error
	GetPreExecPlan(command string) (*validator.PreExecPlan, error)
}

// MetricsCollector interface for metrics collection
type MetricsCollector interface {
	RecordSuggestion()
	RecordValidation(passed bool)
	RecordConfirmation()
	RecordCorrection()
	RecordExecution(command string, success bool, duration time.Duration)
	GetMetrics() *metrics.SessionMetrics
}

// SessionMetrics type alias for compatibility
type SessionMetrics = metrics.SessionMetrics

// SecurityManager interface for security operations
type SecurityManager interface {
	ValidateCommand(command string) error
	IsWhitelisted(command string) bool
	AuditAction(action string, user string, result string)
}

// Agent interface for agent operations
type Agent interface {
	StartDiagnosticSession(query string) (interface{}, error)
	ExecuteAction(action string) (string, error)
	IsActionAllowed(action string) bool
	GetWhitelistedCommands() []string
}

// FileManager interface for file operations
type FileManager interface {
	ReadFile(path string) (string, error)
	WriteFile(path string, content string) error
	FileExists(path string) bool
	CreateDirectory(path string) error
}