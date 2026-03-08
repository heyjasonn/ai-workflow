# Handoff Sequence V2

1. Researcher -> Planner (validated by `researcher-to-planner.schema.json`)
2. Planner -> Implementor (validated by `planner-to-implementor.schema.json`)
3. Implementor -> Tester (validated by `implementor-to-tester.schema.json`)
4. Tester -> Reviewer (validated by `tester-to-reviewer.schema.json`)
5. Reviewer -> Final Summary (validated by `final-summary.schema.json`)

## Validation Rules

- Reject unknown keys.
- Reject missing required keys.
- Reject wrong value types.
- Retry once with explicit error feedback.
- Fail to manual intervention after second invalid output.
