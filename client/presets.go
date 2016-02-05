package client

type Preset struct {
    ID          int     `json:"id"`
    Name        string  `json:"Name"`
    LockMode    int     `json:"LockMode"`
    Seq         float64 `json:"presetSno"`
}

// Filtering
const listPresetsExclude = -2
const listPresetsInclude = -1

type listPresets struct {
    ScreenDest  int     `json:"ScreenDest"`
    AuxDest     int     `json:"AuxDest"`
}

// Default is to return all presets
func (client *Client) ListPresets() (presetList []Preset, err error) {
    request := Request{
        Method:     "listPresets",
        Params:     struct{}{},
    }

    if err := client.doResult(&request, &presetList); err != nil {
        return nil, err
    } else {
        return presetList, nil
    }
}

func (client *Client) ListPresetsX(screenID int, auxID int) (presetList []Preset, err error) {
    request := Request{
        Method:     "listPresets",
        Params:     listPresets{
            ScreenDest:     screenID,
            AuxDest:        auxID,
        },
    }

    if err := client.doResult(&request, &presetList); err != nil {
        return nil, err
    } else {
        return presetList, nil
    }
}


