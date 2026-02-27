package diff

import (
    "bufio"
    "bytes"
    "context"
    "errors"
    "os/exec"
    "path/filepath"
    "strings"
)

type FileChange struct {
    Path      string
    IsNewFile bool
}

func ExtractFilePaths(patch string) []FileChange {
    var files []FileChange
    scanner := bufio.NewScanner(strings.NewReader(patch))
    var oldPath string
    for scanner.Scan() {
        line := scanner.Text()
        if strings.HasPrefix(line, "--- ") {
            oldPath = strings.TrimSpace(strings.TrimPrefix(line, "--- "))
        }
        if strings.HasPrefix(line, "+++ ") {
            newPath := strings.TrimSpace(strings.TrimPrefix(line, "+++ "))
            if newPath == "/dev/null" {
                continue
            }
            path := strings.TrimPrefix(newPath, "b/")
            isNew := oldPath == "/dev/null"
            files = append(files, FileChange{Path: path, IsNewFile: isNew})
        }
    }
    return files
}

func ApplyPatch(ctx context.Context, rootDir string, patch string) error {
    if strings.TrimSpace(patch) == "" {
        return errors.New("empty patch")
    }
    cmd := exec.CommandContext(ctx, "git", "apply", "-")
    cmd.Dir = filepath.Clean(rootDir)
    cmd.Stdin = bytes.NewBufferString(patch)
    output, err := cmd.CombinedOutput()
    if err != nil {
        return errors.New(string(output))
    }
    return nil
}
