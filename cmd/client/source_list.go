package main

import (
    "fmt"
)

type SourceList struct {

}

func init() {
    parser.AddCommand("source-list", "List sources", "", &SourceList{})
}

func (cmd *SourceList) Execute(args []string) error {
    if client, err := options.ClientOptions.Client(); err != nil {
        return err
    } else if sourceList, err := client.ListSources(); err != nil {
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
