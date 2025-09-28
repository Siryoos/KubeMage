// interfaces.go - Interface definitions for better testability and modularity
package main

import (
	"io"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// OllamaClient interface for LLM operations
type OllamaClient interface {
	GenerateCommand(prompt, model string) (string, error)
	StreamResponse(prompt, systemPrompt, model string) tea.Cmd
	ResolveModel(preferred string, allowFallback bool) (string, string, error)
	ListModels() ([]string, error)
}

// CommandExecutor interface for command execution
type CommandExecutor interface {
	ExecuteCommand(command string, timeout time.Duration, p *tea.Program) tea.Cmd
	ExecutePreviewCheck(check PreviewCheck, p *tea.Program) tea.Cmd
	ParseCommand(command string) interface{} // Using interface{} to avoid circular dependency
}

// ContextProvider interface for Kubernetes context operations
type ContextProvider interface {
	GetCurrentContext() (string, error)
	GetCurrentNamespace() (string, error)
	BuildContextSummary() (*KubeContextSummary, error)
}

// ConfigManager interface for configuration operations
type ConfigManager interface {
	LoadConfig() (*AppConfig, error)
	SaveConfig(cfg *AppConfig) error
	UpdateModelInConfig(scope, newModel string) error
	SetActiveConfig(cfg *AppConfig)
	ActiveConfig() *AppConfig
}

// Validator interface for command validation
type Validator interface {
	BuildPreExecPlan(cmd string) PreExecPlan
	IsWhitelistedAction(cmd string) bool
	ValidateCommand(cmd string) error
}

// MetricsCollector interface for metrics tracking
type MetricsCollector interface {
	RecordSuggestion()
	RecordValidation(passed bool)
	RecordEdit(applied bool)
	RecordResolution()
	RecordSafetyViolation()
	GetMetrics() *SessionMetrics
	DumpJSON(w io.Writer) error
}

// UIComponent interface for UI components
type UIComponent interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (tea.Model, tea.Cmd)
	View() string
}

// IntelligenceAnalyzer interface for smart analysis
type IntelligenceAnalyzer interface {
	AnalyzeContext(ctx *KubeContextSummary) (*AnalysisSession, error)
	GenerateRecommendations(session *AnalysisSession) ([]Recommendation, error)
	AssessRisk(action string, context *KubeContextSummary) (*RiskLevel, error)
}

// FileManager interface for file operations
type FileManager interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte) error
	CreateBackup(path string) error
	ParseDiff(diffContent string) (interface{}, error)
}

// SecurityManager interface for security operations
type SecurityManager interface {
	RedactText(text string) string
	ValidateCommandSafety(cmd string) error
	DetectSecrets(text string) []string
}

// Agent interface for autonomous operations
type Agent interface {
	StartDiagnosticSession(query string) (interface{}, error)
	ExecuteAction(action string) (string, error)
	IsActionAllowed(action string) bool
	GetWhitelistedCommands() []string
}

// NotificationHandler interface for user notifications
type NotificationHandler interface {
	ShowToast(message string, level string)
	ShowModal(title, content string)
	ShowError(err error)
	ShowWarning(message string)
}

// ThemeManager interface for UI theming
type ThemeManager interface {
	GetTheme() interface{}
	SetTheme(name string) error
	GetAvailableThemes() []string
}

// LogManager interface for log operations
type LogManager interface {
	ParseLogs(content string) []interface{}
	FilterLogs(logs []interface{}, filter interface{}) []interface{}
	HighlightSearchTerm(content, term string) string
}

// DiffManager interface for diff operations
type DiffManager interface {
	ParseDiff(diffContent string) (interface{}, error)
	ApplyDiff(diff interface{}, targetPath string) error
	ValidateDiff(diff interface{}) error
}

// ModelResolver interface for model management
type ModelResolver interface {
	ResolveModel(preferred string, allowFallback bool) (string, string, error)
	ListAvailableModels() ([]string, error)
	ValidateModel(model string) error
}

// ContextBuilder interface for context building
type ContextBuilder interface {
	BuildKubeContext() (*KubeContextSummary, error)
	BuildPromptContext(query string) string
	InjectContext(prompt string) string
}

// CommandParser interface for command parsing
type CommandParser interface {
	ParseKubectlCommand(cmd string) (interface{}, error)
	ParseHelmCommand(cmd string) (interface{}, error)
	ParseGenDeployCommand(cmd string) (interface{}, error)
	ParseGenHelmCommand(cmd string) (interface{}, error)
}

// SessionManager interface for session management
type SessionManager interface {
	CreateSession() interface{}
	SaveSession(session interface{}) error
	LoadSession(id string) (interface{}, error)
	DeleteSession(id string) error
}

// CacheManager interface for caching operations
type CacheManager interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
	Delete(key string)
	Clear()
}

// EventBus interface for event handling
type EventBus interface {
	Subscribe(eventType string, handler func(interface{}))
	Publish(eventType string, data interface{})
	Unsubscribe(eventType string, handler func(interface{}))
}

// ResourceManager interface for resource management
type ResourceManager interface {
	GetResourceInfo(resourceType, name string) (interface{}, error)
	ListResources(resourceType string) ([]interface{}, error)
	DescribeResource(resourceType, name string) (string, error)
}

// PromptBuilder interface for prompt construction
type PromptBuilder interface {
	BuildSystemPrompt(mode string) string
	BuildUserPrompt(query string, context *KubeContextSummary) string
	BuildAgentPrompt(query string) string
}

// ResponseProcessor interface for response processing
type ResponseProcessor interface {
	ProcessCommandResponse(response string) (string, error)
	ProcessChatResponse(response string) (string, error)
	ProcessAgentResponse(response string) (string, error)
}


// CommandResult represents the result of a command execution
type CommandResult struct {
	Command   string `json:"command"`
	Output    string `json:"output"`
	Error     string `json:"error,omitempty"`
	ExitCode  int    `json:"exit_code"`
	Duration  time.Duration `json:"duration"`
}

// AnalysisResult represents the result of an analysis operation
type AnalysisResult struct {
	SessionID string                 `json:"session_id"`
	Findings  []string               `json:"findings"`
	Actions   []IntelligentAction    `json:"actions"`
	Confidence float64               `json:"confidence"`
	Timestamp time.Time              `json:"timestamp"`
}
