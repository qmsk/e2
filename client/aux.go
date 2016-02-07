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

type AuxDestCol struct {
    AuxDest         map[int]AuxDest
}

func (col *AuxDestCol) UnmarshalXML(d *xml.Decoder, e xml.StartElement) error {
    id, err := xmlID(e)
    if err != nil {
        return err
    }

    auxDest := col.AuxDest[id]

    if err := d.DecodeElement(&auxDest, &e); err != nil {
        return err
    }

    if col.AuxDest == nil {
        col.AuxDest = make(map[int]AuxDest)
    }

    col.AuxDest[id] = auxDest

    return nil
}

func (col AuxDestCol) List() (items []AuxDest) {
    var keys []int

    for key, _ := range col.AuxDest {
        keys = append(keys, key)
    }

    sort.Ints(keys)

    for _, key := range keys {
        items = append(items, col.AuxDest[key])
    }

    return items
}
