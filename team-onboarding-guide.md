# Team Onboarding Guide

1. Read `vision.md`, `scope-v1.md`, `non-goals.md`.
2. Read agent specs in `agents/`.
3. Read standards in `standards/`.
4. Run orchestration MVP with `workflow orchestrate`.
   - Stub mode (default): `workflow orchestrate "your requirement"`
   - Codex CLI mode: `LLM_PROVIDER=codex-cli CODEX_CLI_COMMAND=codex workflow orchestrate "your requirement"`
   - Cursor CLI mode: `LLM_PROVIDER=cursor-cli CURSOR_CLI_COMMAND=cursor workflow orchestrate "your requirement"`
   - Claude CLI mode: `LLM_PROVIDER=claude-cli CLAUDE_CLI_COMMAND=claude workflow orchestrate "your requirement"`
   - Optional args: `*_CLI_ARGS` (space-separated). For Codex, empty args default to `exec -` (non-interactive).
5. Review outputs under `runs/<run-id>/`.
6. Score the run using `evaluation-rubric.md`.
