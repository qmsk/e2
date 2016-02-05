package client

import (
    "fmt"
)

type SourceType int

const SourceTypeInput       SourceType  = 0
const SourceTypeDest        SourceType  = 2

func (sourceType SourceType) String() string {
    switch sourceType {
    case SourceTypeInput:
        return "input"
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
    ID              int         `json:"id"`
    Name            string      `json:"Name"`
    HSize           int         `json:"HSize"`
    VSize           int         `json:"VSize"`
    Type            SourceType  `json:"SrcType"`

    InputCfgIndex   int         `json:"InputCfgIndex"`
    UserKeyIndex    int         `json:"UserKeyIndex"`
    StillIndex      int         `json:"StillIndex"`
    DestIndex       int         `json:"DestIndex"`

    InputVideoStatus    InputVideoStatus    `json:"InputCfgVideoStatus"`    // 0 unless InputCfgIndex >= 0
}

func (self Source) cacheID() int { return self.ID }

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
