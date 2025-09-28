package main

import (
	"testing"
)

func TestBuildContextSummary(t *testing.T) {
	// Test that BuildContextSummary returns a result even when kubectl might not be available
	summary, err := BuildContextSummary()

	// Should not fail completely - even if kubectl fails, it should return something
	if summary == nil {
		t.Fatal("BuildContextSummary returned nil summary")
	}

	if err != nil {
		t.Logf("BuildContextSummary had error (expected if kubectl not available): %v", err)
	}

	// Verify required fields are set (even if to defaults)
	if summary.Context == "" {
		t.Log("Context is empty (expected if kubectl not available)")
	}
	if summary.Namespace == "" {
		t.Log("Namespace is empty (expected if kubectl not available)")
	}

	// Verify maps are initialized
	if summary.PodPhaseCounts == nil {
		t.Error("PodPhaseCounts map should be initialized")
	}
	if summary.PodProblemCounts == nil {
		t.Error("PodProblemCounts map should be initialized")
	}

	// Verify RenderedOneLiner is generated
	if summary.RenderedOneLiner == "" {
		t.Error("RenderedOneLiner should not be empty")
	}

	t.Logf("Generated context summary: %s", summary.RenderedOneLiner)
}

func TestGetCurrentContext(t *testing.T) {
	// Test that GetCurrentContext handles errors gracefully
	ctx, err := GetCurrentContext()
	if err != nil {
		t.Logf("GetCurrentContext failed (expected if kubectl not available): %v", err)
	} else {
		t.Logf("Current context: %s", ctx)
	}
}

func TestGetCurrentNamespace(t *testing.T) {
	// Test that GetCurrentNamespace handles errors gracefully
	ns, err := GetCurrentNamespace()
	if err != nil {
		t.Logf("GetCurrentNamespace failed (expected if kubectl not available): %v", err)
	} else {
		t.Logf("Current namespace: %s", ns)
	}
}
