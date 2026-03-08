package orchestrator

import "context"

type PromptBuilder interface {
	Build(step Step, ctx PromptContext) (string, error)
}

type ProviderAdapter interface {
	Run(ctx context.Context, agent string, prompt string) (string, error)
}

type ContractValidator interface {
	Validate(step Step, rawOutput string) ValidationResult
}

type ArtifactStore interface {
	Save(runID string, step Step, artifacts StepArtifacts) (string, error)
}
