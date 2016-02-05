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
            if aux, err := client.AuxDestination(auxDest.ID); err != nil {
                fmt.Printf("\tAux %d: error=%v\n", auxDest.ID, err)
            } else {
                fmt.Printf("\tAux %d: name=%v\n", auxDest.ID, aux.Name)
            }
        }

        fmt.Printf("Screen Destinations: %d\n", len(preset.ScreenDest))
        for _, screenDest := range preset.ScreenDest {
            if screen, err := client.ScreenDestination(screenDest.ID); err != nil {
                fmt.Printf("\tScreen %d: error=%v\n", screenDest.ID, err)
            } else {
                fmt.Printf("\tScreen %d: name=%v\n", screenDest.ID, screen.Name)
            }
        }
    }

    return nil
}
