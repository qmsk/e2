package client

import (
    "fmt"
    // "log"
)

type PresetRecallEvent struct {
    PresetID        int     `json:"preset_id"`
}

func (event PresetRecallEvent) String() string {
    return fmt.Sprintf("preset recall preset=%d", event.PresetID)
}

type AuxActiveEvent struct {
    AuxID           int     `json:"aux_id"`

    Active          bool    `json:"active"`
}

func (event AuxActiveEvent) String() string {
    return fmt.Sprintf("aux=%d active=%v", event.AuxID, event.Active)
}

type AuxEvent struct {
    AuxID           int     `json:"aux_id"`
    SourceID        int     `json:"source_id"`

    Program         bool    `json:"program"`
    Preview         bool    `json:"preview"`

    Source          Source  `json:"source"`
}

func (event AuxEvent) String() string {
    return fmt.Sprintf("aux=%d source=%d preview=%v program=%v", event.AuxID, event.SourceID, event.Program, event.Preview)
}

type ScreenActiveEvent struct {
    ScreenID        int     `json:"screen_id"`

    Active          bool    `json:"active"`
}

func (event ScreenActiveEvent) String() string {
    return fmt.Sprintf("screen=%d active=%v", event.ScreenID, event.Active)
}

type ScreenLayerSourceEvent struct {
    ScreenID        int     `json:"screen_id"`
    LayerID         int     `json:"layer_id"`
    SourceID        int     `json:"source_id"`

    Source          Source  `json:"source"`
}

func (event ScreenLayerSourceEvent) String() string {
    return fmt.Sprintf("screen=%d layer=%d source=%d", event.ScreenID, event.LayerID, event.SourceID)
}

type ScreenLayerEvent struct {
    ScreenID        int     `json:"screen_id"`
    LayerID         int     `json:"layer_id"`

    Program         bool    `json:"program"`
    Preview         bool    `json:"preview"`
}

func (event ScreenLayerEvent) String() string {
    return fmt.Sprintf("screen=%d layer=%d program=%v preview=%v", event.ScreenID, event.LayerID, event.Program, event.Preview)
}

type ScreenTransitionEvent struct {
    ScreenID        int     `json:"screen_id"`

    // Done if !InProgress
    InProgress      bool    `json:"transition_inprogress"`
    Auto            bool    `json:"transition_auto"`
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
    defer close(eventChan)

    for xmlPacket := range xmlClient.listenChan {
        if xmlPacket.DestMgr != nil {
            for _, auxDest := range xmlPacket.DestMgr.AuxDest {
                // log.Printf("Client.listenEvents: DestMgr.AuxDest %#v\n", auxDest)

                if auxDest.IsActive != nil {
                    eventChan <- AuxActiveEvent{auxDest.ID, *auxDest.IsActive > 0}
                }
                if auxDest.Source != nil {
                    source := *auxDest.Source

                    if auxDest.PvwLastSrcIndex != nil && *auxDest.PvwLastSrcIndex >= 0 {
                        source.ID = *auxDest.PvwLastSrcIndex

                        eventChan <- AuxEvent{
                            AuxID:      auxDest.ID,
                            SourceID:   source.ID,

                            Preview:    true,

                            Source:     source,
                        }
                    }
                    if auxDest.PgmLastSrcIndex != nil && *auxDest.PgmLastSrcIndex >= 0 {
                        source.ID = *auxDest.PgmLastSrcIndex

                        eventChan <- AuxEvent{
                            AuxID:      auxDest.ID,
                            SourceID:   source.ID,

                            Program:    true,

                            Source:     source,
                        }
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
