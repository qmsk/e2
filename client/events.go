package client

import (
    "fmt"
    "log"
)

type PresetRecallEvent struct {
    PresetID        int
}

func (event PresetRecallEvent) String() string {
    return fmt.Sprintf("preset recall %d", event.PresetID)
}

type Event interface {}

func (client *Client) listenEvents(xmlClient *xmlClient, eventChan chan Event) {
    for xmlPacket := range xmlClient.listenChan {
        log.Printf("xmlClient: %#v\n", xmlPacket)

        if xmlPacket.PresetMgr != nil {
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
