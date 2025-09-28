# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

KubeMage is an LLM-powered terminal UI application for Kubernetes and Helm management with advanced agentic capabilities. It provides an interactive chat interface that can generate kubectl/helm commands, execute them with comprehensive safety controls, perform diagnostics, and support diff-first editing workflows.

## Architecture

### Core Safety & Agent Framework

KubeMage implements a **ReAct-lite agent loop** with strict safety controls:

- **Safety-First Execution**: All mutating operations require dry-run validation and explicit confirmation
- **Whitelisted Agent Actions**: ReAct agent can only execute read-only kubectl/helm commands
- **Validator Gate**: `validator.go` implements `PreExecPlan` with danger level analysis and safety checks
- **Context Scrying**: Automatic injection of cluster state into every LLM prompt via `[CTX]` banners
- **Self-Correction**: Failed commands trigger automatic correction prompts to the LLM

### Component Architecture

**Core Application Flow:**
- **main.go**: CLI/TUI mode dispatcher with `--metrics` flag support
- **tui.go**: Bubble Tea UI with multi-pane layout, streaming, and comprehensive slash command support
- **ollama.go**: Ollama API client with model resolution and context injection

**Safety & Validation Pipeline:**
- **validator.go**: `PreExecPlan` generation with danger level analysis (low/medium/high/critical)
- **exec.go**: Command execution with timeout, streaming output, and validation failure handling
- **context.go**: `KubeContextSummary` generation for token-efficient cluster state

**Agent & Diagnostics:**
- **diagnostics.go**: ReAct-lite agent implementation with whitelisted command execution
- **metrics.go**: Comprehensive session metrics (TSR, CAR, EAR, MTR, SVB)

**Diff-First Editing System:**
- **workspace.go**: Workspace indexing for charts, templates, values, and manifests
- **generate.go**: Content generation with validation pipelines
- **editing.go**: Unified diff application with backup management
- **validation.go**: Multi-stage validation (helm lint, kubectl dry-run, etc.)

### Key Data Structures

**Safety Planning:**
```go
type PreExecPlan struct {
    Original            string
    Checks              []PreviewCheck  // helm lint, kubectl validation
    FirstRunCommand     string         // --dry-run variant
    RequireSecondConfirm bool          // after dry-run
    RequireTypedConfirm  bool          // "yes" confirmation for dangerous ops
    DangerLevel         string         // "low", "medium", "high", "critical"
}
```

**Agent Loop:**
```go
type ReActSession struct {
    MaxSteps    int            // ≤5 steps maximum
    Steps       []ReActStep    // Action → Observation history
    Completed   bool           // Final: statement received
}
```

**Context Injection:**
```go
type KubeContextSummary struct {
    Context          string         // kubectl context
    Namespace        string         // active namespace
    PodPhaseCounts   map[string]int // Running=X, Pending=Y
    PodProblemCounts map[string]int // CrashLoopBackOff=N, etc.
    RenderedOneLiner string         // [CTX] compact prompt prefix
}
```

## Development Commands

```bash
# Build the application
make build

# Run all tests
make test

# Run single test file
go test -v ./validator_test.go ./validator.go

# Run specific test function
go test -run TestBuildPreExecPlan -v

# Run the application (builds first)
make run

# CLI mode with metrics
go run . --query "list failing pods" --metrics

# TUI mode with specific model
go run . --model llama3.1:8b
```

## Testing Strategy

### Unit Test Coverage
- `validator_test.go`: PreExecPlan generation and danger detection
- `validator_safety_test.go`: Safety mechanism validation
- `context_test.go`: Kubernetes context summarization
- `diagnostics_test.go`: ReAct agent workflow and whitelisting
- `command_parsers_test.go`: Helm command parsing
- `tui_test.go`: UI component testing

### Test Patterns
```bash
# Run tests with coverage
go test -cover ./...

# Run safety-critical tests
go test -run TestSafety -v

# Run agent tests
go test -run TestReAct -v

# Run with race detection
go test -race ./...
```

## Configuration

The application uses `config.yaml` with dual model support:
```yaml
models:
  chat: "llama3.1:8b"          # Interactive chat
  generation: "llama3.1:13b"   # Diff generation, corrections
num_ctx: 4096                  # Context window
keep_alive: "5m"               # Model persistence
truncation:
  message: 1200                # UI message limit
history_length: 10             # Chat history retention
ollama_host: "http://localhost:11434"
```

## Safety Guarantees

1. **Dry-Run First**: All `kubectl apply|create|patch|delete` and `helm install|upgrade` operations automatically append `--dry-run=client` for first execution
2. **Second Confirmation**: After successful dry-run, requires Ctrl+E again to execute the real command
3. **Typed Confirmation**: Dangerous operations (bulk deletes, wildcards, force operations) require typing "yes"
4. **Whitelisted Agent**: ReAct agent can only execute read-only commands: `kubectl get|describe|logs|top|api-resources|version|explain` and `helm lint|template|version|show|get`
5. **Context Injection**: Every LLM prompt includes `[CTX]` banner with current cluster state
6. **Secret Redaction**: Automatic redaction of tokens, passwords, and base64 blobs before sending to LLM

## Slash Commands Reference

**Core Commands:**
- `/help` - Toggle inline help
- `/model list` - List available Ollama models
- `/model set chat <name>` - Switch chat model
- `/model set generation <name>` - Switch diff/generation model
- `/ctx` - Show current cluster context
- `/ns set <namespace>` - Switch active namespace
- `/metrics` - Display session metrics
- `/resolve [note]` - Mark current task as resolved

**Agent & Diagnostics:**
- `/agent` - Toggle ReAct agent mode
- `/diag-pod <pod-name>` - Run comprehensive pod diagnostics

**Diff-First Editing:**
- `/edit-yaml <path> <instruction>` - Generate unified diff for manifest
- `/edit-values <path> <instruction>` - Generate diff for Helm values
- `/gen-deploy <name> --image <img>` - Generate deployment manifest
- `/gen-helm <chart> [flags]` - Generate Helm chart skeleton
- `/cancel` - Cancel pending diff/generation operations

## Metrics Tracking

KubeMage tracks comprehensive session metrics accessible via `/metrics` or `--metrics` flag:

- **TSR (Task Success Rate)**: Resolutions / Suggestions
- **CAR (Command Accuracy Rate)**: Validations Passed / Total Validations
- **EAR (Edit Accuracy Rate)**: Edits Applied / Edits Suggested
- **MTR (Mean Turns to Resolution)**: Average conversation turns per task
- **SVB (Safety Violations Blocked)**: Dangerous operations prevented

## ReAct Agent Protocol

The agent follows a structured loop for autonomous diagnostics:

```
Action:
kubectl describe pod failing-pod-xyz

Observation:
Pod status shows ImagePullBackOff error...

Action:
kubectl get events --field-selector involvedObject.name=failing-pod-xyz

Observation:
Failed to pull image "nginx:invalid-tag": image not found

Final:
The pod is failing due to an invalid image tag "nginx:invalid-tag".
The image doesn't exist in the registry.

Next steps:
```sh
kubectl patch deployment myapp -p '{"spec":{"template":{"spec":{"containers":[{"name":"nginx","image":"nginx:latest"}]}}}}'
```

**Safety Controls:**
- Maximum 5 steps per session
- Only whitelisted read-only commands allowed
- Automatic termination on non-whitelisted command attempts

## Context Scrying

Every LLM prompt includes a compact context banner:
```
[CTX] ctx=prod-cluster ns=default pods:{Running=7,Pending=2} podProblems:{CrashLoopBackOff=1} deploy=3 svc=9
```

This provides:
- Current kubectl context and namespace
- Pod phase distribution
- Problem pod counts (CrashLoopBackOff, ImagePullBackOff, OOMKilled)
- Resource counts (deployments, services)

## Dependencies

**Core Stack:**
- Go 1.24+ with toolchain go1.24.7
- Charm TUI stack: bubbletea, bubbles, lipgloss
- YAML parsing: gopkg.in/yaml.v3
- Standard library for HTTP, execution, and crypto

**External Dependencies:**
- `kubectl` CLI (required)
- `helm` CLI (for Helm operations)
- Ollama server (local or remote)
- Active Kubernetes cluster context