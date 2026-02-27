package agent

import "encoding/json"

type ResearchInput struct {
    Requirement string `json:"requirement"`
}

type PlanningInput struct {
    ResearchJSON json.RawMessage `json:"research_json"`
    SourceTree   []string        `json:"source_tree"`
}

type TaskSplitInput struct {
    PlanJSON json.RawMessage `json:"plan_json"`
}

type ExecutorInput struct {
    Task TaskItem `json:"task"`
}

type ExecutorOutput struct {
    Patch string `json:"patch"`
}

type TaskItem struct {
    ID            string   `json:"id"`
    Description   string   `json:"description"`
    AffectedFiles []string `json:"affected_files"`
    DependsOn     []string `json:"depends_on"`
    Status        string   `json:"status"`
}
