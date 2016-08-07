package client

import (
	"encoding/xml"
	"fmt"
	"sort"
)

type PresetSno struct {
	Group, Index	int
}

func (sno PresetSno) String() string {
	return fmt.Sprintf("%d.%d", sno.Group, sno.Index)
}

func (sno *PresetSno) parse(value string) error {
	if _, err := fmt.Sscanf(value, "%d.%d", &sno.Group, &sno.Index); err != nil {
		return err
	} else {
		return nil
	}
}

func (sno *PresetSno) UnmarshalXML(d *xml.Decoder, e xml.StartElement) error {
	var value string

	if err := d.DecodeElement(&value, &e); err != nil {
		return err
	}

	return sno.parse(value)
}

func (sno *PresetSno) UnmarshalJSON(value []byte) error {
	return sno.parse(string(value))
}

func (sno PresetSno) MarshalJSON(value string) ([]byte, error) {
	// as float
	return []byte(sno.String()), nil
}

type Preset struct {
	ID       int       `json:"id" xml:"id,attr"`
	Name     string    `json:"Name"`
	LockMode int       `json:"LockMode"`
	Sno      PresetSno `json:"presetSno" xml:"presetSno"`
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
