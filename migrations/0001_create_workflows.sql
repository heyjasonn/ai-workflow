CREATE TABLE IF NOT EXISTS workflows (
    id TEXT PRIMARY KEY,
    requirement TEXT NOT NULL,
    state VARCHAR NOT NULL,
    research_json JSONB,
    plan_json JSONB,
    task_list_json JSONB,
    execution_result_json JSONB,
    test_report_json JSONB,
    approvals_json JSONB,
    created_at TIMESTAMP NOT NULL
);
