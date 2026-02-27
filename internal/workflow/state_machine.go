package workflow

import "fmt"

type Transition struct {
    From WorkflowState
    To   WorkflowState
}

var allowedTransitions = map[Transition]struct{}{
    {From: StateInit, To: StateResearchDone}:      {},
    {From: StateResearchDone, To: StatePlanDone}:  {},
    {From: StatePlanDone, To: StateTaskSplitDone}: {},
    {From: StateTaskSplitDone, To: StateExecuting}: {},
    {From: StateExecuting, To: StateTestDone}:     {},
    {From: StateTestDone, To: StateCompleted}:     {},
    {From: StateTestDone, To: StateExecuting}:     {},
    {From: StateExecuting, To: StateFailed}:       {},
    {From: StateTestDone, To: StateFailed}:        {},
}

func ValidateTransition(from, to WorkflowState) error {
    if _, ok := allowedTransitions[Transition{From: from, To: to}]; ok {
        return nil
    }
    return fmt.Errorf("invalid transition from %s to %s", from, to)
}
