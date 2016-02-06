package client

import (
    "fmt"
    // "log"
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

type ScreenActiveEvent struct {
    ScreenID        int
    Active          bool
}

func (event ScreenActiveEvent) String() string {
    return fmt.Sprintf("screen=%d active=%v", event.ScreenID, event.Active)
}

type ScreenLayerSourceEvent struct {
    ScreenID        int
    LayerID         int
    SourceID        int

    Source          Source
}

func (event ScreenLayerSourceEvent) String() string {
    return fmt.Sprintf("screen=%d layer=%d source=%d", event.ScreenID, event.LayerID, event.SourceID)
}

type ScreenLayerEvent struct {
    ScreenID        int
    LayerID         int

    Program         bool
    Preview         bool
}

func (event ScreenLayerEvent) String() string {
    return fmt.Sprintf("screen=%d layer=%d program=%v preview=%v", event.ScreenID, event.LayerID, event.Program, event.Preview)
}

type ScreenTransitionEvent struct {
    ScreenID        int

    InProgress      bool
    Auto            bool
}

func (event ScreenTransitionEvent) String() string {
    if !event.InProgress {
        return fmt.Sprintf("screen=%d transition done", event.ScreenID)
    } else {
        return fmt.Sprintf("screen=%d transition auto=%v", event.ScreenID, event.Auto)
    }
}

type Event interface {
    String()    string
}

func (client *Client) listenEvents(xmlClient *xmlClient, eventChan chan Event) {
    for xmlPacket := range xmlClient.listenChan {
        if xmlPacket.DestMgr != nil {
            for _, auxDest := range xmlPacket.DestMgr.AuxDest {
                // log.Printf("Client.listenEvents: DestMgr.AuxDest %#v\n", auxDest)

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
                // log.Printf("Client.listenEvents: DestMgr.ScreenDest %#v\n", screenDest)

                if screenDest.IsActive != nil {
                    eventChan <- ScreenActiveEvent{screenDest.ID, *screenDest.IsActive > 0}
                }

                for _, layer := range screenDest.Layer {
                    if layer.Source != nil && layer.LastSrcIdx != nil {
                        source := *layer.Source
                        source.ID= *layer.LastSrcIdx

                        eventChan <- ScreenLayerSourceEvent{
                            ScreenID:   screenDest.ID,
                            LayerID:    layer.ID,
                            SourceID:   *layer.LastSrcIdx,

                            Source:     source,
                        }
                    }

                    if layer.PvwMode != nil || layer.PgmMode != nil {
                        eventChan <- ScreenLayerEvent{
                            ScreenID:   screenDest.ID,
                            LayerID:    layer.ID,

                            Preview:    layer.PvwMode != nil && *layer.PvwMode > 0,
                            Program:    layer.PgmMode != nil && *layer.PgmMode > 0,
                        }
                    }
                }

                for _, transition := range screenDest.Transition {
                    if transition.TransInProg == nil {

                    } else if *transition.TransInProg == 0 {
                        eventChan <- ScreenTransitionEvent{
                            ScreenID:   screenDest.ID,
                        }
                    } else if transition.AutoTransInProg != nil {
                        eventChan <- ScreenTransitionEvent{
                            ScreenID:   screenDest.ID,

                            InProgress: true,
                            Auto:       *transition.AutoTransInProg > 0,
                        }
                    } else {
                        eventChan <- ScreenTransitionEvent{
                            ScreenID:   screenDest.ID,

                            InProgress: true,
                        }
                    }
                }
            }
        }

        if xmlPacket.PresetMgr != nil {
            // log.Printf("Client.listenEvents: PresetMgr %#v\n", xmlPacket.PresetMgr)

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
