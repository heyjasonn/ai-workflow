package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnv(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, ".env")
	content := "# comment\nLLM_PROVIDER=codex-cli\nCODEX_CLI_COMMAND='codex'\nexport CODEX_CLI_ARGS=exec --json\n"
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	t.Setenv("LLM_PROVIDER", "")
	t.Setenv("CODEX_CLI_COMMAND", "")
	t.Setenv("CODEX_CLI_ARGS", "")

	if err := loadDotEnv(p); err != nil {
		t.Fatalf("loadDotEnv error: %v", err)
	}

	if got := os.Getenv("LLM_PROVIDER"); got != "codex-cli" {
		t.Fatalf("LLM_PROVIDER mismatch: %q", got)
	}
	if got := os.Getenv("CODEX_CLI_COMMAND"); got != "codex" {
		t.Fatalf("CODEX_CLI_COMMAND mismatch: %q", got)
	}
	if got := os.Getenv("CODEX_CLI_ARGS"); got != "exec --json" {
		t.Fatalf("CODEX_CLI_ARGS mismatch: %q", got)
	}
}

func TestLoadDotEnv_NoOverwrite(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, ".env")
	if err := os.WriteFile(p, []byte("LLM_PROVIDER=codex-cli\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	t.Setenv("LLM_PROVIDER", "cursor-cli")
	if err := loadDotEnv(p); err != nil {
		t.Fatalf("loadDotEnv error: %v", err)
	}
	if got := os.Getenv("LLM_PROVIDER"); got != "cursor-cli" {
		t.Fatalf("expected existing env preserved, got %q", got)
	}
}
