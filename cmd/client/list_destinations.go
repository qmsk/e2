package main

import (
    "fmt"
)

type ListDestinations struct {

}

func init() {
    parser.AddCommand("list-destinations", "List destinations", "", &ListDestinations{})
}

func (cmd *ListDestinations) Execute(args []string) error {
    if client, err := options.ClientOptions.Client(); err != nil {
        return err
    } else if listDestinations, err := client.ListDestinations(); err != nil {
        return err
    } else {
        fmt.Printf("%8s %-8s %s\n", "Type", "ID", "Name")

        for _, screenDestination := range listDestinations.ScreenDestinations {
            fmt.Printf("%8s %-8d %s\n", "Screen", screenDestination.ID, screenDestination.Name)
        }
        for _, auxDestination := range listDestinations.AuxDestinations {
            fmt.Printf("%8s %-8d %s\n", "Aux", auxDestination.ID, auxDestination.Name)
        }
    }

    return nil
}
