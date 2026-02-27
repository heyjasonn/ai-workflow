package agent

import "context"

type Agent interface {
    Execute(ctx context.Context, input any) (output any, err error)
}

type Registry struct {
    Research  Agent
    Planning  Agent
    TaskSplit Agent
    Executor  ExecutorAgent
}

type ExecutorAgent interface {
    Execute(ctx context.Context, input ExecutorInput) (ExecutorOutput, error)
}
