package client

import (
    "log"
)

type System struct {
    //SrcMgr
    DestMgr     DestMgr     `xml:"DestMgr"`
    PresetMgr   PresetMgr   `xml:"PresetMgr"`
}

func (system *System) Reset() {
    log.Printf("client: System.Reset()\n")

    *system = System{}
}
