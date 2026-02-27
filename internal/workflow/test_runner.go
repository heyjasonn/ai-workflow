package workflow

import (
    "context"
    "os/exec"
    "strings"
)

type GoTestRunner struct {
    RootDir string
}

func (g GoTestRunner) Run(ctx context.Context) (TestReport, error) {
    report := TestReport{}

    testOut, testErr := runCmd(ctx, g.RootDir, "go", "test", "./...")
    vetOut, vetErr := runCmd(ctx, g.RootDir, "go", "vet", "./...")

    var errorsList []string
    if testErr != nil {
        errorsList = append(errorsList, strings.TrimSpace(testOut))
    }
    if vetErr != nil {
        errorsList = append(errorsList, strings.TrimSpace(vetOut))
    }

    if testErr != nil || vetErr != nil {
        report.Status = "fail"
        report.Errors = errorsList
    } else {
        report.Status = "pass"
    }

    return report, nil
}

func runCmd(ctx context.Context, dir string, name string, args ...string) (string, error) {
    cmd := exec.CommandContext(ctx, name, args...)
    cmd.Dir = dir
    out, err := cmd.CombinedOutput()
    return string(out), err
}
