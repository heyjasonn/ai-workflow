package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Runner struct {
	PromptBuilder PromptBuilder
	Provider      ProviderAdapter
	Validator     ContractValidator
	Artifacts     ArtifactStore
	MaxRetries    int
}

func (r Runner) Run(ctx context.Context, task TaskInput) (RunResult, error) {
	if r.PromptBuilder == nil || r.Provider == nil || r.Validator == nil || r.Artifacts == nil {
		return RunResult{}, fmt.Errorf("runner dependencies are not fully configured")
	}
	if task.RunID == "" {
		task.RunID = fmt.Sprintf("run-%d", time.Now().UTC().Unix())
	}
	if task.Category == "" {
		task.Category = "new-feature"
	}

	outputs := make(map[Step]json.RawMessage)
	locations := make(map[Step]string)
	retries := r.MaxRetries
	if retries < 0 {
		retries = 0
	}

	for _, step := range WorkflowSteps {
		attemptErrors := []string{}
		stepCompleted := false

		for attempt := 0; attempt <= retries; attempt++ {
			prompt, err := r.PromptBuilder.Build(step, PromptContext{
				Task:        task,
				Previous:    outputs,
				RetryErrors: attemptErrors,
			})
			if err != nil {
				return RunResult{}, err
			}

			rawOutput, err := r.Provider.Run(ctx, string(step), prompt)
			if err != nil {
				return RunResult{}, err
			}

			validation := r.Validator.Validate(step, rawOutput)
			location, err := r.Artifacts.Save(task.RunID, step, StepArtifacts{
				Attempt:    attempt + 1,
				Prompt:     prompt,
				RawOutput:  rawOutput,
				Validation: validation,
			})
			if err != nil {
				return RunResult{}, err
			}
			locations[step] = location

			if validation.Pass {
				outputs[step] = validation.ParsedJSON
				stepCompleted = true
				break
			}

			attemptErrors = append(attemptErrors, strings.Join(validation.Errors, "; "))
		}

		if !stepCompleted {
			return RunResult{
				RunID:              task.RunID,
				Status:             "manual_intervention",
				ManualIntervention: true,
				FailedStep:         step,
				Outputs:            outputs,
				ArtifactLocations:  locations,
			}, nil
		}
	}

	summary, err := buildSummary(outputs[StepReviewer])
	if err != nil {
		return RunResult{}, err
	}

	summaryRaw, _ := json.Marshal(summary)
	summaryValidation := ValidationResult{Pass: true, ParsedJSON: summaryRaw}
	summaryLocation, err := r.Artifacts.Save(task.RunID, StepFinalSummary, StepArtifacts{
		Attempt:    1,
		Prompt:     "auto-generated final summary",
		RawOutput:  string(summaryRaw),
		Validation: summaryValidation,
	})
	if err != nil {
		return RunResult{}, err
	}
	locations[StepFinalSummary] = summaryLocation

	return RunResult{
		RunID:              task.RunID,
		Status:             "completed",
		ManualIntervention: false,
		Summary:            summary,
		Outputs:            outputs,
		ArtifactLocations:  locations,
	}, nil
}

func buildSummary(raw json.RawMessage) (FinalSummary, error) {
	var reviewer ReviewerOutput
	if err := json.Unmarshal(raw, &reviewer); err != nil {
		return FinalSummary{}, err
	}
	return FinalSummary{
		CriticalIssues:      reviewer.CriticalIssues,
		Improvements:        reviewer.Improvements,
		SecurityFindings:    reviewer.SecurityFindings,
		PerformanceFindings: reviewer.PerformanceFindings,
		MergeReadiness:      reviewer.MergeReadiness,
	}, nil
}
