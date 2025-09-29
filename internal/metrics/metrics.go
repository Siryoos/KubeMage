package metrics

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type SessionMetrics struct {
	Suggestions       int   `json:"suggestions"`
	ValidationsPassed int   `json:"validations_passed"`
	ValidationsFailed int   `json:"validations_failed"`
	Confirmations     int   `json:"confirmations"`
	Corrections       int   `json:"corrections"`
	EditsSuggested    int   `json:"edits_suggested"`
	EditsApplied      int   `json:"edits_applied"`
	Resolutions       int   `json:"resolutions"`
	SafetyBlocks      int   `json:"safety_blocks"`
	TurnsTracked      []int `json:"-"`
	currentTurns      int
	trackingTurns     bool
	SessionStartedAt  time.Time `json:"session_started_at"`
}

type MetricsSnapshot struct {
	TSR float64 `json:"tsr"`
	CAR float64 `json:"car"`
	EAR float64 `json:"ear"`
	MTR float64 `json:"mtr"`
	SVB int     `json:"svb"`

	Suggestions       int `json:"suggestions"`
	ValidationsPassed int `json:"validations_passed"`
	ValidationsFailed int `json:"validations_failed"`
	Confirmations     int `json:"confirmations"`
	Corrections       int `json:"corrections"`
	EditsSuggested    int `json:"edits_suggested"`
	EditsApplied      int `json:"edits_applied"`
	Resolutions       int `json:"resolutions"`
	SafetyBlocks      int `json:"safety_blocks"`
	SessionSeconds    int `json:"session_seconds"`
}

func NewSessionMetrics() *SessionMetrics {
	return &SessionMetrics{SessionStartedAt: time.Now()}
}

func (m *SessionMetrics) RecordSuggestion() {
	m.Suggestions++
	m.currentTurns = 0
	m.trackingTurns = true
}

func (m *SessionMetrics) RecordValidation(passed bool) {
	if passed {
		m.ValidationsPassed++
	} else {
		m.ValidationsFailed++
	}
}

func (m *SessionMetrics) RecordConfirmation() {
	m.Confirmations++
}

func (m *SessionMetrics) RecordCorrection() {
	m.Corrections++
}

func (m *SessionMetrics) RecordEditSuggestion() {
	m.EditsSuggested++
}

func (m *SessionMetrics) RecordEditApplied() {
	m.EditsApplied++
}

func (m *SessionMetrics) RecordResolution() {
	m.Resolutions++
	if m.trackingTurns {
		m.TurnsTracked = append(m.TurnsTracked, m.currentTurns)
	}
	m.trackingTurns = false
	m.currentTurns = 0
}

func (m *SessionMetrics) RecordSafetyBlock() {
	m.SafetyBlocks++
}

func (m *SessionMetrics) RecordTurn() {
	if m.trackingTurns {
		m.currentTurns++
	}
}

func (m *SessionMetrics) TSR() float64 {
	if m.Suggestions == 0 {
		return 0
	}
	return float64(m.Resolutions) / float64(m.Suggestions)
}

func (m *SessionMetrics) CAR() float64 {
	total := m.ValidationsPassed + m.ValidationsFailed
	if total == 0 {
		return 0
	}
	return float64(m.ValidationsPassed) / float64(total)
}

func (m *SessionMetrics) EAR() float64 {
	if m.EditsSuggested == 0 {
		return 0
	}
	return float64(m.EditsApplied) / float64(m.EditsSuggested)
}

func (m *SessionMetrics) MTR() float64 {
	if len(m.TurnsTracked) == 0 {
		return 0
	}
	total := 0
	for _, t := range m.TurnsTracked {
		total += t
	}
	return float64(total) / float64(len(m.TurnsTracked))
}

func (m *SessionMetrics) Snapshot() MetricsSnapshot {
	elapsed := time.Since(m.SessionStartedAt)
	return MetricsSnapshot{
		TSR:               m.TSR(),
		CAR:               m.CAR(),
		EAR:               m.EAR(),
		MTR:               m.MTR(),
		SVB:               m.SafetyBlocks,
		Suggestions:       m.Suggestions,
		ValidationsPassed: m.ValidationsPassed,
		ValidationsFailed: m.ValidationsFailed,
		Confirmations:     m.Confirmations,
		Corrections:       m.Corrections,
		EditsSuggested:    m.EditsSuggested,
		EditsApplied:      m.EditsApplied,
		Resolutions:       m.Resolutions,
		SafetyBlocks:      m.SafetyBlocks,
		SessionSeconds:    int(elapsed.Seconds()),
	}
}

func (m *SessionMetrics) DumpJSON(w io.Writer) {
	snap := m.Snapshot()
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		fmt.Fprintf(w, "error marshaling metrics: %v\n", err)
		return
	}
	fmt.Fprintf(w, "%s\n", data)
}

func (m *SessionMetrics) Dump() {
	m.DumpJSON(os.Stdout)
}

func (m *SessionMetrics) Table() string {
	snap := m.Snapshot()
	rows := []string{
		"┌────────────────────────────┬───────────┐",
		fmt.Sprintf("│ %-26s │ %9.2f │", "Task Success Rate", snap.TSR*100),
		fmt.Sprintf("│ %-26s │ %9.2f │", "Command Accuracy", snap.CAR*100),
		fmt.Sprintf("│ %-26s │ %9.2f │", "Edit Accuracy", snap.EAR*100),
		fmt.Sprintf("│ %-26s │ %9.2f │", "Mean Turns", snap.MTR),
		fmt.Sprintf("│ %-26s │ %9d │", "Safety Blocks", snap.SVB),
		"└────────────────────────────┴───────────┘",
		fmt.Sprintf("Suggestions: %d  Validations: %d/%d  Edits: %d/%d  Resolutions: %d",
			snap.Suggestions, snap.ValidationsPassed, snap.ValidationsFailed, snap.EditsApplied, snap.EditsSuggested, snap.Resolutions),
	}
	return strings.Join(rows, "\n")
}
