package main

import (
    "fmt"
)

type ShowScreen struct {
    ID      int     `long:"screen-id" required:"true"`
}

func init() {
    parser.AddCommand("show-screen", "Show screen content", "", &ShowScreen{})
}

func (cmd *ShowScreen) Execute(args []string) error {
    if client, err := options.ClientOptions.Client(); err != nil {
        return err
    } else if content, err := client.ListContent(cmd.ID); err != nil {
        return err
    } else {
        fmt.Printf("Screen %d: %s\n", content.ID, content.Name)

        fmt.Printf("BG Layers: %d\n", len(content.BGLayers))
        for _, bgLayer := range content.BGLayers {
            if bgLayer.ShowMatte != 0 {
                fmt.Printf("\tBG Layer %d: matte color=%v\n", bgLayer.ID, bgLayer.Color)
            } else if bgLayer.LastBGSourceIndex >= 0 {
                fmt.Printf("\tBG Layer %d: source id=%v\n", bgLayer.ID, bgLayer.LastBGSourceIndex)
            } else {
                fmt.Printf("\tBG Layer %d: unknown\n", bgLayer.ID)
            }
        }

        fmt.Printf("Layers: %d\n", len(content.Layers))
        for _, layer := range content.Layers {
            fmt.Printf("\tLayer %d: pgm=%v pvw=%d source=%d\n", layer.ID, layer.PgmMode, layer.PvwMode, layer.LastSrcIdx)
        }
    }

    return nil
}
