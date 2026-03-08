package orchestrator

import "encoding/json"

type Step string

const (
	StepResearcher   Step = "researcher"
	StepPlanner      Step = "planner"
	StepImplementor  Step = "implementor"
	StepTester       Step = "tester"
	StepReviewer     Step = "reviewer"
	StepFinalSummary Step = "final-summary"
)

var WorkflowSteps = []Step{
	StepResearcher,
	StepPlanner,
	StepImplementor,
	StepTester,
	StepReviewer,
}

type TaskInput struct {
	RunID       string            `json:"run_id"`
	Category    string            `json:"category"`
	Requirement string            `json:"requirement"`
	Context     map[string]string `json:"context,omitempty"`
}

type PromptContext struct {
	Task        TaskInput
	Previous    map[Step]json.RawMessage
	RetryErrors []string
}

type ValidationResult struct {
	Pass       bool            `json:"pass"`
	Errors     []string        `json:"errors,omitempty"`
	ParsedJSON json.RawMessage `json:"parsed_json,omitempty"`
}

type StepArtifacts struct {
	Attempt    int              `json:"attempt"`
	Prompt     string           `json:"prompt"`
	RawOutput  string           `json:"raw_output"`
	Validation ValidationResult `json:"validation"`
}

type RunResult struct {
	RunID              string                   `json:"run_id"`
	Status             string                   `json:"status"`
	ManualIntervention bool                     `json:"manual_intervention"`
	FailedStep         Step                     `json:"failed_step,omitempty"`
	Summary            FinalSummary             `json:"summary,omitempty"`
	Outputs            map[Step]json.RawMessage `json:"outputs,omitempty"`
	ArtifactLocations  map[Step]string          `json:"artifact_locations,omitempty"`
}

type ResearcherOutput struct {
	ProblemSummary     string   `json:"problem_summary"`
	Requirements       []string `json:"requirements"`
	ImpactedComponents []string `json:"impacted_components"`
	Dependencies       []string `json:"dependencies"`
	EdgeCases          []string `json:"edge_cases"`
	Risks              []string `json:"risks"`
	OpenQuestions      []string `json:"open_questions"`
	TestScenarios      []string `json:"test_scenarios"`
}

type APIContract struct {
	Method                string   `json:"method"`
	Path                  string   `json:"path"`
	RequestFields         []string `json:"request_fields"`
	ResponseFields        []string `json:"response_fields"`
	ErrorContract         []string `json:"error_contract"`
	BackwardCompatibility string   `json:"backward_compatibility"`
}

type PlannerOutput struct {
	ArchitectureOverview string      `json:"architecture_overview"`
	RequestFlow          []string    `json:"request_flow"`
	APIContract          APIContract `json:"api_contract"`
	DBChanges            []string    `json:"db_changes"`
	ImplementationSteps  []string    `json:"implementation_steps"`
	Constraints          []string    `json:"constraints"`
	Assumptions          []string    `json:"assumptions"`
}

type ImplementorOutput struct {
	ChangedFiles      []string `json:"changed_files"`
	ImplementedRules  []string `json:"implemented_rules"`
	KnownLimitations  []string `json:"known_limitations"`
	AreasNeedingTests []string `json:"areas_needing_tests"`
	IntegrationPoints []string `json:"integration_points"`
}

type TesterOutput struct {
	TestsAdded         []string `json:"tests_added"`
	CoveredScenarios   []string `json:"covered_scenarios"`
	UncoveredScenarios []string `json:"uncovered_scenarios"`
	ObservedRisks      []string `json:"observed_risks"`
}

type ReviewerOutput struct {
	CriticalIssues      []string `json:"critical_issues"`
	Improvements        []string `json:"improvements"`
	SecurityFindings    []string `json:"security_findings"`
	PerformanceFindings []string `json:"performance_findings"`
	MergeReadiness      string   `json:"merge_readiness"`
}

type FinalSummary struct {
	CriticalIssues      []string `json:"critical_issues"`
	Improvements        []string `json:"improvements"`
	SecurityFindings    []string `json:"security_findings"`
	PerformanceFindings []string `json:"performance_findings"`
	MergeReadiness      string   `json:"merge_readiness"`
}
