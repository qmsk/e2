package main

import (
    "fmt"
)

type ListSources struct {

}

func init() {
    parser.AddCommand("list-sources", "List sources", "", &ListSources{})
}

func (cmd *ListSources) Execute(args []string) error {
    if client, err := options.ClientOptions.Client(); err != nil {
        return err
    } else if sourceList, err := client.ListSources(); err != nil {
        return err
    } else {
        fmt.Printf("%8s %-8s %s\n", "Type", "ID", "Name")

        for _, source := range sourceList {
            fmt.Printf("%8v %-8d %s\n", source.Type, source.ID, source.Name)
        }
    }

    return nil
}
