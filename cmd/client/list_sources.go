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
    } else if sourceList, err := client.Sources(); err != nil {
        return err
    } else {
        fmt.Printf("%8s %-8s %-20s %s\n", "ID", "Type", "Name", "Status")

        for _, source := range sourceList {
            status := ""

            if source.InputCfgIndex >= 0 {
                status = fmt.Sprintf("size=%4dx%-4d video=%-8v", source.HSize, source.VSize, source.InputVideoStatus)
            }

            fmt.Printf("%8d %-8v %-20s %s\n", source.ID, source.Type, source.Name, status)
        }
    }

    return nil
}
