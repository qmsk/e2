package client

import (
    "fmt"
    "sort"
    "encoding/xml"
)

type InputVideoStatus int

const InputVideoStatusOK    = 1
const InputVideoStatusBad   = 4

func (vs InputVideoStatus) String() string {
    switch vs {
    case InputVideoStatusOK:
        return "ok"
    case InputVideoStatusBad:
        return "bad"
    default:
        return fmt.Sprintf("%d", int(vs))
    }
}

type InputCfg struct {
    ID                  int         `json:"id" xml:"id,attr"`

    InputCfgType        int
    Name                string
    InputCfgVideoStatus InputVideoStatus

    ConfigOwner         string      `xml:"Config>Owner"`
    ConfigContact       string      `xml:"Config>Contact"`
}

// XML
type InputCfgCol map[int]InputCfg

func (col *InputCfgCol) UnmarshalXML(d *xml.Decoder, e xml.StartElement) error {
    return unmarshalXMLMap(col, d, e)
}

func (col InputCfgCol) MarshalJSON() ([]byte, error) {
    return marshalJSONMap(col)
}

func (col InputCfgCol) List() (items []InputCfg) {
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
