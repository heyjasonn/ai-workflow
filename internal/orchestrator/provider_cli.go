package orchestrator

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type CommandRunner interface {
	Run(ctx context.Context, command string, args []string, stdin string) (string, error)
}

type ExecCommandRunner struct{}

func (r ExecCommandRunner) Run(ctx context.Context, command string, args []string, stdin string) (string, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Stdin = strings.NewReader(stdin)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("%s %v failed: %s", command, args, msg)
	}
	return strings.TrimSpace(stdout.String()), nil
}

type CLICommandAdapter struct {
	Name    string
	Command string
	Args    []string
	Runner  CommandRunner
}

func (a CLICommandAdapter) Run(ctx context.Context, agent string, prompt string) (string, error) {
	if strings.TrimSpace(a.Command) == "" {
		return "", fmt.Errorf("command is not configured for %s", a.Name)
	}
	runner := a.Runner
	if runner == nil {
		runner = ExecCommandRunner{}
	}
	input := fmt.Sprintf("Agent: %s\nReturn strict JSON only.\n%s", agent, prompt)
	return runner.Run(ctx, a.Command, a.Args, input)
}

func NewCodexCLIAdapterFromEnv() (ProviderAdapter, error) {
	return newCLIAdapterFromEnv("codex", "CODEX_CLI_COMMAND", "CODEX_CLI_ARGS")
}

func NewCursorCLIAdapterFromEnv() (ProviderAdapter, error) {
	return newCLIAdapterFromEnv("cursor", "CURSOR_CLI_COMMAND", "CURSOR_CLI_ARGS")
}

func NewClaudeCLIAdapterFromEnv() (ProviderAdapter, error) {
	return newCLIAdapterFromEnv("claude", "CLAUDE_CLI_COMMAND", "CLAUDE_CLI_ARGS")
}

func newCLIAdapterFromEnv(name string, commandKey string, argsKey string) (ProviderAdapter, error) {
	command := os.Getenv(commandKey)
	if strings.TrimSpace(command) == "" {
		command = name
	}
	args := strings.Fields(os.Getenv(argsKey))
	args = defaultCLIArgs(name, args)
	resolved, err := resolveCLICommand(name, command)
	if err != nil {
		return nil, fmt.Errorf("%s command not found. Set %s to your installed binary path", name, commandKey)
	}
	return CLICommandAdapter{
		Name:    name,
		Command: resolved,
		Args:    args,
	}, nil
}

func defaultCLIArgs(name string, args []string) []string {
	if len(args) > 0 {
		return args
	}
	switch name {
	case "codex":
		return []string{"exec", "-"}
	default:
		return args
	}
}

func resolveCLICommand(name string, command string) (string, error) {
	if path, err := exec.LookPath(command); err == nil {
		return path, nil
	}
	for _, candidate := range fallbackCommandPaths(name, command) {
		if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("command not found")
}

func fallbackCommandPaths(name string, command string) []string {
	base := filepath.Base(command)
	if base != command {
		// Command already looks like a path; avoid overriding user intent.
		return nil
	}
	switch name {
	case "codex":
		return []string{
			"/Applications/Codex.app/Contents/Resources/codex",
		}
	case "cursor":
		return []string{
			"/Applications/Cursor.app/Contents/MacOS/Cursor",
		}
	case "claude":
		return []string{
			"/Applications/Claude.app/Contents/MacOS/Claude",
		}
	default:
		return nil
	}
}
