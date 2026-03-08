package orchestrator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type JSONContractValidator struct{}

func (v JSONContractValidator) Validate(step Step, rawOutput string) ValidationResult {
	if strings.TrimSpace(rawOutput) == "" {
		return ValidationResult{Pass: false, Errors: []string{"empty output"}}
	}

	switch step {
	case StepResearcher:
		var out ResearcherOutput
		return validateStruct(rawOutput, &out, validateResearcher)
	case StepPlanner:
		var out PlannerOutput
		return validateStruct(rawOutput, &out, validatePlanner)
	case StepImplementor:
		var out ImplementorOutput
		return validateStruct(rawOutput, &out, validateImplementor)
	case StepTester:
		var out TesterOutput
		return validateStruct(rawOutput, &out, validateTester)
	case StepReviewer:
		var out ReviewerOutput
		return validateStruct(rawOutput, &out, validateReviewer)
	default:
		return ValidationResult{Pass: false, Errors: []string{fmt.Sprintf("unknown step: %s", step)}}
	}
}

func validateStruct[T any](raw string, target *T, extraChecks func(*T) []string) ValidationResult {
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		return ValidationResult{Pass: false, Errors: []string{fmt.Sprintf("invalid JSON: %v", err)}}
	}
	if dec.More() {
		return ValidationResult{Pass: false, Errors: []string{"invalid JSON: trailing content"}}
	}
	if checks := extraChecks(target); len(checks) > 0 {
		return ValidationResult{Pass: false, Errors: checks}
	}
	parsed, _ := json.Marshal(target)
	return ValidationResult{Pass: true, ParsedJSON: parsed}
}

func validateResearcher(out *ResearcherOutput) []string {
	var errs []string
	if strings.TrimSpace(out.ProblemSummary) == "" {
		errs = append(errs, "problem_summary is required")
	}
	if out.Requirements == nil {
		errs = append(errs, "requirements is required")
	}
	if out.ImpactedComponents == nil {
		errs = append(errs, "impacted_components is required")
	}
	if out.Dependencies == nil {
		errs = append(errs, "dependencies is required")
	}
	if out.EdgeCases == nil {
		errs = append(errs, "edge_cases is required")
	}
	if out.Risks == nil {
		errs = append(errs, "risks is required")
	}
	if out.OpenQuestions == nil {
		errs = append(errs, "open_questions is required")
	}
	if out.TestScenarios == nil {
		errs = append(errs, "test_scenarios is required")
	}
	return errs
}

func validatePlanner(out *PlannerOutput) []string {
	var errs []string
	if strings.TrimSpace(out.ArchitectureOverview) == "" {
		errs = append(errs, "architecture_overview is required")
	}
	if out.RequestFlow == nil {
		errs = append(errs, "request_flow is required")
	}
	if strings.TrimSpace(out.APIContract.Method) == "" {
		errs = append(errs, "api_contract.method is required")
	}
	if strings.TrimSpace(out.APIContract.Path) == "" {
		errs = append(errs, "api_contract.path is required")
	}
	if out.APIContract.RequestFields == nil {
		errs = append(errs, "api_contract.request_fields is required")
	}
	if out.APIContract.ResponseFields == nil {
		errs = append(errs, "api_contract.response_fields is required")
	}
	if out.APIContract.ErrorContract == nil {
		errs = append(errs, "api_contract.error_contract is required")
	}
	if strings.TrimSpace(out.APIContract.BackwardCompatibility) == "" {
		errs = append(errs, "api_contract.backward_compatibility is required")
	}
	if out.DBChanges == nil {
		errs = append(errs, "db_changes is required")
	}
	if out.ImplementationSteps == nil {
		errs = append(errs, "implementation_steps is required")
	}
	if out.Constraints == nil {
		errs = append(errs, "constraints is required")
	}
	if out.Assumptions == nil {
		errs = append(errs, "assumptions is required")
	}
	return errs
}

func validateImplementor(out *ImplementorOutput) []string {
	var errs []string
	if out.ChangedFiles == nil {
		errs = append(errs, "changed_files is required")
	}
	if len(out.ChangedFiles) == 0 {
		errs = append(errs, "changed_files must contain at least one file")
	}
	if out.ImplementedRules == nil {
		errs = append(errs, "implemented_rules is required")
	}
	if out.KnownLimitations == nil {
		errs = append(errs, "known_limitations is required")
	}
	if out.AreasNeedingTests == nil {
		errs = append(errs, "areas_needing_tests is required")
	}
	if out.IntegrationPoints == nil {
		errs = append(errs, "integration_points is required")
	}
	return errs
}

func validateTester(out *TesterOutput) []string {
	var errs []string
	if out.TestsAdded == nil {
		errs = append(errs, "tests_added is required")
	}
	if out.CoveredScenarios == nil {
		errs = append(errs, "covered_scenarios is required")
	}
	if out.UncoveredScenarios == nil {
		errs = append(errs, "uncovered_scenarios is required")
	}
	if out.ObservedRisks == nil {
		errs = append(errs, "observed_risks is required")
	}
	return errs
}

func validateReviewer(out *ReviewerOutput) []string {
	var errs []string
	if out.CriticalIssues == nil {
		errs = append(errs, "critical_issues is required")
	}
	if out.Improvements == nil {
		errs = append(errs, "improvements is required")
	}
	if out.SecurityFindings == nil {
		errs = append(errs, "security_findings is required")
	}
	if out.PerformanceFindings == nil {
		errs = append(errs, "performance_findings is required")
	}
	if strings.TrimSpace(out.MergeReadiness) == "" {
		errs = append(errs, "merge_readiness is required")
	}
	valid := map[string]struct{}{
		"Not ready":             {},
		"Ready with follow-ups": {},
		"Ready":                 {},
	}
	if _, ok := valid[out.MergeReadiness]; !ok {
		errs = append(errs, "merge_readiness must be one of: Not ready, Ready with follow-ups, Ready")
	}
	return errs
}

func MarshalPretty(v any) []byte {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
	return buf.Bytes()
}
