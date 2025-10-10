//go:build cli

package cli

import (
    "fmt"
    "github.com/spf13/cobra"
)

var (
    Version   = "0.0.1"
    GitCommit = "dev"
)

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Print version information",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Printf("gosper %s (%s)\n", Version, GitCommit)
    },
}
