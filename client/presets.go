package client

import (
    "fmt"
    "sort"
    "strings"
    "encoding/xml"
)

type Preset struct {
    ID          int     `json:"id" xml:"id,attr"`
    Name        string  `json:"Name"`
    LockMode    int     `json:"LockMode"`
    Sno         float64 `json:"presetSno" xml:"presetSno"`
}

func (preset Preset) ParseOrder() (group int, index int) {
    // one awesome hack
    sno := strings.Trim(fmt.Sprintf("%f", preset.Sno), "0")

    if _, err := fmt.Sscanf(sno, "%d.%d", &group, &index); err != nil {
        // 0.0 is invalid..
        return 0, 0
    }

    return
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

// XML
type PresetMap map[int]Preset

func (m *PresetMap) UnmarshalXML(d *xml.Decoder, e xml.StartElement) error {
    if id, err := xmlID(e); err != nil {
        return err
    } else {
        value := (*m)[id]

        if err := d.DecodeElement(&value, &e); err != nil {
            return err
        }

        if *m == nil {
            *m = make(PresetMap)
        }

        (*m)[id] = value

        return nil
    }
}

func (m PresetMap) List() []Preset {
    var keys []int
    var items []Preset

    for key, _ := range m {
        keys = append(keys, key)
    }

    sort.Ints(keys)


    for _, key := range keys {
        items = append(items, m[key])
    }

    return items
}

type PresetMgr struct {
    ID          int             `xml:"id,attr"`

    LastRecall  int             `xml:"LastRecall"`

    Preset      PresetMap       `xml:"Preset"`
}
