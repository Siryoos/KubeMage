package main

import "strings"

// GenerationType enumerates supported generation workflows
type GenerationType string

const (
	GenerationTypeDeployment GenerationType = "deployment"
	GenerationTypeHelmChart  GenerationType = "helm-chart"
)

// GenerationPhase captures the lifecycle state for a generation session
type GenerationPhase int

const (
	GenerationPhaseNone GenerationPhase = iota
	GenerationPhaseAwaiting
	GenerationPhasePreview
	GenerationPhaseCompleted
)

// GenerationSession holds state related to LLM-backed file generation
type GenerationSession struct {
	Type          GenerationType
	Name          string
	TargetPath    string // file output (for manifests)
	TargetDir     string // directory output (for helm charts)
	DeployOptions *DeploymentOptions
	HelmOptions   *HelmChartOptions
	phase         GenerationPhase
	response      strings.Builder
}

// NewGenerationSession initializes a new generation session context
func NewGenerationSession(genType GenerationType, name, targetPath, targetDir string) *GenerationSession {
	return &GenerationSession{
		Type:       genType,
		Name:       name,
		TargetPath: targetPath,
		TargetDir:  targetDir,
		phase:      GenerationPhaseAwaiting,
	}
}

// Phase returns the current lifecycle phase
func (gs *GenerationSession) Phase() GenerationPhase {
	if gs == nil {
		return GenerationPhaseNone
	}
	return gs.phase
}

// SetPhase updates the session phase
func (gs *GenerationSession) SetPhase(phase GenerationPhase) {
	if gs != nil {
		gs.phase = phase
	}
}

// AppendResponse stores streamed LLM output
func (gs *GenerationSession) AppendResponse(chunk string) {
	if gs != nil {
		gs.response.WriteString(chunk)
	}
}

// RawResponse returns the collected LLM response string
func (gs *GenerationSession) RawResponse() string {
	if gs == nil {
		return ""
	}
	return gs.response.String()
}
