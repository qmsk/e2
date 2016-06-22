package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Listen struct {
	JSON bool `long:"json" help:"Output JSON"`
}

func init() {
	parser.AddCommand("listen", "Listen for state updates", "", &Listen{})
}

func (cmd *Listen) Execute(args []string) error {
	if client, err := options.ClientOptions.Client(); err != nil {
		return err
	} else {
		for {
			if dcpDevice, err := client.Read(); err != nil {
				return fmt.Errorf("dcp:Client.Read: %v\n", err)
			} else {
				if cmd.JSON {
					if err := json.NewEncoder(os.Stdout).Encode(dcpDevice); err != nil {
						return fmt.Errorf("JSON Encode: %v\n", err)
					}
				} else {
					dcpDevice.Print(os.Stdout)
				}
			}
		}
	}
}
