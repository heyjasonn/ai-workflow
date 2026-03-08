package orchestrator

import (
	"encoding/json"
	"fmt"
	"strings"
)

type DefaultPromptBuilder struct{}

func (b DefaultPromptBuilder) Build(step Step, ctx PromptContext) (string, error) {
	var sb strings.Builder
	sb.WriteString("You are part of a multi-agent backend workflow. Return strict JSON only.\n")
	sb.WriteString(fmt.Sprintf("Task category: %s\n", ctx.Task.Category))
	sb.WriteString(fmt.Sprintf("Requirement: %s\n", ctx.Task.Requirement))
	if len(ctx.Task.Context) > 0 {
		ctxJSON, err := json.Marshal(ctx.Task.Context)
		if err != nil {
			return "", err
		}
		sb.WriteString(fmt.Sprintf("Additional context: %s\n", string(ctxJSON)))
	}

	if len(ctx.Previous) > 0 {
		sb.WriteString("Previous handoff outputs:\n")
		for _, s := range WorkflowSteps {
			if out, ok := ctx.Previous[s]; ok {
				sb.WriteString(fmt.Sprintf("- %s: %s\n", s, string(out)))
			}
		}
	}

	if len(ctx.RetryErrors) > 0 {
		sb.WriteString("Validation errors from previous attempt:\n")
		for _, err := range ctx.RetryErrors {
			sb.WriteString(fmt.Sprintf("- %s\n", err))
		}
		sb.WriteString("Fix every error and return complete schema.\n")
	}

	sb.WriteString("Required output schema:\n")
	sb.WriteString(requiredSchema(step))
	return sb.String(), nil
}

func requiredSchema(step Step) string {
	switch step {
	case StepResearcher:
		return `{"problem_summary":"string","requirements":["string"],"impacted_components":["string"],"dependencies":["string"],"edge_cases":["string"],"risks":["string"],"open_questions":["string"],"test_scenarios":["string"]}`
	case StepPlanner:
		return `{"architecture_overview":"string","request_flow":["string"],"api_contract":{"method":"string","path":"string","request_fields":["string"],"response_fields":["string"],"error_contract":["string"],"backward_compatibility":"string"},"db_changes":["string"],"implementation_steps":["string"],"constraints":["string"],"assumptions":["string"]}`
	case StepImplementor:
		return `{"changed_files":["string"],"implemented_rules":["string"],"known_limitations":["string"],"areas_needing_tests":["string"],"integration_points":["string"]}`
	case StepTester:
		return `{"tests_added":["string"],"covered_scenarios":["string"],"uncovered_scenarios":["string"],"observed_risks":["string"]}`
	case StepReviewer:
		return `{"critical_issues":["string"],"improvements":["string"],"security_findings":["string"],"performance_findings":["string"],"merge_readiness":"Not ready|Ready with follow-ups|Ready"}`
	default:
		return "{}"
	}
}
