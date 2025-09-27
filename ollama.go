package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// OllamaRequest represents the request payload for the Ollama API.
// See: https://github.com/ollama/ollama/blob/main/docs/api.md#generate-a-completion
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	System string `json:"system"`
	Stream bool   `json:"stream"`
}

// OllamaResponse represents the response payload from the Ollama API.
// In streaming mode, each response is a separate JSON object.
// The `Done` field is true when the stream is complete.
type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// Generate sends a prompt to the Ollama API and returns the response.
func Generate(prompt string) (string, error) {
	systemPrompt := "You are KubeMage, an AI assistant helping with Kubernetes and Helm. You translate user requests into kubectl/helm commands and provide step-by-step explanations. You NEVER execute actions without user confirmation. Favor read-only queries and dry-runs first. Do not suggest destructive actions (like kubectl delete with wildcards) unless explicitly asked, and even then, warn the user. Respond concisely and helpfully. Only output the kubectl/helm command for the following request."

	// Create the request payload
	requestPayload := OllamaRequest{
		Model:  "codellama:7b", // Default model, can be changed
		Prompt: prompt,
		System: systemPrompt,
		Stream: false, // For now, we don't stream the response
	}

	// Marshal the payload to JSON
	payloadBytes, err := json.Marshal(requestPayload)
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %w", err)
	}

	// Make the HTTP request to the Ollama API
	res, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("error making request to Ollama API: %w", err)
	}
	defer res.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	// Unmarshal the response
	var ollamaResponse OllamaResponse
	if err := json.Unmarshal(body, &ollamaResponse); err != nil {
		return "", fmt.Errorf("error unmarshaling response: %w", err)
	}

	return ollamaResponse.Response, nil
}

// GenerateStream sends a prompt to the Ollama API and streams the response.
func GenerateStream(prompt string, ch chan<- string, model string) {
	defer close(ch)

	systemPrompt := "You are KubeMage, an AI assistant helping with Kubernetes and Helm. You translate user requests into kubectl/helm commands and provide step-by-step explanations. You NEVER execute actions without user confirmation. Favor read-only queries and dry-runs first. Do not suggest destructive actions (like kubectl delete with wildcards) unless explicitly asked, and even then, warn the user. Respond concisely and helpfully. Only output the kubectl/helm command for the following request."

	requestPayload := OllamaRequest{
		Model:  model,
		Prompt: prompt,
		System: systemPrompt,
		Stream: true,
	}

	payloadBytes, err := json.Marshal(requestPayload)
	if err != nil {
		ch <- fmt.Sprintf("Error: %v", err)
		return
	}

	res, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		ch <- fmt.Sprintf("Error: %v", err)
		return
	}
	defer res.Body.Close()

	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		var ollamaResponse OllamaResponse
		if err := json.Unmarshal(scanner.Bytes(), &ollamaResponse); err != nil {
			ch <- fmt.Sprintf("Error: %v", err)
			continue
		}
		ch <- ollamaResponse.Response
		if ollamaResponse.Done {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		ch <- fmt.Sprintf("Error: %v", err)
	}
}
