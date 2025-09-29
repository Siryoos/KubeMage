package llm

import (
	"context"
)

// OllamaClient implements the Client interface for Ollama
type OllamaClient struct {
	model    string
	endpoint string
}

// NewOllamaClient creates a new Ollama LLM client
func NewOllamaClient(opts Options) *OllamaClient {
	if opts.Model == "" {
		opts.Model = defaultModelName
	}
	if opts.Endpoint == "" {
		opts.Endpoint = ollamaBaseURL()
	}
	
	return &OllamaClient{
		model:    opts.Model,
		endpoint: opts.Endpoint,
	}
}

// Complete generates a completion for the given prompt
func (c *OllamaClient) Complete(ctx context.Context, prompt string) (string, error) {
	return c.CompleteWithSystem(ctx, chatAssistantSystemPrompt, prompt)
}

// CompleteWithSystem generates a completion with a system prompt
func (c *OllamaClient) CompleteWithSystem(ctx context.Context, system, prompt string) (string, error) {
	// Use the existing GenerateCommand function
	// TODO: Refactor to use context properly
	return GenerateCommand(prompt, c.model)
}

// Stream generates a streaming completion
func (c *OllamaClient) Stream(ctx context.Context, prompt string, handler StreamHandler) error {
	return c.StreamWithSystem(ctx, chatAssistantSystemPrompt, prompt, handler)
}

// StreamWithSystem generates a streaming completion with a system prompt
func (c *OllamaClient) StreamWithSystem(ctx context.Context, system, prompt string, handler StreamHandler) error {
	// Use the existing GenerateChatStream function
	ch := make(chan string)
	
	go func() {
		GenerateChatStream(prompt, ch, c.model, system)
	}()
	
	// Handle timeout using context
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case chunk, ok := <-ch:
			if !ok {
				return handler("", true)
			}
			if err := handler(chunk, false); err != nil {
				return err
			}
		}
	}
}

// IsAvailable checks if the Ollama service is available
func (c *OllamaClient) IsAvailable(ctx context.Context) bool {
	// Try to resolve the model to check if Ollama is available
	_, _, err := ResolveModel(c.model, true)
	return err == nil
}

// GetModel returns the current model name
func (c *OllamaClient) GetModel() string {
	return c.model
}

// GenerateCommandOnly generates a kubectl/helm command without explanation
func (c *OllamaClient) GenerateCommandOnly(ctx context.Context, prompt string) (string, error) {
	return GenerateCommand(prompt, c.model)
}