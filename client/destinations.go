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
