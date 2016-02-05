package main

import (
    "github.com/qmsk/e2/client"
    "fmt"
)

type ShowPreset struct {
    ID      int     `long:"preset-id" required:"true"`
}

func init() {
    parser.AddCommand("show-preset", "Show preset destinations", "", &ShowPreset{})
}

func (cmd *ShowPreset) Execute(args []string) error {
    c, err := options.ClientOptions.Client()
    if err != nil {
        return err
    }

    auxDestinations := make(map[int]client.AuxDestination)
    screenDestinations := make(map[int]client.ScreenDestination)

    if listDestinations, err := c.ListDestinations(); err != nil {
        return err
    } else {
        for _, auxDestination := range listDestinations.AuxDestinations {
            auxDestinations[auxDestination.ID] = auxDestination
        }

        for _, screenDestination := range listDestinations.ScreenDestinations {
            screenDestinations[screenDestination.ID] = screenDestination
        }
    }

    if preset, err := c.ListDestinationsForPreset(cmd.ID); err != nil {
        return err
    } else {
        fmt.Printf("Preset %d: %s\n", preset.ID, preset.Name)

        fmt.Printf("Aux Destinations: %d\n", len(preset.AuxDest))
        for _, auxDest := range preset.AuxDest {
            if aux, found := auxDestinations[auxDest.ID]; !found {
                fmt.Printf("\tAux %d: \n", auxDest.ID)
            } else {
                fmt.Printf("\tAux %d: name=%v\n", auxDest.ID, aux.Name)
            }
        }

        fmt.Printf("Screen Destinations: %d\n", len(preset.ScreenDest))
        for _, screenDest := range preset.ScreenDest {
            if screen, found := screenDestinations[screenDest.ID]; !found {
                fmt.Printf("\tScreen %d: \n", screenDest.ID)
            } else {
                fmt.Printf("\tScreen %d: name=%v\n", screenDest.ID, screen.Name)
            }
        }
    }

    return nil
}
