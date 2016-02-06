package client

type AuxDestination struct {
    ID              int     `json:"id"`
    Name            string  `json:"name"`
    AuxStreamMode   int     `json:"AuxStreamMode"`
}

type AuxDest struct {
    ID              int             `xml:"id,attr"`

    IsActive        *int            `xml:"IsActive"`

    Transition      *Transition     `xml:"Transition"`

    PvwLastSrcIndex *int            `xml:"PvwLastSrcIndex"`
    PgmLastSrcIndex *int            `xml:"PgmLastSrcIndex"`

    Source          *Source         `xml:"Source"`
}

type AuxDestCol struct {
    ID          int         `xml:"id,attr"`

    AuxDests    []AuxDest   `xml:"AuxDest"`
}


