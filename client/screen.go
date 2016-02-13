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
    LayerCollection     LayerCollection `xml:"LayerCollection"`
}

type ScreenDestCol map[int]ScreenDest

func (col *ScreenDestCol) UnmarshalXML(d *xml.Decoder, e xml.StartElement) error {
    return unmarshalXMLMap(col, d, e)
}

func (col ScreenDestCol) List() (items []ScreenDest) {
    var keys []int

    for key, _ := range col {
        keys = append(keys, key)
    }

    sort.Ints(keys)

    for _, key := range keys {
        items = append(items, col[key])
    }

    return items
}
