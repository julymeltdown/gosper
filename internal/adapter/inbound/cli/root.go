//go:build cli

package cli

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "gosper",
    Short: "Gosper – CLI for speech-to-text via whisper.cpp",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Default behavior: show help
        return cmd.Help()
    },
}

// Execute runs the root command.
func Execute() error {
    ctx, cancel := signalContext()
    defer cancel()
    // Pass ctx to subcommands via context if needed
    rootCmd.SetContext(ctx)
    // attach subcommands
    rootCmd.AddCommand(versionCmd)
    return rootCmd.Execute()
}

func signalContext() (context.Context, context.CancelFunc) {
    ctx, cancel := context.WithCancel(context.Background())
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-c
        fmt.Fprintln(os.Stderr, "\nReceived interrupt, shutting down...")
        cancel()
    }()
    return ctx, cancel
}
