# KubeMage 🧙‍♂️

**KubeMage** is a pilot-ready, safety-first, "local Claude/Gemini" for Kubernetes & Helm operations. Built with Go + Bubble Tea TUI and powered by local Ollama models, it provides an agentic AI assistant with comprehensive safety controls for Kubernetes cluster management.

## 🎯 Key Features

### 🛡️ Safety-First Architecture
- **Enforced Validator Gate**: All mutating operations require dry-run validation first
- **Second Confirmation**: Real commands only execute after successful preview and explicit confirmation
- **Danger Detection**: Automatic detection of dangerous patterns with typed confirmation requirements
- **Context Scrying**: Automatic cluster state injection into every LLM prompt for awareness

### 🤖 Agentic Capabilities
- **ReAct-lite Agent Loop**: Autonomous diagnostics with read-only command whitelisting (≤5 steps)
- **Self-Correction**: Failed commands trigger automatic LLM correction prompts
- **Diff-First Editing**: All file modifications shown as unified diffs before application
- **Comprehensive Metrics**: Track TSR, CAR, EAR, MTR, and SVB metrics

### 🚨 Security Hygiene
- **Secret Redaction**: Automatic redaction of tokens, passwords, and secrets from prompts
- **Command Whitelisting**: Agent loop limited to safe, read-only operations
- **Secure Execution**: Preference for `exec.Command()` over shell execution

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- kubectl CLI configured with cluster access
- helm CLI (for Helm operations)
- Ollama server running locally or remote

### Installation
```bash
git clone https://github.com/siryoos/kubemage
cd kubemage
go build -o kubemage .
```

### Basic Usage

**CLI Mode** (one-shot command generation):
```bash
# Generate a kubectl command
./kubemage "list all failing pods"
./kubemage --query "show deployment logs for nginx"
./kubemage --model llama3.1:8b "create a service for my app"
```

**TUI Mode** (interactive chat):
```bash
# Start interactive mode
./kubemage

# With metrics tracking
./kubemage --metrics
```

## 🎮 TUI Interface

### Layout Modes
- **Three-Pane Layout**: Chat | Preview/Diff | Output/Logs (F2 to cycle)
- **Vertical Split**: Side-by-side chat and preview
- **Horizontal Split**: Top/bottom layout
- **Chat Only**: Full-screen chat

### Key Bindings
- **`Ctrl+E`**: Validate/dry-run command; second `Ctrl+E` applies real command
- **`Ctrl+P`**: Open command palette
- **`F2`**: Cycle layout modes
- **`Esc`**: Cancel current operation
- **`/`**: Search in chat history
- **`]`**: Expand truncated output
- **`c`**: Copy selected block
- **`n/p`**: Navigate diff hunks (in diff view)

### Status Footer
```
ctx:prod-cluster ns:kube-system model:llama3.1:8b time:14:30:45
```
*Red accent if production-like namespace detected*

## 📋 Slash Commands

### Core Commands
- **`/help`** - Toggle inline help
- **`/ctx`** - Show current cluster context
- **`/ns set <namespace>`** - Switch active namespace
- **`/metrics`** - Display session metrics
- **`/resolve [note]`** - Mark current task as resolved

### Model Management
- **`/model list`** - List available Ollama models
- **`/model set chat <name>`** - Switch chat model
- **`/model set generation <name>`** - Switch diff/generation model

### Diagnostics & Agent
- **`/agent`** - Toggle ReAct agent mode
- **`/diag-pod <pod-name>`** - Run comprehensive pod diagnostics

### Diff-First Editing
- **`/edit-yaml <path> <instruction>`** - Generate unified diff for manifest
- **`/edit-values <path> <instruction>`** - Generate diff for Helm values
- **`/gen-deploy <name> --image <img>`** - Generate deployment manifest
- **`/gen-helm <chart> [flags]`** - Generate Helm chart skeleton
- **`/cancel`** - Cancel pending diff/generation operations

## 🛡️ Safety Guarantees

### 1. Enforced Validator & Second Confirm
- **Mutating kubectl** (`apply|create|patch|delete`) → runs with `--dry-run=client` first
- **Helm operations** (`install|upgrade`) → runs `helm lint` + `helm template --dry-run` before apply
- **Second confirmation** required after successful previews to execute real command

### 2. Dangerous Pattern Detection
- **Bulk deletes**: `delete all --all`, `--all-namespaces` → RED banner + require typing "yes"
- **Wildcard selectors**: `*` patterns → Critical warning + typed confirmation
- **Force operations**: `--force`, `--grace-period=0` → High risk warning
- **Cluster resources**: nodes, namespaces, PVs → Critical protection

### 3. Read-Only Agent Whitelist
ReAct agent can only execute:
- **kubectl**: `get|describe|logs|top|api-resources|version|explain`
- **helm**: `lint|template|version|show|get`
- Mutating operations rejected with safety error

## 🔍 Context Scrying

Every LLM prompt includes a compact context banner:
```
[CTX] ctx=prod-cluster ns=default pods:{Running=7,Pending=2} podProblems:{CrashLoopBackOff=1} deploy=3 svc=9
```

This provides:
- Current kubectl context and namespace
- Pod phase distribution
- Problem pod counts (CrashLoopBackOff, ImagePullBackOff, OOMKilled)
- Resource counts (deployments, services)

## 🤖 ReAct Agent Protocol

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

## 📊 Metrics Tracking

KubeMage tracks comprehensive session metrics:

### Core Metrics
- **TSR (Task Success Rate)**: Resolutions / Suggestions
- **CAR (Command Accuracy Rate)**: Validations Passed / Total Validations
- **EAR (Edit Accuracy Rate)**: Edits Applied / Edits Suggested
- **MTR (Mean Turns to Resolution)**: Average conversation turns per task
- **SVB (Safety Violations Blocked)**: Dangerous operations prevented

### Viewing Metrics
```bash
# In-TUI display
/metrics

# JSON export on exit
./kubemage --metrics
```

### Sample Output
```
┌────────────────────────────┬───────────┐
│ Task Success Rate          │     85.00 │
│ Command Accuracy           │     92.30 │
│ Edit Accuracy              │     78.50 │
│ Mean Turns                 │      3.20 │
│ Safety Blocks              │         5 │
└────────────────────────────┴───────────┘
Suggestions: 12  Validations: 24/26  Edits: 11/14  Resolutions: 10
```

## ⚙️ Configuration

Configuration via `config.yaml`:
```yaml
models:
  chat: "llama3.1:8b"          # Interactive chat model
  generation: "llama3.1:13b"   # Diff generation, corrections
num_ctx: 4096                  # Context window size
keep_alive: "5m"               # Model persistence
truncation:
  message: 1200                # UI message truncation
history_length: 10             # Chat history retention
theme: "default"               # UI theme
ollama_host: "http://localhost:11434"
```

## 🔧 Diff-First Editing

All file modifications use a diff-first workflow:

1. **Generate**: LLM creates unified diff
2. **Preview**: Show diff with syntax highlighting
3. **Validate**: Run appropriate validators (kubectl dry-run, helm lint)
4. **Confirm**: User approves before application
5. **Apply**: Patch applied with backup creation

### Example Workflow
```bash
/edit-values ./helm-chart/values.yaml "increase nginx replicas to 3"
```
→ Shows unified diff
→ Runs `helm template` validation
→ Awaits confirmation
→ Applies patch with backup

## 🔒 Security Features

### Secret Redaction
Automatic detection and redaction of:
- JWT tokens
- Bearer tokens
- Key-value secrets (`password=`, `token=`)
- Base64 encoded blobs
- secretKeyRef YAML blocks

### Secure Command Execution
- Preference for `exec.Command(name, args...)` over shell execution
- Argument sanitization and validation
- Timeout controls for all operations

## 🧪 Testing

```bash
# Run all tests
go test -v ./...

# Test specific components
go test -run TestBuildPreExecPlan -v
go test -run TestReActSession -v
go test -run TestSafety -v

# Run with race detection
go test -race ./...
```

### Test Coverage
- **Validator Safety**: PreExecPlan generation and danger detection
- **Agent Whitelisting**: ReAct command validation
- **Context Summarization**: Kubernetes state parsing
- **Command Parsing**: Helm/kubectl command analysis
- **Security**: Redaction and sanitization

## 🏗️ Architecture

### Core Components
- **main.go**: CLI/TUI dispatcher with metrics support
- **tui.go**: Bubble Tea UI with multi-pane layout and streaming
- **validator.go**: Safety validation with PreExecPlan generation
- **context.go**: Kubernetes context summarization and injection
- **diagnostics.go**: ReAct agent implementation with whitelisting
- **ollama.go**: LLM integration with context injection
- **exec.go**: Secure command execution with streaming
- **diff.go**: Unified diff parsing and rendering
- **redact.go**: Security hygiene and secret redaction
- **metrics.go**: Comprehensive session metrics tracking

### Data Flow
1. **User Input** → Slash command or natural language
2. **Context Injection** → Add `[CTX]` banner to prompt
3. **LLM Processing** → Generate command or response
4. **Safety Validation** → Build PreExecPlan with checks
5. **Preview Execution** → Run dry-run/lint validations
6. **User Confirmation** → Second confirmation for real execution
7. **Metrics Tracking** → Update TSR/CAR/EAR/MTR/SVB counters

## 🤝 Contributing

1. Fork the repository
2. Create feature branch: `git checkout -b feature/amazing-feature`
3. Run tests: `go test -v ./...`
4. Commit changes: `git commit -m 'Add amazing feature'`
5. Push branch: `git push origin feature/amazing-feature`
6. Open Pull Request

## 📝 License

MIT License - see LICENSE file for details.

## 🙏 Acknowledgments

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) for the excellent TUI framework
- [Ollama](https://ollama.ai/) for local LLM serving
- The Kubernetes and Helm communities for their robust CLI tools

---

**⚠️  Production Warning**: KubeMage provides safety controls but always verify commands before execution in production environments. The AI assistant is a tool to aid decision-making, not replace human judgment.