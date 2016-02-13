package client

import (
    "sort"
    "encoding/xml"
)

type AuxDest struct {
    ID              int             `json:"id" xml:"id,attr"`
    Name            string          `json:"name" xml:"Name"`

    AuxStreamMode   int             `xml:"AuxStreamMode"`
    IsActive        int             `xml:"IsActive"`
    PvwLastSrcIndex int             `xml:"PvwLastSrcIndex"`
    PgmLastSrcIndex int             `xml:"PgmLastSrcIndex"`

    Transition      *Transition     `xml:"Transition"`
    Source          *Source         `xml:"Source"`
}

type AuxDestCol map[int]AuxDest

func (col *AuxDestCol) UnmarshalXML(d *xml.Decoder, e xml.StartElement) error {
    return unmarshalXMLMap(col, d, e)
}

func (col AuxDestCol) MarshalJSON() ([]byte, error) {
    return marshalJSONMap(col)
}

func (col AuxDestCol) List() (items []AuxDest) {
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
