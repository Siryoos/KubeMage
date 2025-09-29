package llm

import (
	"context"
	"strings"
)

// Client defines the interface for LLM interactions
type Client interface {
	// Complete generates a completion for the given prompt
	Complete(ctx context.Context, prompt string) (string, error)
	
	// CompleteWithSystem generates a completion with a system prompt
	CompleteWithSystem(ctx context.Context, system, prompt string) (string, error)
	
	// Stream generates a streaming completion
	Stream(ctx context.Context, prompt string, handler StreamHandler) error
	
	// StreamWithSystem generates a streaming completion with a system prompt
	StreamWithSystem(ctx context.Context, system, prompt string, handler StreamHandler) error
	
	// IsAvailable checks if the LLM service is available
	IsAvailable(ctx context.Context) bool
	
	// GetModel returns the current model name
	GetModel() string
}

// StreamHandler processes streaming responses
type StreamHandler func(chunk string, done bool) error

// Options configures the LLM client
type Options struct {
	Model       string
	Endpoint    string
	Temperature float64
	MaxTokens   int
}

// MockClient implements Client for testing
type MockClient struct {
	Responses map[string]string
	Model     string
	Available bool
}

// NewMockClient creates a new mock LLM client
func NewMockClient() *MockClient {
	return &MockClient{
		Responses: make(map[string]string),
		Model:     "mock-model",
		Available: true,
	}
}

// Complete generates a mocked completion
func (m *MockClient) Complete(ctx context.Context, prompt string) (string, error) {
	if resp, ok := m.Responses[prompt]; ok {
		return resp, nil
	}
	return "Mock response for: " + prompt, nil
}

// CompleteWithSystem generates a mocked completion with system prompt
func (m *MockClient) CompleteWithSystem(ctx context.Context, system, prompt string) (string, error) {
	key := system + "|" + prompt
	if resp, ok := m.Responses[key]; ok {
		return resp, nil
	}
	return m.Complete(ctx, prompt)
}

// Stream generates a mocked streaming completion
func (m *MockClient) Stream(ctx context.Context, prompt string, handler StreamHandler) error {
	response := "Mock streaming response"
	if resp, ok := m.Responses[prompt]; ok {
		response = resp
	}
	
	// Simulate streaming
	words := strings.Fields(response)
	for _, word := range words {
		if err := handler(word+" ", false); err != nil {
			return err
		}
	}
	return handler("", true)
}

// StreamWithSystem generates a mocked streaming completion with system prompt
func (m *MockClient) StreamWithSystem(ctx context.Context, system, prompt string, handler StreamHandler) error {
	return m.Stream(ctx, prompt, handler)
}

// IsAvailable checks if the mock service is available
func (m *MockClient) IsAvailable(ctx context.Context) bool {
	return m.Available
}

// GetModel returns the mock model name
func (m *MockClient) GetModel() string {
	return m.Model
}