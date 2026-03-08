package orchestrator

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRunner_IntegrationWithRetryAndArtifacts(t *testing.T) {
	tmp := t.TempDir()

	responses := DefaultStubResponses()
	responses[StepPlanner] = []string{
		`{"architecture_overview":"missing many required keys"}`,
		responses[StepPlanner][0],
	}

	provider := &StaticSequenceProvider{Responses: responses}
	runner := Runner{
		PromptBuilder: DefaultPromptBuilder{},
		Provider:      provider,
		Validator:     JSONContractValidator{},
		Artifacts:     FileArtifactStore{BaseDir: tmp},
		MaxRetries:    1,
	}

	result, err := runner.Run(context.Background(), TaskInput{
		RunID:       "run-001",
		Category:    "new-feature",
		Requirement: "Create order endpoint",
	})
	if err != nil {
		t.Fatalf("runner returned error: %v", err)
	}
	if result.Status != "completed" {
		t.Fatalf("expected completed, got %s", result.Status)
	}
	if result.ManualIntervention {
		t.Fatalf("manual intervention should be false")
	}
	if result.Summary.MergeReadiness == "" {
		t.Fatalf("missing final summary")
	}

	for _, step := range append(WorkflowSteps, StepFinalSummary) {
		loc, ok := result.ArtifactLocations[step]
		if !ok {
			t.Fatalf("missing artifact location for step %s", step)
		}
		for _, f := range []string{"prompt.md", "raw_output.json", "validation.json", "metadata.json"} {
			if _, err := os.Stat(filepath.Join(loc, f)); err != nil {
				t.Fatalf("missing artifact file %s for step %s: %v", f, step, err)
			}
		}
	}
}

func TestRunner_RetryThenManualIntervention(t *testing.T) {
	tmp := t.TempDir()

	provider := &StaticSequenceProvider{Responses: map[Step][]string{
		StepResearcher: {`{"problem_summary":"x","requirements":[],"impacted_components":[],"dependencies":[],"edge_cases":[],"risks":[],"open_questions":[],"test_scenarios":[]}`},
		StepPlanner:    {`{"architecture_overview":"only bad"}`, `{"architecture_overview":"still bad"}`},
	}}

	runner := Runner{
		PromptBuilder: DefaultPromptBuilder{},
		Provider:      provider,
		Validator:     JSONContractValidator{},
		Artifacts:     FileArtifactStore{BaseDir: tmp},
		MaxRetries:    1,
	}

	result, err := runner.Run(context.Background(), TaskInput{
		RunID:       "run-fail-001",
		Category:    "bugfix",
		Requirement: "Fix duplicate processing",
	})
	if err != nil {
		t.Fatalf("runner returned error: %v", err)
	}
	if result.Status != "manual_intervention" {
		t.Fatalf("expected manual_intervention, got %s", result.Status)
	}
	if !result.ManualIntervention {
		t.Fatalf("manual intervention should be true")
	}
	if result.FailedStep != StepPlanner {
		t.Fatalf("failed step mismatch: got %s", result.FailedStep)
	}
}
