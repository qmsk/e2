package main

import (
    "log"
    "os"
)

type Listen struct {

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
            system.Print(os.Stdout)
        }
    }

    return nil
}
