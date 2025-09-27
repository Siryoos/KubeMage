# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

KubeMage is an LLM-powered terminal UI application for Kubernetes and Helm management. It provides an interactive chat interface that can generate kubectl/helm commands, execute them, and provide diagnostics.

## Architecture

### Core Components

- **main.go**: Entry point with dual modes:
  - CLI mode: Direct command generation via `--query` flag
  - TUI mode: Interactive chat interface (default)

- **tui.go**: Bubble Tea-based terminal UI with chat interface
  - Uses Charm libraries (bubbles, bubbletea, lipgloss) for UI components
  - Implements streaming responses from Ollama
  - Supports slash commands (e.g., `/model`, `/exec`, `/context`)

- **ollama.go**: Ollama API integration for LLM interactions
  - Handles streaming responses via HTTP
  - Model resolution and validation
  - System prompts for Kubernetes/Helm context

- **config.go**: YAML-based configuration management
  - Default config path: `config.yaml`
  - Configurable model, tokens, history length

### Key Features

- **Command Generation**: Natural language to kubectl/helm commands
- **Command Execution**: Built-in execution with real-time output streaming
- **Context Awareness**: Automatic kubectl context and namespace detection
- **Diagnostics**: Structured diagnostic plans for troubleshooting
- **Model Switching**: Runtime model selection via `/model` command

### Data Flow

1. User input → TUI (tui.go)
2. LLM request → Ollama API (ollama.go)
3. Command execution → exec.go
4. Context gathering → context.go
5. Diagnostics → diagnostics.go

## Development Commands

```bash
# Build the application
make build

# Run tests
make test

# Run the application (builds first)
make run

# Direct execution with query
go run . --query "list all pods"

# Run in TUI mode
go run .
```

## Configuration

The application uses `config.yaml` for configuration:
- `model`: Ollama model name (default: "codellama:7b")
- `max_tokens`: Maximum response tokens
- `truncation_size`: Output truncation limit
- `chat_history_length`: Chat history retention

## Testing

- Unit tests: `*_test.go` files
- Run with: `go test ./...`
- Current test coverage includes TUI and validator components

## Dependencies

Built with Go 1.24+ and uses:
- Charm stack (bubbletea, bubbles, lipgloss) for TUI
- YAML parsing (gopkg.in/yaml.v3)
- Standard library for HTTP and execution

## Key Files

- `exec.go`: Command execution with streaming output
- `context.go`: Kubernetes context and cluster state gathering
- `diagnostics.go`: Diagnostic plan generation and execution
- `generate.go`: Prompt generation utilities
- `metrics.go`: Usage metrics collection
- `validator.go`: Input validation and sanitization