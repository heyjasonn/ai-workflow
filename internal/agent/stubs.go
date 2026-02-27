package agent

import (
    "context"
    "errors"
)

type StaticAgent struct {
    Output any
}

func (s StaticAgent) Execute(ctx context.Context, input any) (any, error) {
    if s.Output == nil {
        return nil, errors.New("agent not configured")
    }
    return s.Output, nil
}

type StaticExecutor struct {
    Patch string
}

func (s StaticExecutor) Execute(ctx context.Context, input ExecutorInput) (ExecutorOutput, error) {
    if s.Patch == "" {
        return ExecutorOutput{}, errors.New("executor not configured")
    }
    return ExecutorOutput{Patch: s.Patch}, nil
}
