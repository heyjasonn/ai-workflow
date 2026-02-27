package main

import (
    "fmt"
    "os"

    "github.com/heyjasonn/ai-workflow/internal/cli"
)

func main() {
    app, err := cli.NewApp()
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
    if err := app.Root.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
