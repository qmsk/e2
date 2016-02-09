package client

import (
    "fmt"
)

type SourceType int

const SourceTypeInput       SourceType  = 0
const SourceTypeStill       SourceType  = 1
const SourceTypeDest        SourceType  = 2

func (sourceType SourceType) String() string {
    switch sourceType {
    case SourceTypeInput:
        return "input"
    case SourceTypeStill:
        return "still"
    case SourceTypeDest:
        return "dest"
    default:
        return fmt.Sprintf("%d", int(sourceType))
    }
}

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

type Source struct {
    ID              int         `json:"id" xml:"id,attr"`

    Name            string
    HSize           int         `json:"HSize" xml:"AOIRect>HSize"`
    VSize           int         `json:"VSize" xml:"AOIRect>VSize"`
    SrcType         SourceType

    UserKeyIndex    int

    // -1 unless Type == SourceTypeInput
    InputCfgIndex       int
    InputVideoStatus    InputVideoStatus    `json:"InputCfgVideoStatus" xml:"-"` // XXX: xml

    // -1 unless Type == SourceTypeStill
    StillIndex      int

    // -1 unless Type == SourceTypeDest
    DestIndex       int
}

type listSources struct {
    Type            int         `json:"type"`
}

const listSourcesTypeInput      = 0
const listSourcesTypeBackground = 1

// Default is to return all
func (client *Client) ListSources() (sourceList []Source, err error) {
    request := Request{
        Method:     "listSources",
        Params:     struct{}{},
    }

    if err := client.doResult(&request, &sourceList); err != nil {
        return nil, err
    } else {
        return sourceList, nil
    }
}
