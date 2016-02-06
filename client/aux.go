package client

type AuxDestination struct {
    ID              int     `json:"id"`
    Name            string  `json:"name"`
    AuxStreamMode   int     `json:"AuxStreamMode"`
}

type AuxDest struct {
    ID      int     `xml:"id,attr"`
}

type AuxDestCol struct {
    ID          int         `xml:"id,attr"`

    AuxDests    []AuxDest   `xml:"AuxDest"`
}


