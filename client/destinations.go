package client

type AuxDestination struct {
    ID              int     `json:"id"`
    Name            string  `json:"name"`
    AuxStreamMode   int     `json:"AuxStreamMode"`
}

type ScreenDestination struct {
    ID      int     `json:"id"`
    Name    string  `json:"name"`
    HSize   int     `json:"HSize"`
    VSize   int     `json:"VSize"`
    Layers  int     `json:"Layers"`

    // DestOutMapCol
}

type listDestinations struct {
    Type    int     `json:"type"`
}

const listDestinationsTypeAll       = 0
const listDestinationsTypeScreen    = 1
const listDestinationsTypeAux       = 2

type ListDestinations struct {
    AuxDestination          []AuxDestination        `json:"AuxDestination"`
    ScreenDestination       []ScreenDestination     `json:"ScreenDestination"`
}

func (client *Client) ListDestinations() (result ListDestinations, err error) {
    request := Request{
        Method:     "listDestinations",
        Params:     listDestinations{
            Type:           listDestinationsTypeAll,
        },
    }

    if err := client.doResult(&request, &result); err != nil {
        return result, err
    } else {
        return result, nil
    }
}

// Screen Content
type listContent struct {
    ID      int     `json:"id"`
}

type Layer struct {
    ID          int     `json:"id"`

    LastSrcIdx  int     `json:"LastSrcIdx"`

    PgmMode     int     `json:"PgmMode"`
    PvwMode     int     `json:"PvwMode"`

    PgmZOrder   int     `json:"PgmZOrder"`
    PvwZOrder   int     `json:"PvwZOrder"`

    // Source
    // Window
    // Mask
}

type BGColor struct {
    ID          int     `json:"id"`
    Red         int     `json:"Red"`
    Green       int     `json:"Green"`
    Blue        int     `json:"Blue"`
}

type BGLayer struct {
    ID                  int         `json:"id"`

    LastBGSourceIndex   int         `"json:"LastBGSourceIndex"`

    ShowMatte           int         `json:"BGShowMatte"`
    Color               BGColor     `json:"BGColor"`
}

type ListContent struct {
    ID          int             `json:"id"`
    Name        string          `json:"Name"`

    Layers      []Layer         `json:"Layers"`
    BGLayers    []BGLayer       `json:"BgLyr"`

    // Transition
}

func (client *Client) ListContent(screenID int) (result ListContent, err error) {
    request := Request{
        Method:     "listContent",
        Params:     listContent{
            ID:     screenID,
        },
    }

    if err := client.doResult(&request, &result); err != nil {
        return result, err
    } else {
        return result, nil
    }
}
