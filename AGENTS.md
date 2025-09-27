# Repository Guidelines

## Project Structure & Module Organization
- Single Go module (`github.com/siryoos/kubemage`) in the repo root; `main.go` wires together the TUI runtime.
- `tui.go` owns Bubble Tea state, `ollama.go` handles Ollama streaming, and `exec.go` wraps shell execution used for `/model` and generated commands.
- The `kubemage` binary in the root is a build artifact; regenerate it locally rather than editing or committing new binaries.

## Build, Test, and Development Commands
- `go mod tidy` updates `go.mod`/`go.sum` whenever dependencies change.
- `go build -o kubemage ./...` produces the CLI; run `go run .` for a quick local session.
- `GO111MODULE=on go test ./...` executes unit tests; add `-v` while iterating to surface flaky behavior.
- The TUI depends on an Ollama server at `http://localhost:11434`; export `OLLAMA_HOST` if you target a different endpoint.

## Coding Style & Naming Conventions
- Always run `gofmt` (or `go fmt ./...`) before committing; default tab indentation and CamelCase exports keep the codebase idiomatic.
- Prefer short, specific function names tied to Bubble Tea events (e.g., `generateStreamCmd`); keep command constants like `user`, `assist`, `execSender` lowercase.
- Centralize styling in `defaultStyles`; add new visual variants there rather than scattering lipgloss setup.
- Document non-obvious concurrency (goroutines, channels) with brief comments to aid future agents.

## Testing Guidelines
- Add table-driven tests for pure functions such as `parseCommand`; guard HTTP integrations with `httptest.Server` to avoid hitting live Ollama instances.
- Mock Bubble Tea messages where possible; assert that generated commands and state transitions match expectations.
- Aim for coverage on any new parsing or command-safety logic; include benchmark stubs when optimizing streaming performance.

## Commit & Pull Request Guidelines
- Follow Conventional Commits (`feat(tui): describe change`) as seen in recent history; keep the subject ≤72 characters and elaborate in the body when needed.
- Reference linked issues, summarize manual TUI verification (`go run .`, sample prompt/response), and note any Ollama model requirements in the PR description.
- Before requesting review, ensure binaries are rebuilt locally, tests pass, and docs (this file, `GEMINI.md`, README) reflect user-facing changes.

## Ollama & Security Notes
- Treat generated shell commands as untrusted; keep `/model` changes scoped to known-safe models and surface warnings for destructive operations.
- Never embed API keys or cluster credentials in source or commits; rely on local environment variables and `.gitignore` for secrets.
