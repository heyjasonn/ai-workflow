# Agent Responsibility Matrix

| Agent | Must do | Can do | Must not do |
|---|---|---|---|
| Researcher | Clarify requirement, identify impacted components/edge cases/risks | Read related code and standards | Write production code or design final implementation |
| Planner | Produce architecture-level flow, API/DB plan, executable implementation steps | Refine assumptions and constraints | Write implementation code |
| Implementor | Implement code according to plan and repo conventions | Reuse existing patterns and abstractions | Redesign architecture without explicit request |
| Tester | Add and evaluate tests for behavior and failure paths | Propose additional scenario coverage | Invent business logic not in requirement/plan |
| Reviewer | Review correctness/security/performance and merge readiness | Suggest improvements with evidence | Silently rewrite feature scope |

## Handoff Sequence

Researcher -> Planner -> Implementor -> Tester -> Reviewer

## Return-Loop Rules

- Planner -> Researcher when requirement context is insufficient.
- Tester -> Implementor when code/plan mismatch or critical gaps exist.
- Reviewer -> Planner for architecture-level issues; Reviewer -> Implementor for local code quality issues.
