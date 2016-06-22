package main

import (
	"encoding/json"
	"log"
	"os"
)

type Listen struct {
	JSON bool `long:"json" help:"Output JSON"`
}

func init() {
	parser.AddCommand("listen", "Listen XML packets", "", &Listen{})
}

func (cmd *Listen) Execute(args []string) error {
	if xmlClient, err := options.ClientOptions.XMLClient(); err != nil {
		return err
	} else if listenChan, err := xmlClient.Listen(); err != nil {
		return err
	} else {
		log.Printf("Listen...\n")

		for system := range listenChan {
			if cmd.JSON {
				if err := json.NewEncoder(os.Stdout).Encode(system); err != nil {
					log.Printf("JSON Encode: %v\n", err)
				}
			} else {
				system.Print(os.Stdout)
			}
		}
	}

	return nil
}
