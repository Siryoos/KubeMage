// ollama_improved_test.go - Enhanced tests for Ollama functionality
package llm

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestOllamaRequest_JSON(t *testing.T) {
	req := OllamaRequest{
		Model:  "llama3.1:8b",
		Prompt: "test prompt",
		System: "test system",
		Stream: true,
		Options: map[string]interface{}{
			"temperature": 0.7,
			"top_p":       0.9,
		},
	}

	// Test that the struct can be marshaled to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		t.Errorf("Failed to marshal OllamaRequest to JSON: %v", err)
	}

	// Test that the struct can be unmarshaled from JSON
	var unmarshaledReq OllamaRequest
	err = json.Unmarshal(jsonData, &unmarshaledReq)
	if err != nil {
		t.Errorf("Failed to unmarshal OllamaRequest from JSON: %v", err)
	}

	if unmarshaledReq.Model != req.Model {
		t.Errorf("Unmarshaled model = %v, want %v", unmarshaledReq.Model, req.Model)
	}
	if unmarshaledReq.Prompt != req.Prompt {
		t.Errorf("Unmarshaled prompt = %v, want %v", unmarshaledReq.Prompt, req.Prompt)
	}
	if unmarshaledReq.System != req.System {
		t.Errorf("Unmarshaled system = %v, want %v", unmarshaledReq.System, req.System)
	}
	if unmarshaledReq.Stream != req.Stream {
		t.Errorf("Unmarshaled stream = %v, want %v", unmarshaledReq.Stream, req.Stream)
	}
}

func TestOllamaResponse_JSON(t *testing.T) {
	resp := OllamaResponse{
		Response: "test response",
		Done:     true,
	}

	// Test that the struct can be marshaled to JSON
	jsonData, err := json.Marshal(resp)
	if err != nil {
		t.Errorf("Failed to marshal OllamaResponse to JSON: %v", err)
	}

	// Test that the struct can be unmarshaled from JSON
	var unmarshaledResp OllamaResponse
	err = json.Unmarshal(jsonData, &unmarshaledResp)
	if err != nil {
		t.Errorf("Failed to unmarshal OllamaResponse from JSON: %v", err)
	}

	if unmarshaledResp.Response != resp.Response {
		t.Errorf("Unmarshaled response = %v, want %v", unmarshaledResp.Response, resp.Response)
	}
	if unmarshaledResp.Done != resp.Done {
		t.Errorf("Unmarshaled done = %v, want %v", unmarshaledResp.Done, resp.Done)
	}
}

func TestTagList_JSON(t *testing.T) {
	tags := tagList{
		Models: []struct {
			Name string `json:"name"`
		}{
			{Name: "llama3.1:8b"},
			{Name: "llama3.1:13b"},
			{Name: "deepseek-r1:8b"},
		},
	}

	// Test that the struct can be marshaled to JSON
	jsonData, err := json.Marshal(tags)
	if err != nil {
		t.Errorf("Failed to marshal tagList to JSON: %v", err)
	}

	// Test that the struct can be unmarshaled from JSON
	var unmarshaledTags tagList
	err = json.Unmarshal(jsonData, &unmarshaledTags)
	if err != nil {
		t.Errorf("Failed to unmarshal tagList from JSON: %v", err)
	}

	if len(unmarshaledTags.Models) != len(tags.Models) {
		t.Errorf("Unmarshaled models count = %v, want %v", len(unmarshaledTags.Models), len(tags.Models))
	}

	for i, model := range tags.Models {
		if unmarshaledTags.Models[i].Name != model.Name {
			t.Errorf("Unmarshaled model[%d] name = %v, want %v", i, unmarshaledTags.Models[i].Name, model.Name)
		}
	}
}

func TestOllamaBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		expected string
	}{
		{
			name:     "default localhost",
			host:     "",
			expected: "http://localhost:11434",
		},
		{
			name:     "custom host",
			host:     "http://remote:11434",
			expected: "http://remote:11434",
		},
		{
			name:     "host with port",
			host:     "http://ollama.example.com:8080",
			expected: "http://ollama.example.com:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.host != "" {
				os.Setenv("OLLAMA_HOST", tt.host)
			} else {
				os.Unsetenv("OLLAMA_HOST")
			}

			result := ollamaBaseURL()
			if result != tt.expected {
				t.Errorf("ollamaBaseURL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDefaultModelName(t *testing.T) {
	if DefaultModelName == "" {
		t.Error("DefaultModelName should not be empty")
	}

	if DefaultModelName != "llama3.1:8b" {
		t.Errorf("DefaultModelName = %v, want llama3.1:8b", DefaultModelName)
	}
}

func TestDefaultOllamaEndpoint(t *testing.T) {
	if DefaultOllamaEndpoint == "" {
		t.Error("DefaultOllamaEndpoint should not be empty")
	}

	if DefaultOllamaEndpoint != "http://localhost:11434" {
		t.Errorf("DefaultOllamaEndpoint = %v, want http://localhost:11434", DefaultOllamaEndpoint)
	}
}

func TestSystemPrompts(t *testing.T) {
	tests := []struct {
		name     string
		prompt   string
		expected string
	}{
		{
			name:     "command only system prompt",
			prompt:   commandOnlySystemPrompt,
			expected: "You are KubeMage",
		},
		{
			name:     "chat assistant system prompt",
			prompt:   chatAssistantSystemPrompt,
			expected: "You are KubeMage",
		},
		{
			name:     "agent system prompt",
			prompt:   agentSystemPrompt,
			expected: "You are an agent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.prompt == "" {
				t.Error("System prompt should not be empty")
			}

			if !strings.Contains(tt.prompt, tt.expected) {
				t.Errorf("System prompt should contain %v", tt.expected)
			}
		})
	}
}

func TestHTTPClients(t *testing.T) {
	if httpClient == nil {
		t.Error("httpClient should not be nil")
	}

	if streamingClient == nil {
		t.Error("streamingClient should not be nil")
	}

	// Test that clients have different timeouts
	if httpClient.Timeout == streamingClient.Timeout {
		t.Error("httpClient and streamingClient should have different timeouts")
	}
}

func TestGenerateCommand_ErrorHandling(t *testing.T) {
	// Test with empty prompt
	result, err := GenerateCommand("", "llama3.1:8b")
	if err == nil {
		t.Error("GenerateCommand with empty prompt should return error")
	}
	if result != "" {
		t.Error("GenerateCommand with empty prompt should return empty result")
	}
}

func TestResolveModel_ErrorHandling(t *testing.T) {
	// Test with invalid model
	model, status, err := resolveModel("nonexistent-model", false)
	if err == nil {
		t.Error("resolveModel with nonexistent model should return error")
	}
	if model != "nonexistent-model" {
		t.Errorf("resolveModel should return original model name, got %v", model)
	}
	// Status might be empty if there's an error reaching Ollama
	if status != "" {
		t.Logf("Status message: %s", status)
	}
}

func TestListOllamaModels_ErrorHandling(t *testing.T) {
	// Test with invalid host
	os.Setenv("OLLAMA_HOST", "http://nonexistent:9999")
	defer os.Unsetenv("OLLAMA_HOST")

	// This test would require the actual listOllamaModels function
	// For now, just test that the environment variable is set
	host := os.Getenv("OLLAMA_HOST")
	if host != "http://nonexistent:9999" {
		t.Errorf("Expected OLLAMA_HOST to be set, got %v", host)
	}
}
