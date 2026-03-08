# DB Migration Rules

- Prefer additive schema changes before destructive cleanup.
- Use nullable/default strategy for backward compatibility.
- Define backfill strategy for new required fields.
- Include rollback notes for each migration change.
- Add indexes for new query-critical paths.
