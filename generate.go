package main

import "fmt"

// GetManifestGenerationPrompt constructs a prompt for the LLM to generate a Kubernetes manifest.
func GetManifestGenerationPrompt(kind string, name string) string {
	return fmt.Sprintf("Generate a Kubernetes manifest for a %s named %s.", kind, name)
}

// GetHelmGenerationPrompt constructs a prompt for the LLM to generate a Helm chart.
func GetHelmGenerationPrompt(name string) string {
	return fmt.Sprintf("Generate a Helm chart skeleton for an application named %s.", name)
}