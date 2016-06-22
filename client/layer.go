package client

import (
	"encoding/xml"
	"sort"
)

type BGColor struct {
	ID    int `json:"id" xml:"id,attr"`
	Red   int `json:"Red" xml:"Red"`
	Green int `json:"Green" xml:"Green"`
	Blue  int `json:"Blue" xml:"Blue"`
}

type BGLayer struct {
	ID   int    `json:"id" xml:"id,attr"`
	Name string `xml:"name"` // XXX: useful?

	LastBGSourceIndex int `"json:"LastBGSourceIndex" xml:"LastBGSourceIndex"`

	ShowMatte int     `json:"BGShowMatte" xml:"BGShowMatte"`
	Color     BGColor `json:"BGColor" xml:"BGColor"`
}

type Transition struct {
	ID int `xml:"id,attr"`

	ArmMode         int `xml:"ArmMode"`
	TransPos        int `xml:"TransPos"`
	AutoTransInProg int `xml:"AutoTransInProg"`
	TransInProg     int `xml:"TransInProg"`
}

type Layer struct {
	ID int `json:"id" xml:"id,attr"`

	PgmMode        int
	PvwMode        int
	IsActive       int
	PgmZOrder      int
	PvwZOrder      int
	LastSrcIdx     int // -1 if invalid
	LastUserKeyIdx int
	Name           string

	Source *Source `json:"-" xml:"LayerCfg>Source"` // XXX: JSON is different!
	//Window
	//Mask
}

type LayerCollection map[int]Layer

func (col *LayerCollection) UnmarshalXML(d *xml.Decoder, e xml.StartElement) error {
	return unmarshalXMLCol(col, d, e)
}

func (col LayerCollection) MarshalJSON() ([]byte, error) {
	return marshalJSONMap(col)
}

func (col LayerCollection) List() (items []Layer) {
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
