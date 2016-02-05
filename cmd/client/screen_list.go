package main

import (
    "fmt"
)

type ScreenList struct {

}

func init() {
    parser.AddCommand("screen-list", "List Screen destinations", "", &ScreenList{})
}

func (cmd *ScreenList) Execute(args []string) error {
    if client, err := options.ClientOptions.Client(); err != nil {
        return err
    } else if screenDestinations, err := client.ScreenDestinations(); err != nil {
        return err
    } else {
        fmt.Printf("%-8s %s\n", "Screen", "Name")

        for _, screenDestination := range screenDestinations {
            fmt.Printf("%-8d %s\n", screenDestination.ID, screenDestination.Name)
        }
    }

    return nil
}
