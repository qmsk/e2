package main

import (
    "fmt"
)

type PresetList struct {

}

func init() {
    parser.AddCommand("preset-list", "List presets", "", &PresetList{})
}

func (cmd *PresetList) Execute(args []string) error {
    if client, err := options.ClientOptions.Client(); err != nil {
        return err
    } else if presets, err := client.ListPresets(); err != nil {
        return err
    } else {
        fmt.Printf("#%-7d %-8s     %s\n", len(presets), "Seq", "Name")

        for _, preset := range presets {
            fmt.Printf("%-8d %-8.2f     %s\n", preset.ID, preset.Sno, preset.Name)
        }
    }

    return nil
}
