package orchestrator

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeRunner struct {
	gotCommand string
	gotArgs    []string
	gotStdin   string
	stdout     string
	err        error
}

func (f *fakeRunner) Run(ctx context.Context, command string, args []string, stdin string) (string, error) {
	_ = ctx
	f.gotCommand = command
	f.gotArgs = append([]string(nil), args...)
	f.gotStdin = stdin
	if f.err != nil {
		return "", f.err
	}
	return f.stdout, nil
}

func TestCLICommandAdapter_Run(t *testing.T) {
	r := &fakeRunner{stdout: `{"ok":true}`}
	a := CLICommandAdapter{
		Name:    "codex",
		Command: "codex",
		Args:    []string{"exec", "--json"},
		Runner:  r,
	}

	out, err := a.Run(context.Background(), string(StepResearcher), "hello")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if out != `{"ok":true}` {
		t.Fatalf("unexpected output: %s", out)
	}
	if r.gotCommand != "codex" {
		t.Fatalf("unexpected command: %s", r.gotCommand)
	}
	if len(r.gotArgs) != 2 || r.gotArgs[0] != "exec" {
		t.Fatalf("unexpected args: %v", r.gotArgs)
	}
	if !strings.Contains(r.gotStdin, "Agent: researcher") {
		t.Fatalf("prompt missing agent context: %s", r.gotStdin)
	}
}

func TestCLICommandAdapter_RunError(t *testing.T) {
	a := CLICommandAdapter{
		Name:    "cursor",
		Command: "cursor",
		Runner:  &fakeRunner{err: errors.New("boom")},
	}
	_, err := a.Run(context.Background(), string(StepPlanner), "hello")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestNewCodexCLIAdapterFromEnv(t *testing.T) {
	t.Setenv("CODEX_CLI_COMMAND", "sh")
	t.Setenv("CODEX_CLI_ARGS", "-c cat")

	provider, err := NewCodexCLIAdapterFromEnv()
	if err != nil {
		t.Fatalf("constructor failed: %v", err)
	}
	out, err := provider.Run(context.Background(), string(StepTester), "{}")
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if !strings.Contains(out, "Agent: tester") {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestDefaultCLIArgs_Codex(t *testing.T) {
	got := defaultCLIArgs("codex", nil)
	if len(got) != 2 || got[0] != "exec" || got[1] != "-" {
		t.Fatalf("unexpected default args: %v", got)
	}
}

func TestNewCursorCLIAdapterFromEnv_CommandNotFound(t *testing.T) {
	t.Setenv("CURSOR_CLI_COMMAND", "/tmp/definitely-not-installed-binary")
	_, err := NewCursorCLIAdapterFromEnv()
	if err == nil {
		t.Fatalf("expected constructor error")
	}
}

func TestNewClaudeCLIAdapterFromEnv_CommandNotFound(t *testing.T) {
	t.Setenv("CLAUDE_CLI_COMMAND", "/tmp/definitely-not-installed-binary")
	_, err := NewClaudeCLIAdapterFromEnv()
	if err == nil {
		t.Fatalf("expected constructor error")
	}
}
