package client

import (
    "sort"
    "encoding/xml"
)

type ScreenDestination struct {
    ID      int     `json:"id"`
    Name    string  `json:"name"`
    HSize   int     `json:"HSize"`
    VSize   int     `json:"VSize"`
    Layers  int     `json:"Layers"`

    //DestOutMapCol
}

type ScreenDest struct {
    ID                  int             `xml:"id,attr"`

    IsActive            int
    Name                string
    HSize               int
    VSize               int

    BGLayer             []BGLayer       `xml:"BGLyr"`
    Transition          []Transition    `xml:"Transition"`
    LayerCollection     LayerCollection `xml:"LayerCollection>Layer"`
}

type ScreenDestCol struct {
    ScreenDest      map[int]ScreenDest
}

func (col *ScreenDestCol) UnmarshalXML(d *xml.Decoder, e xml.StartElement) error {
    id, err := xmlID(e)
    if err != nil {
        return err
    }

    screenDest := col.ScreenDest[id]

    if err := d.DecodeElement(&screenDest, &e); err != nil {
        return err
    }

    if col.ScreenDest == nil {
        col.ScreenDest = make(map[int]ScreenDest)
    }

    col.ScreenDest[id] = screenDest

    return nil
}

func (col ScreenDestCol) List() (items []ScreenDest) {
    var keys []int

    for key, _ := range col.ScreenDest {
        keys = append(keys, key)
    }

    sort.Ints(keys)

    for _, key := range keys {
        items = append(items, col.ScreenDest[key])
    }

    return items
}
