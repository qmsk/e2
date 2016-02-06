package main

import (
    "fmt"
    "log"
)

type Events struct {

}

func init() {
    parser.AddCommand("events", "Listen events", "", &Events{})
}

func (cmd *Events) Execute(args []string) error {
    if client, err := options.ClientOptions.Client(); err != nil {
        return err
    } else if eventChan, err := client.ListenEvents(); err != nil {
        return err
    } else {
        log.Printf("Listen events...\n")

        for event := range eventChan {
            fmt.Printf("%v\n", event)
        }
    }

    return nil
}
