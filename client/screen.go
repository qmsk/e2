package client

type ScreenDestination struct {
    ID      int     `json:"id"`
    Name    string  `json:"name"`
    HSize   int     `json:"HSize"`
    VSize   int     `json:"VSize"`
    Layers  int     `json:"Layers"`

    // DestOutMapCol
}

type Layer struct {
    ID          int     `json:"id" xml:"id,attr"`
    Name        string  `xml:"name"`

    LastSrcIdx  *int    `json:"LastSrcIdx" xml:"LastSrcIdx"`        // normalized to nil if -1

    PgmMode     *int    `json:"PgmMode" xml:"PgmMode"`
    PvwMode     *int    `json:"PvwMode" xml:"PvwMode"`

    PgmZOrder   int     `json:"PgmZOrder"`  // XXX: xml?
    PvwZOrder   int     `json:"PvwZOrder"`  // XXX: xml?

    Source      *Source `json:"-" xml:"LayerCfg>Source"`      // XXX: JSON is different!
    // Window
    // Mask
}

type BGColor struct {
    ID          int     `json:"id" xml:"id,attr"`
    Red         int     `json:"Red" xml:"Red"`
    Green       int     `json:"Green" xml:"Green"`
    Blue        int     `json:"Blue" xml:"Blue"`
}

type BGLayer struct {
    ID                  int         `json:"id" xml:"id,attr"`
    Name                string      `xml:"name"` // XXX: useful?

    LastBGSourceIndex   int         `"json:"LastBGSourceIndex" xml:"LastBGSourceIndex"`

    ShowMatte           int         `json:"BGShowMatte" xml:"BGShowMatte"`
    Color               BGColor     `json:"BGColor" xml:"BGColor"`
}

type Transition struct {
    ID                  int         `xml:"id,attr"`

    ArmMode             *int        `xml:"ArmMode"`
    TransPos            *int        `xml:"TransPos"`
    AutoTransInProg     *int        `xml:"AutoTransInProg"`
    TransInProg         *int        `xml:"TransInProg"`
}

type ScreenDest struct {
    ID                  int             `xml:"id,attr"`

    IsActive            *int            `xml:"IsActive"`
    BGLayer             BGLayer         `xml:"BGLyr"`
    Transition          []Transition    `xml:"Transition"`
    Layer               []Layer         `xml:"LayerCollection>Layer"`
}
