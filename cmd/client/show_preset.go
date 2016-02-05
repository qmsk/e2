package main

import (
    "fmt"
)

type ShowPreset struct {
    ID      int     `long:"preset-id" required:"true"`
}

func init() {
    parser.AddCommand("show-preset", "Show preset destinations", "", &ShowPreset{})
}

func (cmd *ShowPreset) Execute(args []string) error {
    if client, err := options.ClientOptions.Client(); err != nil {
        return err
    } else if preset, err := client.ListDestinationsForPreset(cmd.ID); err != nil {
        return err
    } else {
        fmt.Printf("Preset %d: %s\n", preset.ID, preset.Name)

        fmt.Printf("Aux Destinations: %d\n", len(preset.AuxDest))
        for _, auxDest := range preset.AuxDest {
            fmt.Printf("\tAux %d\n", auxDest.ID)
        }

        fmt.Printf("Screen Destinations: %d\n", len(preset.ScreenDest))
        for _, screenDest := range preset.ScreenDest {
            fmt.Printf("\tScreen %d\n", screenDest.ID)
        }
    }

    return nil
}
