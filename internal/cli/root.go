package cli

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "path/filepath"

    "github.com/heyjasonn/ai-workflow/internal/agent"
    "github.com/heyjasonn/ai-workflow/internal/llm"
    "github.com/heyjasonn/ai-workflow/internal/repository"
    "github.com/heyjasonn/ai-workflow/internal/workflow"
    "github.com/spf13/cobra"
)

type App struct {
    Root *cobra.Command
}

func NewApp() (*App, error) {
    root := &cobra.Command{Use: "workflow", Short: "AI workflow engine"}

    app := &App{Root: root}
    engine, err := buildEngine()
    if err != nil {
        return nil, err
    }

    root.AddCommand(createCmd(engine))
    root.AddCommand(showCmd(engine))
    root.AddCommand(listCmd(engine))
    root.AddCommand(statusCmd(engine))
    root.AddCommand(runCmd(engine))
    root.AddCommand(approveCmd(engine))
    root.AddCommand(rejectCmd(engine))
    root.AddCommand(applyCmd(engine))

    return app, nil
}

func buildEngine() (*workflow.Engine, error) {
    repo, err := buildRepository()
    if err != nil {
        return nil, err
    }

    registry := buildAgentRegistry()

    rootDir, err := os.Getwd()
    if err != nil {
        return nil, err
    }

    return &workflow.Engine{
        Repo:       repo,
        Agents:     registry,
        RootDir:    rootDir,
        TestRunner: workflow.GoTestRunner{RootDir: rootDir},
    }, nil
}

func buildRepository() (workflow.Repository, error) {
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        return repository.NewMemoryRepository(), nil
    }

    db, err := repository.OpenGormPostgres(dsn)
    if err != nil {
        return nil, err
    }
    migrationsDir := filepath.Join(mustGetwd(), "migrations")
    if err := repository.ApplyMigrations(context.Background(), db, migrationsDir); err != nil {
        return nil, err
    }
    repo := repository.NewGormRepository(db)
    return repo, nil
}

func buildAgentRegistry() agent.Registry {
    provider := os.Getenv("LLM_PROVIDER")
    if provider == "stub" || provider == "" {
        return agent.Registry{
            Research:  agent.StaticAgent{Output: map[string]any{"competitor_features": []string{}, "design_patterns": []string{}, "recommendations": []string{}}},
            Planning:  agent.StaticAgent{Output: map[string]any{"new_packages": []string{}, "modified_files": []string{}, "interfaces": []string{}, "db_changes": []string{}, "risks": []string{}}},
            TaskSplit: agent.StaticAgent{Output: []agent.TaskItem{}},
            Executor:  agent.StaticExecutor{},
        }
    }

    client := llm.OpenAIClient{APIKey: os.Getenv("OPENAI_API_KEY")}
    return agent.Registry{
        Research:  agent.ResearchAgent{LLM: client},
        Planning:  agent.PlanningAgent{LLM: client},
        TaskSplit: agent.TaskSplitAgent{LLM: client},
        Executor:  agent.LLMExecutorAgent{LLM: client},
    }
}

func mustGetwd() string {
    dir, err := os.Getwd()
    if err != nil {
        return "."
    }
    return dir
}

func createCmd(engine *workflow.Engine) *cobra.Command {
    cmd := &cobra.Command{
        Use:  "create",
        Short: "Create workflow",
        Args: cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            wf, err := engine.CreateWorkflow(context.Background(), args[0])
            if err != nil {
                return err
            }
            return printJSON(wf)
        },
    }
    return cmd
}

func showCmd(engine *workflow.Engine) *cobra.Command {
    cmd := &cobra.Command{
        Use:  "show",
        Short: "Show workflow",
        Args: cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            wf, err := engine.Repo.Get(context.Background(), args[0])
            if err != nil {
                return err
            }
            return printJSON(wf)
        },
    }
    return cmd
}

func approveCmd(engine *workflow.Engine) *cobra.Command {
    var step string
    cmd := &cobra.Command{
        Use:  "approve",
        Short: "Approve a step",
        Args: cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            if step == "" {
                return errors.New("missing --step")
            }
            wf, err := engine.ApproveStep(context.Background(), args[0], workflow.ApprovalStep(step))
            if err != nil {
                return err
            }
            return printJSON(wf)
        },
    }
    cmd.Flags().StringVar(&step, "step", "", "step to approve")
    return cmd
}

func rejectCmd(engine *workflow.Engine) *cobra.Command {
    var step string
    cmd := &cobra.Command{
        Use:  "reject",
        Short: "Reject a step",
        Args: cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            if step == "" {
                return errors.New("missing --step")
            }
            wf, err := engine.RejectStep(context.Background(), args[0], workflow.ApprovalStep(step))
            if err != nil {
                return err
            }
            return printJSON(wf)
        },
    }
    cmd.Flags().StringVar(&step, "step", "", "step to reject")
    return cmd
}

func applyCmd(engine *workflow.Engine) *cobra.Command {
    cmd := &cobra.Command{
        Use:  "apply",
        Short: "Apply latest patch",
        Args: cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            wf, err := engine.ApplyPatch(context.Background(), args[0])
            if err != nil {
                return err
            }
            return printJSON(wf)
        },
    }
    return cmd
}

func runCmd(engine *workflow.Engine) *cobra.Command {
    var step string
    cmd := &cobra.Command{
        Use:   "run",
        Short: "Run a previously approved step",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            if step == "" {
                return errors.New("missing --step")
            }
            wf, err := engine.RunStep(context.Background(), args[0], workflow.ApprovalStep(step))
            if err != nil {
                return err
            }
            return printJSON(wf)
        },
    }
    cmd.Flags().StringVar(&step, "step", "", "step to run")
    return cmd
}

func listCmd(engine *workflow.Engine) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "list",
        Short: "List workflows",
        Args:  cobra.MaximumNArgs(0),
        RunE: func(cmd *cobra.Command, args []string) error {
            items, err := engine.ListWorkflows(context.Background())
            if err != nil {
                return err
            }
            return printJSON(items)
        },
    }
    return cmd
}

func statusCmd(engine *workflow.Engine) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "status",
        Short: "Show workflow status and next step",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            status, err := engine.GetStatus(context.Background(), args[0])
            if err != nil {
                return err
            }
            return printJSON(status)
        },
    }
    return cmd
}

func printJSON(v any) error {
    data, err := json.MarshalIndent(v, "", "  ")
    if err != nil {
        return err
    }
    fmt.Println(string(data))
    return nil
}
