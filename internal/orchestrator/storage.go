package orchestrator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type FileArtifactStore struct {
	BaseDir string
}

func (s FileArtifactStore) Save(runID string, step Step, artifacts StepArtifacts) (string, error) {
	if runID == "" {
		return "", fmt.Errorf("runID is required")
	}
	base := s.BaseDir
	if base == "" {
		base = "runs"
	}
	dir := filepath.Join(base, runID, string(step))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	promptMD := renderPromptMarkdown(step, artifacts.Attempt, artifacts.Prompt)
	if err := os.WriteFile(filepath.Join(dir, "prompt.md"), []byte(promptMD), 0o644); err != nil {
		return "", err
	}
	rawJSON, err := normalizeJSON(artifacts.RawOutput)
	if err != nil {
		rawJSON = MarshalPretty(map[string]any{
			"raw_output_text": artifacts.RawOutput,
			"parse_error":     err.Error(),
		})
	}
	if err := os.WriteFile(filepath.Join(dir, "raw_output.json"), rawJSON, 0o644); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dir, "validation.json"), MarshalPretty(artifacts.Validation), 0o644); err != nil {
		return "", err
	}
	meta := map[string]any{"attempt": artifacts.Attempt}
	if err := os.WriteFile(filepath.Join(dir, "metadata.json"), MarshalPretty(meta), 0o644); err != nil {
		return "", err
	}
	return dir, nil
}

func renderPromptMarkdown(step Step, attempt int, prompt string) string {
	return fmt.Sprintf(
		"# Prompt Artifact\n\n- Step: `%s`\n- Attempt: `%d`\n\n## Prompt\n\n~~~text\n%s\n~~~\n",
		step,
		attempt,
		prompt,
	)
}

func normalizeJSON(raw string) ([]byte, error) {
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return nil, err
	}
	return MarshalPretty(v), nil
}
