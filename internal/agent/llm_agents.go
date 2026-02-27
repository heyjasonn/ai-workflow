package agent

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"

    "github.com/heyjasonn/ai-workflow/internal/llm"
)

type ResearchAgent struct {
    LLM llm.LLM
}

func (a ResearchAgent) Execute(ctx context.Context, input any) (any, error) {
    in, ok := input.(ResearchInput)
    if !ok {
        return nil, errors.New("invalid research input")
    }
    prompt := SystemInstruction + "\n" + fmt.Sprintf(ResearchPromptTemplate, in.Requirement)
    raw, err := a.LLM.Complete(ctx, prompt)
    if err != nil {
        return nil, err
    }
    var out struct {
        CompetitorFeatures []string `json:"competitor_features"`
        DesignPatterns     []string `json:"design_patterns"`
        Recommendations    []string `json:"recommendations"`
    }
    if err := json.Unmarshal([]byte(raw), &out); err != nil {
        return nil, err
    }
    return out, nil
}

type PlanningAgent struct {
    LLM llm.LLM
}

func (a PlanningAgent) Execute(ctx context.Context, input any) (any, error) {
    in, ok := input.(PlanningInput)
    if !ok {
        return nil, errors.New("invalid planning input")
    }
    prompt := SystemInstruction + "\n" + fmt.Sprintf(PlanningPromptTemplate, string(in.ResearchJSON), in.SourceTree)
    raw, err := a.LLM.Complete(ctx, prompt)
    if err != nil {
        return nil, err
    }
    var out struct {
        NewPackages   []string `json:"new_packages"`
        ModifiedFiles []string `json:"modified_files"`
        Interfaces    []string `json:"interfaces"`
        DBChanges     []string `json:"db_changes"`
        Risks         []string `json:"risks"`
    }
    if err := json.Unmarshal([]byte(raw), &out); err != nil {
        return nil, err
    }
    return out, nil
}

type TaskSplitAgent struct {
    LLM llm.LLM
}

func (a TaskSplitAgent) Execute(ctx context.Context, input any) (any, error) {
    in, ok := input.(TaskSplitInput)
    if !ok {
        return nil, errors.New("invalid task split input")
    }
    prompt := SystemInstruction + "\n" + fmt.Sprintf(TaskSplitPromptTemplate, string(in.PlanJSON))
    raw, err := a.LLM.Complete(ctx, prompt)
    if err != nil {
        return nil, err
    }
    var out []TaskItem
    if err := json.Unmarshal([]byte(raw), &out); err != nil {
        return nil, err
    }
    return out, nil
}

type LLMExecutorAgent struct {
    LLM llm.LLM
}

func (a LLMExecutorAgent) Execute(ctx context.Context, input ExecutorInput) (ExecutorOutput, error) {
    prompt := SystemInstruction + "\n" + fmt.Sprintf(ExecutorPromptTemplate, input.Task.Description)
    raw, err := a.LLM.Complete(ctx, prompt)
    if err != nil {
        return ExecutorOutput{}, err
    }
    return ExecutorOutput{Patch: raw}, nil
}
