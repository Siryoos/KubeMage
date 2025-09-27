package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// OllamaRequest represents the request payload for the Ollama API.
// See: https://github.com/ollama/ollama/blob/main/docs/api.md#generate-a-completion
type OllamaRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	System  string                 `json:"system"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// OllamaResponse represents the response payload from the Ollama API.
// In streaming mode, each response is a separate JSON object.
// The `Done` field is true when the stream is complete.
type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

type tagList struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

const (
	defaultModelName          = "deepseek-r1:8b"
	defaultOllamaEndpoint     = "http://localhost:11434"
	commandOnlySystemPrompt   = "You are KubeMage, an AI assistant that translates natural language into precise kubectl or helm commands. Always respond with a single command string that can be run as-is. Do not include explanations, markdown, backticks, or additional text. Favor read-only or --dry-run variations when the user intent is ambiguous."
	chatAssistantSystemPrompt = "You are KubeMage, an AI assistant helping with Kubernetes and Helm. Translate user intent into safe kubectl/helm guidance. Answer with short explanations tailored to the cluster context, then conclude with a fenced ```bash code block containing exactly one command that fulfills the request (prefer read-only or --dry-run first when risky). Warn the user about destructive actions and never assume consent."
	agentSystemPrompt         = "You are an agent that can use tools to answer questions. You can use the following tools:\n- `kubectl get ...`\n- `kubectl describe ...`\n- `kubectl logs ...`\n- `kubectl events ...`\n\nTo use a tool, you must respond with an `Action:` block, for example:\n```\nAction: kubectl get pods\n```\n\nI will then execute the tool and provide you with an `Observation:` block containing the output.\n\nWhen you have enough information to answer the user's question, you must respond with a `Final:` block containing your final answer."
)

var (
	httpClient      = &http.Client{Timeout: 10 * time.Second}
	streamingClient = &http.Client{Timeout: 0} // No timeout for streaming
)

// GenerateCommand returns a single kubectl/helm command for one-shot CLI usage.
func GenerateCommand(prompt, model string) (string, error) {
	modelName := model
	if modelName == "" {
		modelName = defaultModelName
	}

	if ctxSum, err := BuildContextSummary(); err == nil && ctxSum != nil {
		// Prepend a short context banner. Keep it tiny to save tokens.
		prompt = fmt.Sprintf("[CTX] %s\n\n%s", ctxSum.RenderedOneLiner, prompt)
	}
	prompt = RedactText(prompt)

	res, err := postOllama(prompt, commandOnlySystemPrompt, modelName, false)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	var ollamaResponse OllamaResponse
	if err := json.Unmarshal(body, &ollamaResponse); err != nil {
		return "", fmt.Errorf("error unmarshaling response: %w", err)
	}

	return strings.TrimSpace(ollamaResponse.Response), nil
}

// GenerateChatStream sends a prompt to the Ollama API and streams the response.
func GenerateChatStream(prompt string, ch chan<- string, model string, systemPrompt string) {
	defer close(ch)

	modelName := model
	if modelName == "" {
		modelName = defaultModelName
	}

	if ctxSum, err := BuildContextSummary(); err == nil && ctxSum != nil {
		// Prepend a short context banner. Keep it tiny to save tokens.
		prompt = fmt.Sprintf("[CTX] %s\n\n%s", ctxSum.RenderedOneLiner, prompt)
	}
	prompt = RedactText(prompt)

	res, err := postOllama(prompt, systemPrompt, modelName, true)
	if err != nil {
		ch <- fmt.Sprintf("Error contacting Ollama: %v\nStart 'ollama serve' locally or set OLLAMA_HOST to a reachable instance.", err)
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

func postOllama(prompt, systemPrompt, model string, stream bool) (*http.Response, error) {
	requestPayload := OllamaRequest{
		Model:  model,
		Prompt: prompt,
		System: systemPrompt,
		Stream: stream,
	}

	if cfg := ActiveConfig(); cfg != nil {
		options := make(map[string]interface{})
		if cfg.NumCtx > 0 {
			options["num_ctx"] = cfg.NumCtx
		}
		if ka := strings.TrimSpace(cfg.KeepAlive); ka != "" {
			options["keep_alive"] = ka
		}
		if len(options) > 0 {
			requestPayload.Options = options
		}
	}

	payloadBytes, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling JSON: %w", err)
	}

	endpoint := ollamaEndpoint()

	client := httpClient
	if stream {
		client = streamingClient
	}
	res, err := client.Post(endpoint, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("error making request to Ollama API: %w", err)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("ollama returned status %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}

	return res, nil
}

func ollamaEndpoint() string { return ollamaBaseURL() + "/api/generate" }

func ollamaBaseURL() string {
	host := strings.TrimSpace(os.Getenv("OLLAMA_HOST"))
	if host == "" {
		host = defaultOllamaEndpoint
	}
	return strings.TrimSuffix(host, "/")
}

func ListModels() ([]string, error) {
	base := ollamaBaseURL()
	resp, err := httpClient.Get(base + "/api/tags")
	if err != nil {
		return nil, fmt.Errorf("failed to reach Ollama at %s: %w", base, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama at %s returned status %d: %s", base, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var tags tagList
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, fmt.Errorf("failed to decode Ollama tags: %w", err)
	}

	var modelNames []string
	for _, m := range tags.Models {
		modelNames = append(modelNames, m.Name)
	}
	return modelNames, nil
}

func resolveModel(preferred string, allowFallback bool) (string, string, error) {
	base := ollamaBaseURL()
	resp, err := httpClient.Get(base + "/api/tags")
	if err != nil {
		return preferred, "", fmt.Errorf("failed to reach Ollama at %s: %w", base, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return preferred, "", fmt.Errorf("Ollama at %s returned status %d: %s", base, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var tags tagList
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return preferred, "", fmt.Errorf("failed to decode Ollama tags: %w", err)
	}

	if len(tags.Models) == 0 {
		msg := fmt.Sprintf("Ollama at %s has no models installed. Pull one with 'ollama pull %s'.", base, preferred)
		if allowFallback {
			return preferred, msg, nil
		}
		return preferred, "", errors.New(msg)
	}

	if preferred == "" {
		preferred = defaultModelName
	}

	for _, m := range tags.Models {
		if m.Name == preferred {
			return preferred, fmt.Sprintf("Connected to Ollama at %s using model %s.", base, preferred), nil
		}
	}

	if !allowFallback {
		return preferred, "", fmt.Errorf("model %s is not available on Ollama host %s. Install it with 'ollama pull %s'.", preferred, base, preferred)
	}

	fallback := tags.Models[0].Name
	status := fmt.Sprintf("Model %s not found. Using %s instead. Run '/model <name>' to switch.", preferred, fallback)
	return fallback, status, nil
}
