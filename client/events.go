package client

import (
    "fmt"
    "log"
)

type PresetRecallEvent struct {
    PresetID        int
}

func (event PresetRecallEvent) String() string {
    return fmt.Sprintf("preset recall preset=%d", event.PresetID)
}

type AuxActiveEvent struct {
    AuxID           int
    Active          bool
}

func (event AuxActiveEvent) String() string {
    return fmt.Sprintf("aux=%d active=%v", event.AuxID, event.Active)
}

type AuxPreviewEvent struct {
    AuxID           int
    SourceID        int

    Source          Source
}

func (event AuxPreviewEvent) String() string {
    return fmt.Sprintf("aux=%d preview source=%d", event.AuxID, event.SourceID)
}

type AuxProgramEvent struct {
    AuxID           int
    SourceID        int

    Source          Source
}

func (event AuxProgramEvent) String() string {
    return fmt.Sprintf("aux=%d program source=%d", event.AuxID, event.SourceID)
}

type Event interface {}

func (client *Client) listenEvents(xmlClient *xmlClient, eventChan chan Event) {
    for xmlPacket := range xmlClient.listenChan {
        if xmlPacket.DestMgr != nil {
            for _, auxDest := range xmlPacket.DestMgr.AuxDest {
                log.Printf("Client.listenEvents: DestMgr.AuxDest %#v\n", auxDest)

                if auxDest.IsActive != nil {
                    eventChan <- AuxActiveEvent{auxDest.ID, *auxDest.IsActive > 0}
                }
                if auxDest.Source != nil {
                    source := *auxDest.Source

                    if auxDest.PvwLastSrcIndex != nil {
                        source.ID = *auxDest.PvwLastSrcIndex
                        eventChan <- AuxPreviewEvent{auxDest.ID, *auxDest.PvwLastSrcIndex, source}
                    }
                    if auxDest.PgmLastSrcIndex != nil {
                        source.ID = *auxDest.PgmLastSrcIndex
                        eventChan <- AuxProgramEvent{auxDest.ID, *auxDest.PgmLastSrcIndex, source}
                    }
                }
            }

            for _, screenDest := range xmlPacket.DestMgr.ScreenDest {
                log.Printf("Client.listenEvents: DestMgr.ScreenDest %#v\n", screenDest)
            }
        }

        if xmlPacket.PresetMgr != nil {
            log.Printf("Client.listenEvents: PresetMgr %#v\n", xmlPacket.PresetMgr)

            if lastRecall := xmlPacket.PresetMgr.LastRecall; lastRecall != nil {
                eventChan <- PresetRecallEvent{*lastRecall}
            }
        }
    }
}

func (client *Client) ListenEvents() (chan Event, error) {
    if xmlClient,err := client.xmlClient(); err != nil {
        return nil, err
    } else{
        eventChan := make(chan Event)

        go client.listenEvents(xmlClient, eventChan)

        return eventChan, nil
    }
}
