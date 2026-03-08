package orchestrator

import (
	"context"
	"fmt"
	"sync"

	"github.com/heyjasonn/ai-workflow/internal/llm"
)

type LLMProviderAdapter struct {
	LLM llm.LLM
}

func (a LLMProviderAdapter) Run(ctx context.Context, agent string, prompt string) (string, error) {
	if a.LLM == nil {
		return "", fmt.Errorf("llm provider not configured")
	}
	full := fmt.Sprintf("Agent: %s\n%s", agent, prompt)
	return a.LLM.Complete(ctx, full)
}

type StaticSequenceProvider struct {
	mu        sync.Mutex
	Responses map[Step][]string
	counters  map[Step]int
}

func (s *StaticSequenceProvider) Run(ctx context.Context, agent string, prompt string) (string, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.counters == nil {
		s.counters = map[Step]int{}
	}
	step := Step(agent)
	series := s.Responses[step]
	if len(series) == 0 {
		return "", fmt.Errorf("no static response for step %s", step)
	}
	idx := s.counters[step]
	if idx >= len(series) {
		idx = len(series) - 1
	}
	s.counters[step]++
	return series[idx], nil
}

func DefaultStubResponses() map[Step][]string {
	return map[Step][]string{
		StepResearcher: {
			`{"problem_summary":"Add idempotent order create endpoint","requirements":["create order endpoint","enforce idempotency key"],"impacted_components":["handler","service","repository"],"dependencies":["postgres"],"edge_cases":["duplicate key","invalid payload"],"risks":["duplicate order creation"],"open_questions":[],"test_scenarios":["happy path","duplicate key"]}`,
		},
		StepPlanner: {
			`{"architecture_overview":"Use existing handler-service-repository layers","request_flow":["handler validate request","service enforce idempotency","repository upsert order"],"api_contract":{"method":"POST","path":"/v1/orders","request_fields":["idempotency_key","items"],"response_fields":["order_id","status"],"error_contract":["400 validation_error","409 duplicate"],"backward_compatibility":"additive endpoint"},"db_changes":["add unique index on idempotency key"],"implementation_steps":["add DTO","implement service","add repository method"],"constraints":["no new dependency"],"assumptions":["existing transaction helper available"]}`,
		},
		StepImplementor: {
			`{"changed_files":["internal/order/handler.go","internal/order/service.go"],"implemented_rules":["idempotency key check","input validation"],"known_limitations":["no async processing"],"areas_needing_tests":["duplicate key path","db failure path"],"integration_points":["postgres orders table"]}`,
		},
		StepTester: {
			`{"tests_added":["internal/order/service_test.go"],"covered_scenarios":["happy path","duplicate key"],"uncovered_scenarios":["network partition"],"observed_risks":["race on concurrent duplicate requests"]}`,
		},
		StepReviewer: {
			`{"critical_issues":[],"improvements":["add integration test for rollback"],"security_findings":["validate idempotency key length"],"performance_findings":["add index to lookup column"],"merge_readiness":"Ready with follow-ups"}`,
		},
	}
}
