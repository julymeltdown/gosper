//go:build cli

package main

import (
    "log"
    "gosper/internal/adapter/inbound/cli"
)

func main() {
    if err := cli.Execute(); err != nil {
        log.Fatalf("gosper: %v", err)
    }
}
