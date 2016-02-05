package client

import (
    "fmt"
)

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

// Preset Destinations
type PresetAuxDest struct {
    ID      int     `json:"id"`
}
type PresetScreenDest struct {
    ID      int     `json:"id"`
}

type PresetDestinations struct {
    Preset

    AuxDest     []PresetAuxDest     `json:"AuxDest"`
    ScreenDest  []PresetScreenDest  `json:"ScreenDest"`
}

type listDestinationsForPreset struct {
    ID      int     `json:"id"`
}

func (client *Client) ListDestinationsForPreset(presetID int) (result PresetDestinations, err error) {
    if presetID < 0 {
        return result, fmt.Errorf("Invalid Preset ID: %v", presetID)
    }

    request := Request{
        Method:     "listDestinationsForPreset",
        Params:     listDestinationsForPreset{
            ID:     presetID,
        },
    }

    if err := client.doResult(&request, &result); err != nil {
        return result, err
    } else {
        return result, nil
    }
}
