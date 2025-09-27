# Project Overview

KubeMage is a CLI tool that acts as an intelligent, conversational assistant for DevOps and platform engineers working with Kubernetes and Helm. It allows users to interact with their clusters and charts using natural language directives, which KubeMage translates into precise `kubectl` and `helm` commands. The project is built in Go and uses the Bubble Tea library for its terminal user interface (TUI). It integrates with the Ollama API to leverage large language models for natural language understanding and command generation.

# Building and Running

**Prerequisites:**

*   Go (version 1.18 or higher)
*   Ollama API running on `http://localhost:11434`

**Building the project:**

```sh
# TODO: Add go mod tidy and go build commands once go.mod is created
```

**Running the project:**

```sh
./kubemage
```

# Development Conventions

*   **File Structure:** The project is organized into multiple files, each with a specific responsibility:
    *   `main.go`: The entry point of the application.
    *   `tui.go`: Contains the logic for the terminal user interface.
    *   `ollama.go`: Handles the communication with the Ollama API.
    *   `exec.go`: Responsible for executing shell commands.
*   **Error Handling:** Errors are handled by returning them from functions and checking for them in the calling code.
*   **Concurrency:** The application uses goroutines and channels to handle concurrent operations, such as streaming responses from the Ollama API and executing commands.
