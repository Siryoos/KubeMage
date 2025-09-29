// services.go - Service implementations using interfaces for better testability
package engine

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/siryoos/kubemage/internal/config"
	"github.com/siryoos/kubemage/internal/engine/validator"
	"github.com/siryoos/kubemage/internal/llm"
	"github.com/siryoos/kubemage/internal/metrics"
)

// OllamaService implements OllamaClient interface
type OllamaService struct {
	client  *http.Client
	baseURL string
}

func NewOllamaService() *OllamaService {
	return &OllamaService{
		client:  httpClient,
		baseURL: ollamaBaseURL(),
	}
}

func (s *OllamaService) GenerateCommand(prompt, model string) (string, error) {
	return llm.GenerateCommand(prompt, model)
}

func (s *OllamaService) StreamResponse(prompt, systemPrompt, model string) tea.Cmd {
	// This would call the actual streamOllama function
	return func() tea.Msg {
		return nil
	}
}

func (s *OllamaService) ResolveModel(preferred string, allowFallback bool) (string, string, error) {
	return resolveModel(preferred, allowFallback)
}

func (s *OllamaService) ListModels() ([]string, error) {
	// This would call the actual listOllamaModels function
	return []string{}, fmt.Errorf("not implemented")
}

// CommandExecutorService implements CommandExecutor interface
type CommandExecutorService struct{}

func NewCommandExecutorService() *CommandExecutorService {
	return &CommandExecutorService{}
}

func (s *CommandExecutorService) ExecuteCommand(command string, timeout time.Duration, p *tea.Program) tea.Cmd {
	// TODO: Implement actual command execution
	return func() tea.Msg {
		return nil
	}
}

func (s *CommandExecutorService) ExecutePreviewCheck(check validator.PreviewCheck, p *tea.Program) tea.Cmd {
	// This would call the actual execPreviewCheck function
	return func() tea.Msg {
		return nil
	}
}

func (s *CommandExecutorService) ParseCommand(command string) interface{} {
	return parseCommand(command)
}

// ContextService implements ContextProvider interface
type ContextService struct{}

func NewContextService() *ContextService {
	return &ContextService{}
}

func (s *ContextService) GetCurrentContext() (string, error) {
	return GetCurrentContext()
}

func (s *ContextService) GetCurrentNamespace() (string, error) {
	return GetCurrentNamespace()
}

func (s *ContextService) BuildContextSummary() (*KubeContextSummary, error) {
	return BuildContextSummary()
}

// ConfigService implements ConfigManager interface
type ConfigService struct{}

func NewConfigService() *ConfigService {
	return &ConfigService{}
}

func (s *ConfigService) LoadConfig() (interface{}, error) {
	return config.LoadConfig()
}

func (s *ConfigService) SaveConfig(cfg interface{}) error {
	if appCfg, ok := cfg.(*config.AppConfig); ok {
		return config.SaveConfig(appCfg)
	}
	return fmt.Errorf("invalid config type")
}

func (s *ConfigService) UpdateModelInConfig(scope, newModel string) error {
	return UpdateModelInConfig(scope, newModel)
}

func (s *ConfigService) SetActiveConfig(cfg interface{}) {
	if appCfg, ok := cfg.(*config.AppConfig); ok {
		config.SetActiveConfig(appCfg)
	}
}

func (s *ConfigService) ActiveConfig() interface{} {
	return config.ActiveConfig()
}

// ValidatorService implements Validator interface
type ValidatorService struct{}

func NewValidatorService() *ValidatorService {
	return &ValidatorService{}
}

func (s *ValidatorService) BuildPreExecPlan(cmd string) validator.PreExecPlan {
	return validator.BuildPreExecPlan(cmd)
}

func (s *ValidatorService) IsWhitelistedAction(cmd string) bool {
	// Simple whitelist check
	whitelisted := []string{"kubectl get", "kubectl describe", "kubectl logs", "helm list"}
	for _, prefix := range whitelisted {
		if strings.HasPrefix(cmd, prefix) {
			return true
		}
	}
	return false
}

func (s *ValidatorService) ValidateCommand(cmd string) error {
	plan := s.BuildPreExecPlan(cmd)
	if plan.DangerLevel == "critical" {
		return fmt.Errorf("command is too dangerous to execute: %s", cmd)
	}
	return nil
}

// MetricsService implements MetricsCollector interface
type MetricsService struct {
	metrics *SessionMetrics
}

func NewMetricsService() *MetricsService {
	return &MetricsService{
		metrics: NewSessionMetrics(),
	}
}

func (s *MetricsService) RecordSuggestion() {
	s.metrics.RecordSuggestion()
}

func (s *MetricsService) RecordValidation(passed bool) {
	s.metrics.RecordValidation(passed)
}

func (s *MetricsService) RecordEdit(applied bool) {
	// This would call the actual RecordEdit method
}

func (s *MetricsService) RecordResolution() {
	// This would call the actual RecordResolution method
}

func (s *MetricsService) RecordSafetyViolation() {
	// This would call the actual RecordSafetyViolation method
}

func (s *MetricsService) GetMetrics() *SessionMetrics {
	return s.metrics
}

func (s *MetricsService) DumpJSON(w io.Writer) error {
	// This would call the actual DumpJSON method
	return nil
}

// SecurityService implements SecurityManager interface
type SecurityService struct{}

func NewSecurityService() *SecurityService {
	return &SecurityService{}
}

func (s *SecurityService) RedactText(text string) string {
	return RedactText(text)
}

func (s *SecurityService) ValidateCommandSafety(cmd string) error {
	plan := BuildPreExecPlan(cmd)
	if plan.DangerLevel == "critical" {
		return fmt.Errorf("command contains critical safety issues: %s", cmd)
	}
	return nil
}

func (s *SecurityService) DetectSecrets(text string) []string {
	// This would implement secret detection logic
	// For now, return empty slice
	return []string{}
}

// AgentService implements Agent interface
type AgentService struct{}

func NewAgentService() *AgentService {
	return &AgentService{}
}

func (s *AgentService) StartDiagnosticSession(query string) (interface{}, error) {
	// This would call the actual StartReActSession function
	return nil, fmt.Errorf("not implemented")
}

func (s *AgentService) ExecuteAction(action string) (string, error) {
	if !s.IsActionAllowed(action) {
		return "", fmt.Errorf("action not allowed: %s", action)
	}
	// Implementation would execute the action
	return "", fmt.Errorf("not implemented")
}

func (s *AgentService) IsActionAllowed(action string) bool {
	// Simple whitelist check
	whitelisted := []string{"kubectl get", "kubectl describe", "kubectl logs", "helm list"}
	for _, prefix := range whitelisted {
		if strings.HasPrefix(action, prefix) {
			return true
		}
	}
	return false
}

func (s *AgentService) GetWhitelistedCommands() []string {
	return []string{
		"kubectl get",
		"kubectl describe",
		"kubectl logs",
		"kubectl top",
		"kubectl api-resources",
		"kubectl version",
		"kubectl explain",
		"helm lint",
		"helm template",
		"helm version",
		"helm show",
		"helm get",
	}
}

// FileService implements FileManager interface
type FileService struct{}

func NewFileService() *FileService {
	return &FileService{}
}

func (s *FileService) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (s *FileService) WriteFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

func (s *FileService) CreateBackup(path string) error {
	// Implementation would create a backup
	return fmt.Errorf("not implemented")
}

func (s *FileService) ParseDiff(diffContent string) (interface{}, error) {
	// Implementation would parse diff content
	return nil, fmt.Errorf("not implemented")
}

// ServiceContainer holds all service instances
type ServiceContainer struct {
	OllamaClient     OllamaClient
	CommandExecutor  CommandExecutor
	ContextProvider  ContextProvider
	ConfigManager    ConfigManager
	Validator        Validator
	MetricsCollector MetricsCollector
	SecurityManager  SecurityManager
	Agent            Agent
	FileManager      FileManager
}

// NewServiceContainer creates a new service container with all services
func NewServiceContainer() *ServiceContainer {
	return &ServiceContainer{
		OllamaClient:     NewOllamaService(),
		CommandExecutor:  NewCommandExecutorService(),
		ContextProvider:  NewContextService(),
		ConfigManager:    NewConfigService(),
		Validator:        NewValidatorService(),
		MetricsCollector: NewMetricsService(),
		SecurityManager:  NewSecurityService(),
		Agent:            NewAgentService(),
		FileManager:      NewFileService(),
	}
}

// Global service container instance
var services *ServiceContainer

// GetServices returns the global service container
func GetServices() *ServiceContainer {
	if services == nil {
		services = NewServiceContainer()
	}
	return services
}

// SetServices sets the global service container (useful for testing)
func SetServices(s *ServiceContainer) {
	services = s
}
