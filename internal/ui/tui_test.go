package ui

import (
	"testing"
)

func TestParseCommand(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "extracts first command block",
			input:    "Here you go:\n```bash\nkubectl get pods\n```\nSome text",
			expected: "kubectl get pods",
		},
		{
			name:     "returns empty without code fences",
			input:    "No command here",
			expected: "",
		},
		{
			name:     "handles multiple code blocks",
			input:    "```bash\nfirst\n```\n```bash\nsecond\n```",
			expected: "first",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := parseCommandFromResponse(tc.input); got != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestBuildChatPrompt(t *testing.T) {
	cfg := &Config{HistoryLength: 10, Truncation: TruncationSettings{Message: 1000}}
	m := &model{config: cfg}
	history := []message{
		{sender: "User", content: "Hello"},
		{sender: "Assistant", content: "Hi there"},
	}
	prompt := m.buildChatPrompt(history)
	expected := "User: Hello\n\nAssistant: Hi there\n\nAssistant:"
	if prompt != expected {
		t.Errorf("expected prompt %q, got %q", expected, prompt)
	}
}
