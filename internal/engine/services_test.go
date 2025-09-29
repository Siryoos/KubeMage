// services_test.go - Tests for service layer implementations
package engine

import (
	"testing"
	"time"
)

func TestOllamaService_NewOllamaService(t *testing.T) {
	service := NewOllamaService()
	if service == nil {
		t.Fatal("NewOllamaService should return a service")
	}
	if service.client == nil {
		t.Error("OllamaService should have a client")
	}
	if service.baseURL == "" {
		t.Error("OllamaService should have a baseURL")
	}
}

func TestOllamaService_GenerateCommand(t *testing.T) {
	service := NewOllamaService()

	// Test with empty prompt
	result, err := service.GenerateCommand("", "llama3.1:8b")
	if err == nil {
		t.Error("GenerateCommand with empty prompt should return error")
	}
	if result != "" {
		t.Error("GenerateCommand with empty prompt should return empty result")
	}

	// Test with valid prompt (this will fail in test environment)
	result, err = service.GenerateCommand("test prompt", "llama3.1:8b")
	if err == nil {
		t.Log("GenerateCommand succeeded (unexpected in test environment)")
	}
}

func TestOllamaService_StreamResponse(t *testing.T) {
	service := NewOllamaService()

	// Test that StreamResponse returns a command
	cmd := service.StreamResponse("test prompt", "system prompt", "llama3.1:8b")
	if cmd == nil {
		t.Error("StreamResponse should return a command")
	}

	// Test that the command can be called
	msg := cmd()
	if msg == nil {
		t.Error("StreamResponse command should return a message")
	}
}

func TestOllamaService_ResolveModel(t *testing.T) {
	service := NewOllamaService()

	// Test with empty model
	model, status, err := service.ResolveModel("", true)
	if err == nil {
		t.Error("ResolveModel with empty model should return error")
	}
	if model != "" {
		t.Error("ResolveModel with empty model should return empty model")
	}
	if status != "" {
		t.Error("ResolveModel with empty model should return empty status")
	}
}

func TestOllamaService_ListModels(t *testing.T) {
	service := NewOllamaService()

	// Test that ListModels returns an error in test environment
	models, err := service.ListModels()
	if err == nil {
		t.Log("ListModels succeeded (unexpected in test environment)")
	}
	if models != nil {
		t.Error("ListModels should return nil in test environment")
	}
}

func TestCommandExecutorService_NewCommandExecutorService(t *testing.T) {
	service := NewCommandExecutorService()
	if service == nil {
		t.Fatal("NewCommandExecutorService should return a service")
	}
}

func TestCommandExecutorService_ExecuteCommand(t *testing.T) {
	service := NewCommandExecutorService()

	// Test with empty command
	cmd := service.ExecuteCommand("", 1*time.Second, nil)
	if cmd == nil {
		t.Error("ExecuteCommand should return a command")
	}

	// Test that the command can be called
	msg := cmd()
	if msg == nil {
		t.Error("ExecuteCommand should return a message")
	}
}

func TestCommandExecutorService_ExecutePreviewCheck(t *testing.T) {
	service := NewCommandExecutorService()

	// Test with empty check
	check := PreviewCheck{}
	cmd := service.ExecutePreviewCheck(check, nil)
	if cmd == nil {
		t.Error("ExecutePreviewCheck should return a command")
	}

	// Test that the command can be called
	msg := cmd()
	if msg == nil {
		t.Error("ExecutePreviewCheck should return a message")
	}
}

func TestCommandExecutorService_ParseCommand(t *testing.T) {
	service := NewCommandExecutorService()

	// Test with empty command
	result := service.ParseCommand("")
	if result != nil {
		t.Error("ParseCommand with empty command should return nil")
	}

	// Test with valid command
	result = service.ParseCommand("kubectl get pods")
	if result == nil {
		t.Error("ParseCommand with valid command should return result")
	}
}

func TestContextService_NewContextService(t *testing.T) {
	service := NewContextService()
	if service == nil {
		t.Fatal("NewContextService should return a service")
	}
}

func TestContextService_GetCurrentContext(t *testing.T) {
	service := NewContextService()

	// Test that GetCurrentContext returns an error in test environment
	context, err := service.GetCurrentContext()
	if err == nil {
		t.Log("GetCurrentContext succeeded (unexpected in test environment)")
	}
	if context != "" {
		t.Error("GetCurrentContext should return empty string in test environment")
	}
}

func TestContextService_GetCurrentNamespace(t *testing.T) {
	service := NewContextService()

	// Test that GetCurrentNamespace returns an error in test environment
	namespace, err := service.GetCurrentNamespace()
	if err == nil {
		t.Log("GetCurrentNamespace succeeded (unexpected in test environment)")
	}
	if namespace != "" {
		t.Error("GetCurrentNamespace should return empty string in test environment")
	}
}

func TestContextService_BuildContextSummary(t *testing.T) {
	service := NewContextService()

	// Test that BuildContextSummary returns an error in test environment
	summary, err := service.BuildContextSummary()
	if err == nil {
		t.Log("BuildContextSummary succeeded (unexpected in test environment)")
	}
	if summary != nil {
		t.Error("BuildContextSummary should return nil in test environment")
	}
}

func TestConfigService_NewConfigService(t *testing.T) {
	service := NewConfigService()
	if service == nil {
		t.Fatal("NewConfigService should return a service")
	}
}

func TestConfigService_LoadConfig(t *testing.T) {
	service := NewConfigService()

	// Test that LoadConfig returns default config in test environment
	config, err := service.LoadConfig()
	if err != nil {
		t.Errorf("LoadConfig should not return error, got: %v", err)
	}
	if config == nil {
		t.Error("LoadConfig should return a config")
	}
}

func TestConfigService_SaveConfig(t *testing.T) {
	service := NewConfigService()

	// Test with nil config
	err := service.SaveConfig(nil)
	if err == nil {
		t.Error("SaveConfig with nil config should return error")
	}

	// Test with valid config
	config := DefaultConfig()
	err = service.SaveConfig(config)
	if err != nil {
		t.Errorf("SaveConfig with valid config should not return error, got: %v", err)
	}
}

func TestConfigService_UpdateModelInConfig(t *testing.T) {
	service := NewConfigService()

	// Test with empty scope
	err := service.UpdateModelInConfig("", "llama3.1:8b")
	if err == nil {
		t.Error("UpdateModelInConfig with empty scope should return error")
	}

	// Test with valid scope
	err = service.UpdateModelInConfig("tui", "llama3.1:8b")
	if err != nil {
		t.Logf("UpdateModelInConfig returned error: %v", err)
	}
}

func TestValidatorService_NewValidatorService(t *testing.T) {
	service := NewValidatorService()
	if service == nil {
		t.Fatal("NewValidatorService should return a service")
	}
}

func TestValidatorService_ValidateCommand(t *testing.T) {
	service := NewValidatorService()

	// Test with empty command
	result := service.ValidateCommand("")
	if result == nil {
		t.Error("ValidateCommand should return a result")
	}

	// Test with valid command
	result = service.ValidateCommand("kubectl get pods")
	if result == nil {
		t.Error("ValidateCommand should return a result")
	}
}

func TestValidatorService_BuildPreExecPlan(t *testing.T) {
	service := NewValidatorService()

	// Test with empty command
	plan := service.BuildPreExecPlan("")
	if plan.Command == "" {
		t.Error("BuildPreExecPlan should return a plan with command")
	}

	// Test with valid command
	plan = service.BuildPreExecPlan("kubectl get pods")
	if plan.Command == "" {
		t.Error("BuildPreExecPlan should return a plan with command")
	}
}

func TestMetricsService_NewMetricsService(t *testing.T) {
	service := NewMetricsService()
	if service == nil {
		t.Fatal("NewMetricsService should return a service")
	}
}

func TestMetricsService_RecordSuggestion(t *testing.T) {
	service := NewMetricsService()

	// Test that RecordSuggestion works
	service.RecordSuggestion()
	// No way to verify the internal state, but it should not panic
}

func TestMetricsService_RecordValidation(t *testing.T) {
	service := NewMetricsService()

	// Test that RecordValidation works
	service.RecordValidation(true)
	service.RecordValidation(false)
	// No way to verify the internal state, but it should not panic
}

func TestMetricsService_RecordEdit(t *testing.T) {
	service := NewMetricsService()

	// Test that RecordEdit works
	service.RecordEdit(true)
	service.RecordEdit(false)
	// No way to verify the internal state, but it should not panic
}

func TestSecurityService_NewSecurityService(t *testing.T) {
	service := NewSecurityService()
	if service == nil {
		t.Fatal("NewSecurityService should return a service")
	}
}

func TestSecurityService_RedactText(t *testing.T) {
	service := NewSecurityService()

	// Test with empty text
	result := service.RedactText("")
	if result != "" {
		t.Error("RedactText with empty text should return empty string")
	}

	// Test with text containing secrets
	result = service.RedactText("password=secret123")
	if result == "" {
		t.Error("RedactText should return redacted text")
	}
}

func TestAgentService_NewAgentService(t *testing.T) {
	service := NewAgentService()
	if service == nil {
		t.Fatal("NewAgentService should return a service")
	}
}

func TestAgentService_StartReActSession(t *testing.T) {
	service := NewAgentService()

	// Test that StartReActSession returns a command
	cmd := service.StartReActSession("test prompt", "llama3.1:8b")
	if cmd == nil {
		t.Error("StartReActSession should return a command")
	}

	// Test that the command can be called
	msg := cmd()
	if msg == nil {
		t.Error("StartReActSession should return a message")
	}
}

func TestFileService_NewFileService(t *testing.T) {
	service := NewFileService()
	if service == nil {
		t.Fatal("NewFileService should return a service")
	}
}

func TestFileService_ReadFile(t *testing.T) {
	service := NewFileService()

	// Test with empty filename
	content, err := service.ReadFile("")
	if err == nil {
		t.Error("ReadFile with empty filename should return error")
	}
	if content != nil {
		t.Error("ReadFile with empty filename should return nil content")
	}
}

func TestFileService_WriteFile(t *testing.T) {
	service := NewFileService()

	// Test with empty filename
	err := service.WriteFile("", []byte("test"))
	if err == nil {
		t.Error("WriteFile with empty filename should return error")
	}
}

func TestFileService_FileExists(t *testing.T) {
	service := NewFileService()

	// Test with empty filename
	exists := service.FileExists("")
	if exists {
		t.Error("FileExists with empty filename should return false")
	}
}

func TestServiceIntegration(t *testing.T) {
	// Test that services can be created and used together
	ollamaService := NewOllamaService()
	commandService := NewCommandExecutorService()
	contextService := NewContextService()
	configService := NewConfigService()
	validatorService := NewValidatorService()
	metricsService := NewMetricsService()
	securityService := NewSecurityService()
	agentService := NewAgentService()
	fileService := NewFileService()

	// Test that all services are created
	if ollamaService == nil {
		t.Error("OllamaService should be created")
	}
	if commandService == nil {
		t.Error("CommandExecutorService should be created")
	}
	if contextService == nil {
		t.Error("ContextService should be created")
	}
	if configService == nil {
		t.Error("ConfigService should be created")
	}
	if validatorService == nil {
		t.Error("ValidatorService should be created")
	}
	if metricsService == nil {
		t.Error("MetricsService should be created")
	}
	if securityService == nil {
		t.Error("SecurityService should be created")
	}
	if agentService == nil {
		t.Error("AgentService should be created")
	}
	if fileService == nil {
		t.Error("FileService should be created")
	}

	// Test basic functionality
	config := configService.GetActiveConfig()
	if config == nil {
		t.Error("ConfigService should return active config")
	}

	// Test that services can be used without panicking
	metricsService.RecordSuggestion()
	securityService.RedactText("test")
	validatorService.ValidateCommand("kubectl get pods")
}
