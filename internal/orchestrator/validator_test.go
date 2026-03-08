package orchestrator

import "testing"

func TestJSONContractValidator_Contracts(t *testing.T) {
	v := JSONContractValidator{}

	tests := []struct {
		name     string
		step     Step
		raw      string
		wantPass bool
	}{
		{
			name:     "researcher pass",
			step:     StepResearcher,
			raw:      `{"problem_summary":"s","requirements":[],"impacted_components":[],"dependencies":[],"edge_cases":[],"risks":[],"open_questions":[],"test_scenarios":[]}`,
			wantPass: true,
		},
		{
			name:     "researcher missing field",
			step:     StepResearcher,
			raw:      `{"problem_summary":"s","requirements":[],"impacted_components":[],"dependencies":[],"edge_cases":[],"risks":[],"open_questions":[]}`,
			wantPass: false,
		},
		{
			name:     "planner pass",
			step:     StepPlanner,
			raw:      `{"architecture_overview":"x","request_flow":[],"api_contract":{"method":"POST","path":"/v1/x","request_fields":[],"response_fields":[],"error_contract":[],"backward_compatibility":"additive"},"db_changes":[],"implementation_steps":[],"constraints":[],"assumptions":[]}`,
			wantPass: true,
		},
		{
			name:     "planner wrong type",
			step:     StepPlanner,
			raw:      `{"architecture_overview":"x","request_flow":{},"api_contract":{"method":"POST","path":"/v1/x","request_fields":[],"response_fields":[],"error_contract":[],"backward_compatibility":"additive"},"db_changes":[],"implementation_steps":[],"constraints":[],"assumptions":[]}`,
			wantPass: false,
		},
		{
			name:     "implementor pass",
			step:     StepImplementor,
			raw:      `{"changed_files":["a.go"],"implemented_rules":[],"known_limitations":[],"areas_needing_tests":[],"integration_points":[]}`,
			wantPass: true,
		},
		{
			name:     "implementor unknown key",
			step:     StepImplementor,
			raw:      `{"changed_files":["a.go"],"implemented_rules":[],"known_limitations":[],"areas_needing_tests":[],"integration_points":[],"extra":1}`,
			wantPass: false,
		},
		{
			name:     "tester pass",
			step:     StepTester,
			raw:      `{"tests_added":[],"covered_scenarios":[],"uncovered_scenarios":[],"observed_risks":[]}`,
			wantPass: true,
		},
		{
			name:     "tester missing",
			step:     StepTester,
			raw:      `{"tests_added":[],"covered_scenarios":[],"uncovered_scenarios":[]}`,
			wantPass: false,
		},
		{
			name:     "reviewer pass",
			step:     StepReviewer,
			raw:      `{"critical_issues":[],"improvements":[],"security_findings":[],"performance_findings":[],"merge_readiness":"Ready"}`,
			wantPass: true,
		},
		{
			name:     "reviewer enum invalid",
			step:     StepReviewer,
			raw:      `{"critical_issues":[],"improvements":[],"security_findings":[],"performance_findings":[],"merge_readiness":"MAYBE"}`,
			wantPass: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := v.Validate(tc.step, tc.raw)
			if res.Pass != tc.wantPass {
				t.Fatalf("pass mismatch: got %v want %v errors=%v", res.Pass, tc.wantPass, res.Errors)
			}
		})
	}
}
