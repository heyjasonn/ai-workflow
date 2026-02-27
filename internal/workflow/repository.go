package workflow

import (
    "context"
    "errors"
)

type Repository interface {
    Create(ctx context.Context, wf *Workflow) error
    Get(ctx context.Context, id string) (*Workflow, error)
    Update(ctx context.Context, wf *Workflow) error
    List(ctx context.Context) ([]*Workflow, error)
}

var ErrNotFound = errors.New("workflow not found")
