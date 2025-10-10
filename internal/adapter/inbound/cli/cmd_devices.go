//go:build cli

package cli

import (
    "fmt"
    "strings"
    "github.com/spf13/cobra"
    "gosper/internal/adapter/outbound/audio"
    "gosper/internal/infrastructure/config"
    "gosper/internal/usecase"
)

var devicesCmd = &cobra.Command{
    Use:   "devices",
    Short: "Manage audio input devices",
}

var devicesListCmd = &cobra.Command{
    Use:   "list",
    Short: "List available input devices",
    RunE: func(cmd *cobra.Command, args []string) error {
        inp := audio.NewInput()
        uc := &usecase.ListDevices{ Audio: inp }
        devs, err := uc.Execute(cmd.Context())
        if err != nil { return err }
        if len(devs) == 0 { fmt.Println("No input devices found"); return nil }
        cur, _ := config.LoadFile(config.DefaultPath())
        for _, d := range devs {
            mark := ""
            if cur.LastDeviceID != "" && (d.ID == cur.LastDeviceID || strings.EqualFold(d.Name, cur.LastDeviceID)) {
                mark = "*"
            }
            fmt.Printf("%s\t%s\t%s\n", d.ID, d.Name, mark)
        }
        return nil
    },
}

var devicesSelectCmd = &cobra.Command{
    Use:   "select <id>",
    Short: "Select default input device for future recordings",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        sel := args[0]
        inp := audio.NewInput()
        uc := &usecase.ListDevices{ Audio: inp }
        devs, err := uc.Execute(cmd.Context())
        if err != nil { return err }
        // resolve selection robustly (exact id, name, prefix, substring, fuzzy)
        store := audio.ResolveDeviceID(devs, sel)
        if store == "" { store = sel }
        c, _ := config.LoadFile(config.DefaultPath())
        c.LastDeviceID = store
        return config.SaveFile(config.DefaultPath(), c)
    },
}

func init() {
    devicesCmd.AddCommand(devicesListCmd)
    devicesCmd.AddCommand(devicesSelectCmd)
    rootCmd.AddCommand(devicesCmd)
}
