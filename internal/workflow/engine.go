package workflow

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/heyjasonn/ai-workflow/internal/agent"
    "github.com/heyjasonn/ai-workflow/internal/diff"
)

type Engine struct {
    Repo     Repository
    Agents   agent.Registry
    RootDir  string
    TestRunner TestRunner
}

type TestRunner interface {
    Run(ctx context.Context) (TestReport, error)
}

func (e *Engine) CreateWorkflow(ctx context.Context, requirement string) (*Workflow, error) {
    wf := &Workflow{
        ID:          agent.NewWorkflowID(),
        Requirement: requirement,
        State:       StateInit,
        CreatedAt:   time.Now().UTC(),
    }
    if err := e.Repo.Create(ctx, wf); err != nil {
        return nil, err
    }
    return wf, nil
}

func (e *Engine) ListWorkflows(ctx context.Context) ([]*Workflow, error) {
    return e.Repo.List(ctx)
}

func (e *Engine) GetStatus(ctx context.Context, workflowID string) (*WorkflowStatus, error) {
    wf, err := e.Repo.Get(ctx, workflowID)
    if err != nil {
        return nil, err
    }
    approvals := parseApprovals(wf.ApprovalsJSON)
    next := nextStepForState(wf, approvals)
    approved := false
    if next != "" {
        if record, ok := approvals[string(next)]; ok && record.Approved {
            approved = true
        }
    }
    pendingTasks, patchReady := summarizeTasks(wf.TaskListJSON)
    return &WorkflowStatus{
        ID:        wf.ID,
        State:     wf.State,
        NextStep:  next,
        Approved:  approved,
        Approvals: approvals,
        PendingTasks: pendingTasks,
        PatchReady:   patchReady,
    }, nil
}

func (e *Engine) ApproveStep(ctx context.Context, workflowID string, step ApprovalStep) (*Workflow, error) {
    wf, err := e.Repo.Get(ctx, workflowID)
    if err != nil {
        return nil, err
    }

    approvals := parseApprovals(wf.ApprovalsJSON)
    approvals[string(step)] = ApprovalRecord{Step: step, Approved: true, ApprovedAt: time.Now().UTC()}
    approvalsJSON, _ := json.Marshal(approvals)
    wf.ApprovalsJSON = approvalsJSON
    if err := e.Repo.Update(ctx, wf); err != nil {
        return nil, err
    }

    if step == ApprovalPatch {
        return wf, nil
    }

    return e.RunStep(ctx, wf.ID, step)
}

func (e *Engine) RejectStep(ctx context.Context, workflowID string, step ApprovalStep) (*Workflow, error) {
    wf, err := e.Repo.Get(ctx, workflowID)
    if err != nil {
        return nil, err
    }
    approvals := parseApprovals(wf.ApprovalsJSON)
    approvals[string(step)] = ApprovalRecord{Step: step, Approved: false, ApprovedAt: time.Now().UTC()}
    approvalsJSON, _ := json.Marshal(approvals)
    wf.ApprovalsJSON = approvalsJSON
    if err := e.Repo.Update(ctx, wf); err != nil {
        return nil, err
    }
    return wf, nil
}

func (e *Engine) RunStep(ctx context.Context, workflowID string, step ApprovalStep) (*Workflow, error) {
    wf, err := e.Repo.Get(ctx, workflowID)
    if err != nil {
        return nil, err
    }
    if err := requireApproval(wf, step); err != nil {
        return nil, err
    }

    switch step {
    case ApprovalResearch:
        return e.runResearch(ctx, wf)
    case ApprovalPlan:
        return e.runPlan(ctx, wf)
    case ApprovalTaskSplit:
        return e.runTaskSplit(ctx, wf)
    case ApprovalPatch:
        return e.runExecution(ctx, wf)
    case ApprovalTest:
        return e.runTests(ctx, wf)
    default:
        return nil, fmt.Errorf("unknown step: %s", step)
    }
}

func requireApproval(wf *Workflow, step ApprovalStep) error {
    switch step {
    case ApprovalResearch, ApprovalPlan, ApprovalTaskSplit, ApprovalTest:
    default:
        return nil
    }
    approvals := parseApprovals(wf.ApprovalsJSON)
    record, ok := approvals[string(step)]
    if !ok || !record.Approved {
        return fmt.Errorf("step %s not approved", step)
    }
    return nil
}

func (e *Engine) runResearch(ctx context.Context, wf *Workflow) (*Workflow, error) {
    if wf.State != StateInit {
        return nil, fmt.Errorf("cannot run research from state %s", wf.State)
    }
    output, err := e.Agents.Research.Execute(ctx, agent.ResearchInput{Requirement: wf.Requirement})
    if err != nil {
        return nil, err
    }
    payload, err := json.Marshal(output)
    if err != nil {
        return nil, err
    }
    if err := validateResearchJSON(payload); err != nil {
        return nil, fmt.Errorf("research JSON invalid: %w", err)
    }
    wf.ResearchJSON = payload
    if err := ValidateTransition(wf.State, StateResearchDone); err != nil {
        return nil, err
    }
    wf.State = StateResearchDone
    if err := e.Repo.Update(ctx, wf); err != nil {
        return nil, err
    }
    return wf, nil
}

func (e *Engine) runPlan(ctx context.Context, wf *Workflow) (*Workflow, error) {
    if wf.State != StateResearchDone {
        return nil, fmt.Errorf("cannot run plan from state %s", wf.State)
    }
    if len(wf.ResearchJSON) == 0 {
        return nil, errors.New("missing research JSON")
    }

    tree, err := e.sourceTree()
    if err != nil {
        return nil, err
    }

    input := agent.PlanningInput{
        ResearchJSON: wf.ResearchJSON,
        SourceTree:   tree,
    }
    output, err := e.Agents.Planning.Execute(ctx, input)
    if err != nil {
        return nil, err
    }
    payload, err := json.Marshal(output)
    if err != nil {
        return nil, err
    }
    if err := validatePlanJSON(payload); err != nil {
        return nil, fmt.Errorf("plan JSON invalid: %w", err)
    }

    wf.PlanJSON = payload
    if err := ValidateTransition(wf.State, StatePlanDone); err != nil {
        return nil, err
    }
    wf.State = StatePlanDone
    if err := e.Repo.Update(ctx, wf); err != nil {
        return nil, err
    }
    return wf, nil
}

func (e *Engine) runTaskSplit(ctx context.Context, wf *Workflow) (*Workflow, error) {
    if wf.State != StatePlanDone {
        return nil, fmt.Errorf("cannot run task split from state %s", wf.State)
    }
    if len(wf.PlanJSON) == 0 {
        return nil, errors.New("missing plan JSON")
    }

    input := agent.TaskSplitInput{PlanJSON: wf.PlanJSON}
    output, err := e.Agents.TaskSplit.Execute(ctx, input)
    if err != nil {
        return nil, err
    }
    payload, err := json.Marshal(output)
    if err != nil {
        return nil, err
    }
    if err := validateTaskListJSON(payload); err != nil {
        return nil, fmt.Errorf("task list JSON invalid: %w", err)
    }

    wf.TaskListJSON = payload
    if err := ValidateTransition(wf.State, StateTaskSplitDone); err != nil {
        return nil, err
    }
    wf.State = StateTaskSplitDone
    if err := e.Repo.Update(ctx, wf); err != nil {
        return nil, err
    }
    return wf, nil
}

func (e *Engine) runExecution(ctx context.Context, wf *Workflow) (*Workflow, error) {
    if wf.State != StateTaskSplitDone && wf.State != StateExecuting {
        return nil, fmt.Errorf("cannot run execution from state %s", wf.State)
    }

    tasks, err := parseTasks(wf.TaskListJSON)
    if err != nil {
        return nil, err
    }
    nextTask, idx := nextPendingTask(tasks)
    if nextTask == nil {
        return nil, errors.New("no pending task")
    }

    output, err := e.Agents.Executor.Execute(ctx, agent.ExecutorInput{Task: agent.TaskItem{
        ID:            nextTask.ID,
        Description:   nextTask.Description,
        AffectedFiles: nextTask.AffectedFiles,
        DependsOn:     nextTask.DependsOn,
        Status:        nextTask.Status,
    }})
    if err != nil {
        return nil, err
    }

    if err := validatePatch(output.Patch, e.RootDir); err != nil {
        return nil, err
    }

    execResult := ExecutionResult{TaskID: nextTask.ID, Patch: output.Patch}
    payload, err := json.Marshal(execResult)
    if err != nil {
        return nil, err
    }

    tasks[idx].Status = "PATCH_READY"
    taskPayload, err := json.Marshal(tasks)
    if err != nil {
        return nil, err
    }

    approvals := parseApprovals(wf.ApprovalsJSON)
    approvals[string(ApprovalPatch)] = ApprovalRecord{Step: ApprovalPatch, Approved: false, ApprovedAt: time.Now().UTC()}
    approvalsJSON, err := json.Marshal(approvals)
    if err != nil {
        return nil, err
    }

    wf.ExecutionResultJSON = payload
    wf.TaskListJSON = taskPayload
    wf.ApprovalsJSON = approvalsJSON
    if wf.State == StateTaskSplitDone {
        if err := ValidateTransition(wf.State, StateExecuting); err != nil {
            return nil, err
        }
        wf.State = StateExecuting
    }
    if err := e.Repo.Update(ctx, wf); err != nil {
        return nil, err
    }
    return wf, nil
}

func (e *Engine) ApplyPatch(ctx context.Context, workflowID string) (*Workflow, error) {
    wf, err := e.Repo.Get(ctx, workflowID)
    if err != nil {
        return nil, err
    }
    if wf.State != StateExecuting {
        return nil, fmt.Errorf("cannot apply patch from state %s", wf.State)
    }
    approvals := parseApprovals(wf.ApprovalsJSON)
    record, ok := approvals[string(ApprovalPatch)]
    if !ok || !record.Approved {
        return nil, fmt.Errorf("step %s not approved", ApprovalPatch)
    }
    var execResult ExecutionResult
    if err := json.Unmarshal(wf.ExecutionResultJSON, &execResult); err != nil {
        return nil, err
    }
    if strings.TrimSpace(execResult.Patch) == "" {
        return nil, errors.New("missing patch")
    }
    if err := diff.ApplyPatch(ctx, e.RootDir, execResult.Patch); err != nil {
        return nil, err
    }

    tasks, err := parseTasks(wf.TaskListJSON)
    if err != nil {
        return nil, err
    }
    for i := range tasks {
        if tasks[i].ID == execResult.TaskID {
            tasks[i].Status = "APPLIED"
            break
        }
    }
    payload, err := json.Marshal(tasks)
    if err != nil {
        return nil, err
    }
    wf.TaskListJSON = payload
    if err := e.Repo.Update(ctx, wf); err != nil {
        return nil, err
    }
    return wf, nil
}

func (e *Engine) runTests(ctx context.Context, wf *Workflow) (*Workflow, error) {
    if wf.State != StateExecuting {
        return nil, fmt.Errorf("cannot run tests from state %s", wf.State)
    }
    if e.TestRunner == nil {
        return nil, errors.New("test runner not configured")
    }

    report, err := e.TestRunner.Run(ctx)
    if err != nil {
        return nil, err
    }
    payload, err := json.Marshal(report)
    if err != nil {
        return nil, err
    }

    wf.TestReportJSON = payload
    if err := ValidateTransition(wf.State, StateTestDone); err != nil {
        return nil, err
    }
    wf.State = StateTestDone

    if report.Status == "pass" {
        if err := ValidateTransition(wf.State, StateCompleted); err == nil {
            wf.State = StateCompleted
        }
    } else {
        tasks, err := parseTasks(wf.TaskListJSON)
        if err != nil {
            return nil, err
        }
        fixTask := TaskSplitItem{
            ID:          fmt.Sprintf("FIX-%d", time.Now().UTC().Unix()),
            Description: "Fix failing tests",
            Status:      "PENDING",
        }
        tasks = append(tasks, fixTask)
        taskPayload, err := json.Marshal(tasks)
        if err != nil {
            return nil, err
        }
        wf.TaskListJSON = taskPayload
        if err := ValidateTransition(wf.State, StateExecuting); err != nil {
            return nil, err
        }
        wf.State = StateExecuting
    }

    if err := e.Repo.Update(ctx, wf); err != nil {
        return nil, err
    }
    return wf, nil
}

func parseApprovals(raw json.RawMessage) map[string]ApprovalRecord {
    if len(raw) == 0 {
        return map[string]ApprovalRecord{}
    }
    var approvals map[string]ApprovalRecord
    if err := json.Unmarshal(raw, &approvals); err != nil {
        return map[string]ApprovalRecord{}
    }
    return approvals
}

func parseTasks(raw json.RawMessage) ([]TaskSplitItem, error) {
    if len(raw) == 0 {
        return nil, errors.New("missing task list")
    }
    var tasks []TaskSplitItem
    if err := json.Unmarshal(raw, &tasks); err != nil {
        return nil, err
    }
    return tasks, nil
}

func nextPendingTask(tasks []TaskSplitItem) (*TaskSplitItem, int) {
    for i := range tasks {
        if tasks[i].Status == "" || tasks[i].Status == "PENDING" {
            tasks[i].Status = "IN_PROGRESS"
            return &tasks[i], i
        }
    }
    return nil, -1
}

func (e *Engine) sourceTree() ([]string, error) {
    var files []string
    err := filepath.Walk(e.RootDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if info.IsDir() {
            if info.Name() == ".git" {
                return filepath.SkipDir
            }
            return nil
        }
        rel, err := filepath.Rel(e.RootDir, path)
        if err != nil {
            return err
        }
        files = append(files, rel)
        return nil
    })
    if err != nil {
        return nil, err
    }
    return files, nil
}

func validatePatch(patch string, root string) error {
    files := diff.ExtractFilePaths(patch)
    if len(files) == 0 {
        return errors.New("patch contains no file changes")
    }
    for _, f := range files {
        if f.IsNewFile {
            continue
        }
        full := filepath.Join(root, f.Path)
        if _, err := os.Stat(full); err != nil {
            return fmt.Errorf("patch references missing file: %s", f.Path)
        }
    }
    return nil
}

func validateResearchJSON(payload []byte) error {
    var out ResearchOutput
    if err := decodeStrict(payload, &out); err != nil {
        return err
    }
    if out.CompetitorFeatures == nil || out.DesignPatterns == nil || out.Recommendations == nil {
        return errors.New("research JSON missing required keys")
    }
    return nil
}

func validatePlanJSON(payload []byte) error {
    var out PlanOutput
    if err := decodeStrict(payload, &out); err != nil {
        return err
    }
    if out.NewPackages == nil || out.ModifiedFiles == nil || out.Interfaces == nil || out.DBChanges == nil || out.Risks == nil {
        return errors.New("plan JSON missing required keys")
    }
    return nil
}

func validateTaskListJSON(payload []byte) error {
    var out []TaskSplitItem
    if err := decodeStrict(payload, &out); err != nil {
        return err
    }
    for _, item := range out {
        if item.ID == "" || item.Description == "" {
            return errors.New("task list contains task missing id or description")
        }
    }
    return nil
}

func decodeStrict(payload []byte, target any) error {
    dec := json.NewDecoder(strings.NewReader(string(payload)))
    dec.DisallowUnknownFields()
    if err := dec.Decode(target); err != nil {
        return fmt.Errorf("invalid JSON payload: %w", err)
    }
    return nil
}

func nextStepForState(wf *Workflow, approvals map[string]ApprovalRecord) ApprovalStep {
    switch wf.State {
    case StateInit:
        return ApprovalResearch
    case StateResearchDone:
        return ApprovalPlan
    case StatePlanDone:
        return ApprovalTaskSplit
    case StateTaskSplitDone:
        return ApprovalPatch
    case StateExecuting:
        if len(wf.ExecutionResultJSON) == 0 {
            return ApprovalPatch
        }
        if record, ok := approvals[string(ApprovalPatch)]; !ok || !record.Approved {
            return ApprovalPatch
        }
        return ApprovalTest
    case StateTestDone:
        return ""
    case StateCompleted, StateFailed:
        return ""
    default:
        return ""
    }
}

func summarizeTasks(raw json.RawMessage) ([]string, bool) {
    tasks, err := parseTasks(raw)
    if err != nil {
        return nil, false
    }
    var pending []string
    patchReady := false
    for _, task := range tasks {
        if task.Status == "" || task.Status == "PENDING" || task.Status == "IN_PROGRESS" {
            pending = append(pending, task.ID)
        }
        if task.Status == "PATCH_READY" {
            patchReady = true
        }
    }
    return pending, patchReady
}
