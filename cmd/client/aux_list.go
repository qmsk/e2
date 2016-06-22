package main

import (
	"fmt"
)

type AuxList struct {
}

func init() {
	parser.AddCommand("aux-list", "List Aux destinations", "", &AuxList{})
}

func (cmd *AuxList) Execute(args []string) error {
	if client, err := options.ClientOptions.Client(); err != nil {
		return err
	} else if auxDestinations, err := client.ListAuxDestinations(); err != nil {
		return err
	} else {
		fmt.Printf("%-8s %s\n", "Aux", "Name")

		for _, auxDestination := range auxDestinations {
			fmt.Printf("%-8d %s\n", auxDestination.ID, auxDestination.Name)
		}
	}

	return nil
}
