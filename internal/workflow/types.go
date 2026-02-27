package workflow

import (
    "encoding/json"
    "time"
)

type WorkflowState string

const (
    StateInit          WorkflowState = "INIT"
    StateResearchDone  WorkflowState = "RESEARCH_DONE"
    StatePlanDone      WorkflowState = "PLAN_DONE"
    StateTaskSplitDone WorkflowState = "TASK_SPLIT_DONE"
    StateExecuting     WorkflowState = "EXECUTING"
    StateTestDone      WorkflowState = "TEST_DONE"
    StateCompleted     WorkflowState = "COMPLETED"
    StateFailed        WorkflowState = "FAILED"
)

type ApprovalStep string

const (
    ApprovalResearch  ApprovalStep = "research"
    ApprovalPlan      ApprovalStep = "plan"
    ApprovalTaskSplit ApprovalStep = "task_split"
    ApprovalPatch     ApprovalStep = "patch"
    ApprovalTest      ApprovalStep = "test"
)

type Workflow struct {
    ID                  string
    Requirement         string
    State               WorkflowState
    ResearchJSON        json.RawMessage
    PlanJSON            json.RawMessage
    TaskListJSON        json.RawMessage
    ExecutionResultJSON json.RawMessage
    TestReportJSON      json.RawMessage
    ApprovalsJSON       json.RawMessage
    CreatedAt           time.Time
}

type ApprovalRecord struct {
    Step      ApprovalStep `json:"step"`
    Approved  bool         `json:"approved"`
    ApprovedAt time.Time   `json:"approved_at"`
}

type ResearchOutput struct {
    CompetitorFeatures []string `json:"competitor_features"`
    DesignPatterns     []string `json:"design_patterns"`
    Recommendations    []string `json:"recommendations"`
}

type PlanOutput struct {
    NewPackages   []string `json:"new_packages"`
    ModifiedFiles []string `json:"modified_files"`
    Interfaces    []string `json:"interfaces"`
    DBChanges     []string `json:"db_changes"`
    Risks         []string `json:"risks"`
}

type TaskSplitItem struct {
    ID            string   `json:"id"`
    Description   string   `json:"description"`
    AffectedFiles []string `json:"affected_files"`
    DependsOn     []string `json:"depends_on"`
    Status        string   `json:"status"`
}

type ExecutionResult struct {
    TaskID string `json:"task_id"`
    Patch  string `json:"patch"`
}

type TestReport struct {
    Status         string   `json:"status"`
    FailedPackages []string `json:"failed_packages"`
    Errors         []string `json:"errors"`
}

type WorkflowStatus struct {
    ID           string         `json:"id"`
    State        WorkflowState  `json:"state"`
    NextStep     ApprovalStep   `json:"next_step"`
    Approved     bool           `json:"approved"`
    Approvals    map[string]ApprovalRecord `json:"approvals"`
    PendingTasks []string       `json:"pending_tasks"`
    PatchReady   bool           `json:"patch_ready"`
}
