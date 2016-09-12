package main

import (
	"fmt"
	"github.com/qmsk/e2/client"
)

type ScreenShow struct {
	ID int `long:"screen-id" required:"true"`
}

func init() {
	parser.AddCommand("screen-show", "Show screen content", "", &ScreenShow{})
}

func (cmd *ScreenShow) Execute(args []string) error {
	c, err := options.ClientOptions.JSONClient()
	if err != nil {
		return err
	}

	sourceMap := make(map[int]client.Source)

	if sourceList, err := c.ListSources(); err != nil {
		return err
	} else {
		for _, source := range sourceList {
			sourceMap[source.ID] = source
		}
	}

	if content, err := c.ListContent(cmd.ID); err != nil {
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
			fmt.Printf("\tLayer %d: pgm=%v pvw=%d\n", layer.ID, layer.PgmMode, layer.PvwMode)

			if layer.LastSrcIdx < 0 {

			} else if source, found := sourceMap[layer.LastSrcIdx]; !found {
				fmt.Printf("\t\tSource %d: \n", layer.LastSrcIdx)
			} else {
				fmt.Printf("\t\tSource %d: name=%v\n", layer.LastSrcIdx, source.Name)
			}
		}
	}

	return nil
}
