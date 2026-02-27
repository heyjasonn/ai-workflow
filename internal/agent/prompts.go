package agent

const (
    SystemInstruction = "You are a deterministic transformer. Return only valid JSON that matches the required schema. No extra keys."

    ResearchPromptTemplate = "Research competitors and patterns for the requirement below. Return strict JSON with keys competitor_features, design_patterns, recommendations. Requirement: %s"

    PlanningPromptTemplate = "Generate an implementation plan. Return strict JSON with keys new_packages, modified_files, interfaces, db_changes, risks. Research: %s. SourceTree: %v"

    TaskSplitPromptTemplate = "Split the plan into tasks. Return a JSON array of tasks with id, description, affected_files, depends_on. Plan: %s"

    ExecutorPromptTemplate = "Generate a unified diff patch for the task. Output only the patch. Task: %s"
)
