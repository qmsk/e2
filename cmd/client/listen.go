package main

import (
    "fmt"
    "log"
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
            fmt.Printf("System:\n")

            for _, auxDest := range system.DestMgr.AuxDestCol.List() {
                fmt.Printf("Aux %d: %v\n\tISActive: %v\n\tPvwLastSrcIndex=%d PgmLastSrcIndex=%d\n", auxDest.ID, auxDest.Name, auxDest.IsActive, auxDest.PvwLastSrcIndex, auxDest.PgmLastSrcIndex)
            }

            for _, screenDest := range system.DestMgr.ScreenDestCol.List() {
                fmt.Printf("Screen %d: %v\n\tIsActive: %v\n\tHSize=%v VSize=%v\n", screenDest.ID, screenDest.Name, screenDest.IsActive, screenDest.HSize, screenDest.VSize)

                for _, layer := range screenDest.LayerCollection.List() {
                    fmt.Printf("\tLayer %d: %v\n\t\tIsActive=%v\n\t\tPgmMode=%v PvwMode=%v\n\t\tLastSrcIdx=%d\n", layer.ID, layer.Name, layer.IsActive, layer.PgmMode, layer.PvwMode, layer.LastSrcIdx)
                }
            }

            fmt.Printf("Presets:\n\tLastRecall: %v\n", system.PresetMgr.LastRecall)

            for _, preset := range system.PresetMgr.Preset.List() {
                fmt.Printf("Preset %d: %v\n\tLockMode: %v\n\tSno: %v\n", preset.ID, preset.Name, preset.LockMode, preset.Sno)
            }
        }
    }

    return nil
}
