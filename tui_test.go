package main

import (
	"strings"
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
			if got := parseCommand(tc.input); got != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestBuildChatPrompt(t *testing.T) {
	history := []message{
		{sender: user, content: "List pods"},
		{sender: assist, content: "Sure, let me help."},
		{sender: execSender, content: "pod output"},
	}

	prompt := buildChatPrompt(history)

	if !strings.Contains(prompt, "User: List pods") {
		t.Fatalf("prompt missing user context: %q", prompt)
	}

	if !strings.Contains(prompt, "Assistant: Sure, let me help.") {
		t.Fatalf("prompt missing assistant context: %q", prompt)
	}

	if !strings.HasSuffix(prompt, "Assistant:") {
		t.Fatalf("prompt should end by cueing the assistant: %q", prompt)
	}
}
