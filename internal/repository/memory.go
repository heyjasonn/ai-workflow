package repository

import (
    "context"
    "sync"

    "github.com/heyjasonn/ai-workflow/internal/workflow"
)

type MemoryRepository struct {
    mu    sync.RWMutex
    store map[string]*workflow.Workflow
}

func NewMemoryRepository() *MemoryRepository {
    return &MemoryRepository{store: map[string]*workflow.Workflow{}}
}

func (r *MemoryRepository) Create(ctx context.Context, wf *workflow.Workflow) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    clone := *wf
    r.store[wf.ID] = &clone
    return nil
}

func (r *MemoryRepository) Get(ctx context.Context, id string) (*workflow.Workflow, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    wf, ok := r.store[id]
    if !ok {
        return nil, workflow.ErrNotFound
    }
    clone := *wf
    return &clone, nil
}

func (r *MemoryRepository) Update(ctx context.Context, wf *workflow.Workflow) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    if _, ok := r.store[wf.ID]; !ok {
        return workflow.ErrNotFound
    }
    clone := *wf
    r.store[wf.ID] = &clone
    return nil
}

func (r *MemoryRepository) List(ctx context.Context) ([]*workflow.Workflow, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    items := make([]*workflow.Workflow, 0, len(r.store))
    for _, wf := range r.store {
        clone := *wf
        items = append(items, &clone)
    }
    return items, nil
}
