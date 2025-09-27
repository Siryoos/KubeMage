package main

import "strings"

// extractHelmReleaseName attempts to extract release name from helm command
func extractHelmReleaseName(cmd string) string {
	parts := strings.Fields(cmd)
	for i, part := range parts {
		if (part == "install" || part == "upgrade") && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return "release"
}

// RequiresTypedConfirmation checks if command needs "yes" typed confirmation
func (p PreExecPlan) RequiresTypedConfirmation() bool {
	for _, note := range p.Notes {
		if strings.Contains(note, "DANGEROUS") {
			return true
		}
	}
	return false
}