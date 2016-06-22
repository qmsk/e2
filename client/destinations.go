package client

type listDestinations struct {
	Type int `json:"type"`
}

const listDestinationsTypeAll = 0
const listDestinationsTypeScreen = 1
const listDestinationsTypeAux = 2

type ListDestinations struct {
	AuxDestinations    []AuxDest           `json:"AuxDestination"`
	ScreenDestinations []ScreenDestination `json:"ScreenDestination"`
}

func (client *Client) ListDestinations() (result ListDestinations, err error) {
	request := Request{
		Method: "listDestinations",
		Params: listDestinations{
			Type: listDestinationsTypeAll,
		},
	}

	if err := client.doResult(&request, &result); err != nil {
		return result, err
	} else {
		return result, nil
	}
}

func (client *Client) ListAuxDestinations() ([]AuxDest, error) {
	var result ListDestinations

	request := Request{
		Method: "listDestinations",
		Params: listDestinations{
			Type: listDestinationsTypeAux,
		},
	}

	if err := client.doResult(&request, &result); err != nil {
		return nil, err
	} else {
		return result.AuxDestinations, nil
	}
}

func (client *Client) ListScreenDestinations() ([]ScreenDestination, error) {
	var result ListDestinations

	request := Request{
		Method: "listDestinations",
		Params: listDestinations{
			Type: listDestinationsTypeScreen,
		},
	}

	if err := client.doResult(&request, &result); err != nil {
		return nil, err
	} else {
		return result.ScreenDestinations, nil
	}
}

// Screen Content
type listContent struct {
	ID int `json:"id"`
}

type ListContent struct {
	ID   int    `json:"id"`
	Name string `json:"Name"`

	Layers   []*Layer  `json:"Layers"`
	BGLayers []BGLayer `json:"BgLyr"`

	// Transition
}

func (client *Client) ListContent(screenID int) (result ListContent, err error) {
	request := Request{
		Method: "listContent",
		Params: listContent{
			ID: screenID,
		},
	}

	if err := client.doResult(&request, &result); err != nil {
		return result, err
	} else {
		return result, nil
	}
}

// XML
type DestMgr struct {
	ID int `xml:"id,attr"`

	AuxDestCol    AuxDestCol    `xml:"AuxDestCol"`
	ScreenDestCol ScreenDestCol `xml:"ScreenDestCol"`
}
