package client

import (
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
)

type Preset struct {
	ID       int     `json:"id" xml:"id,attr"`
	Name     string  `json:"Name"`
	LockMode int     `json:"LockMode"`
	Sno      float64 `json:"presetSno" xml:"presetSno"`
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
	ScreenDest int `json:"ScreenDest"`
	AuxDest    int `json:"AuxDest"`
}

// Default is to return all presets
func (client *Client) ListPresets() (presetList []Preset, err error) {
	request := Request{
		Method: "listPresets",
		Params: struct{}{},
	}

	if err := client.doResult(&request, &presetList); err != nil {
		return nil, err
	} else {
		return presetList, nil
	}
}

func (client *Client) ListPresetsX(screenID int, auxID int) (presetList []Preset, err error) {
	request := Request{
		Method: "listPresets",
		Params: listPresets{
			ScreenDest: screenID,
			AuxDest:    auxID,
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
	ID int `json:"id"`
}
type PresetScreenDest struct {
	ID int `json:"id"`
}

type PresetDestinations struct {
	Preset

	AuxDest    []PresetAuxDest    `json:"AuxDest"`
	ScreenDest []PresetScreenDest `json:"ScreenDest"`
}

type listDestinationsForPreset struct {
	ID int `json:"id"`
}

func (client *Client) ListDestinationsForPreset(presetID int) (result PresetDestinations, err error) {
	if presetID < 0 {
		return result, fmt.Errorf("Invalid Preset ID: %v", presetID)
	}

	request := Request{
		Method: "listDestinationsForPreset",
		Params: listDestinationsForPreset{
			ID: presetID,
		},
	}

	if err := client.doResult(&request, &result); err != nil {
		return result, err
	} else {
		return result, nil
	}
}

// XML
type PresetCol map[int]Preset

// unmarshal each <Preset> separately
// TODO: this is not as efficient, since we COW the PresetCol map on every Preset...
func (col *PresetCol) UnmarshalXML(d *xml.Decoder, e xml.StartElement) error {
	return unmarshalXMLItem(col, d, e)
}

func (col PresetCol) MarshalJSON() ([]byte, error) {
	return marshalJSONMap(col)
}

func (col PresetCol) List() (items []Preset) {
	var keys []int

	for key, _ := range col {
		keys = append(keys, key)
	}

	sort.Ints(keys)

	for _, key := range keys {
		items = append(items, col[key])
	}

	return items
}

type PresetMgr struct {
	ID int `xml:"id,attr"`

	LastRecall int `xml:"LastRecall"`

	// E2 3.1
	ConflictMode int
	ConflictPref int
	TransTime    int
	GuiId        string
	Conflict     int

	// TODO: ignore extra fields
	Preset PresetCol `xml:",any"` // <Preset> or <Add><Preset> or <Remove><Preset>
}
