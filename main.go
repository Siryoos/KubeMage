package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "KubeMage: LLM-powered assistant for Kubernetes and Helm.\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Usage:\n  kubemage [flags] [query]\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Flags:\n")
		flag.PrintDefaults()
	}

	queryFlag := flag.String("query", "", "Translate a natural language request into a kubectl/helm command and print it.")
	modelFlag := flag.String("model", defaultModelName, "Ollama model to use.")
	metricsFlag := flag.Bool("metrics", false, "Dump metrics at the end.")

	flag.Parse()

	query := strings.TrimSpace(*queryFlag)
	if query == "" && flag.NArg() > 0 {
		query = strings.TrimSpace(strings.Join(flag.Args(), " "))
	}

	modelName := strings.TrimSpace(*modelFlag)
	modelForUse := modelName

	if query != "" {
		resolvedModel, statusMsg, err := resolveModel(modelName, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error preparing model: %v\n", err)
			os.Exit(1)
		}
		modelForUse = resolvedModel
		if strings.TrimSpace(statusMsg) != "" {
			fmt.Fprintln(os.Stderr, statusMsg)
		}

		command, err := GenerateCommand(query, modelForUse)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error generating command: %v\n", err)
			os.Exit(1)
		}
		command = strings.TrimSpace(command)
		if command == "" {
			fmt.Fprintln(os.Stderr, "no command produced by model")
			os.Exit(1)
		}
		fmt.Println(command)
		return
	}

	cfg, err := LoadConfig()
	if err != nil {
		cfg = DefaultConfig()
	}
	SetActiveConfig(cfg)

	if host := strings.TrimSpace(cfg.OllamaHost); host != "" {
		_ = os.Setenv("OLLAMA_HOST", host)
	}

	if modelName == "" {
		modelName = cfg.Models.Chat
	}

	m := InitialModel(modelName, cfg, *metricsFlag)
	p := tea.NewProgram(m)
	m.program = p

	if err := p.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Alas, there's been an error: %v\n", err)
		os.Exit(1)
	}

	if *metricsFlag && !m.metricsFlushed {
		m.metrics.DumpJSON(os.Stdout)
		m.metricsFlushed = true
	}
}
