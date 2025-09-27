package main

import (
	"fmt"
	"strings"
)

// DiffMode represents the type of file being edited via diff
type DiffMode string

const (
	DiffModeValues   DiffMode = "values"
	DiffModeManifest DiffMode = "manifest"
)

// DiffPhase tracks the lifecycle of a diff editing session
type DiffPhase int

const (
	DiffPhaseNone DiffPhase = iota
	DiffPhaseAwaiting
	DiffPhasePreview
	DiffPhaseApplied
)

// DiffSession captures state for an in-flight diff-based edit
type DiffSession struct {
	Mode             DiffMode
	FilePath         string
	Instruction      string
	OriginalContent  string
	RedactedContent  string
	Redactions       map[string]string
	phase            DiffPhase
	response         strings.Builder
	ParsedDiff       *UnifiedDiff
	ModifiedContent  string
	modifiedRedacted string
}

// NewDiffSession constructs a diff session for the provided context
func NewDiffSession(mode DiffMode, filePath, instruction string, originalContent string, redaction RedactionResult) *DiffSession {
	return &DiffSession{
		Mode:            mode,
		FilePath:        filePath,
		Instruction:     instruction,
		OriginalContent: originalContent,
		RedactedContent: redaction.Sanitized,
		Redactions:      redaction.Replacements,
		phase:           DiffPhaseAwaiting,
	}
}

// Phase returns the current phase of the diff session
func (ds *DiffSession) Phase() DiffPhase {
	if ds == nil {
		return DiffPhaseNone
	}
	return ds.phase
}

// SetPhase updates the internal phase marker
func (ds *DiffSession) SetPhase(phase DiffPhase) {
	if ds != nil {
		ds.phase = phase
	}
}

// AppendResponse accumulates streamed LLM output
func (ds *DiffSession) AppendResponse(chunk string) {
	if ds != nil {
		ds.response.WriteString(chunk)
	}
}

// RawResponse returns the collected LLM text
func (ds *DiffSession) RawResponse() string {
	if ds == nil {
		return ""
	}
	return ds.response.String()
}

// BuildDiffEditPrompt creates a unified diff prompt tailored to the file type
func BuildDiffEditPrompt(mode DiffMode, filePath, currentContent, instruction string) string {
	var fileDescriptor string
	if mode == DiffModeValues {
		fileDescriptor = "Helm values.yaml configuration"
	} else {
		fileDescriptor = "Kubernetes YAML manifest"
	}

	var sb strings.Builder
	sb.WriteString("You are an expert editor generating a unified diff patch for a single file.\n")
	sb.WriteString("Return ONLY a valid unified diff (patch) with context lines, no explanations or commentary.\n")
	sb.WriteString("Use the exact format produced by `git diff` including the header.\n\n")
	sb.WriteString(fmt.Sprintf("Target file: %s (%s)\n", filePath, fileDescriptor))
	sb.WriteString("Instruction:\n")
	sb.WriteString(instruction)
	sb.WriteString("\n\n")
	sb.WriteString("Current file content:\n")
	sb.WriteString("```yaml\n")
	sb.WriteString(currentContent)
	sb.WriteString("\n```\n\n")
	sb.WriteString("Respond with the diff only. Do not restate the file content outside of the unified diff format.\n")
	return sb.String()
}
