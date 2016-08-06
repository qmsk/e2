package client

import (
	"encoding/xml"
	"sort"
)

type AuxDest struct {
	ID   int  `xml:"id,attr"`
	Name string

	AuxStreamMode   int
	IsActive        int
	PvwLastSrcIndex int
	PgmLastSrcIndex int

	// XXX: XML Collections
	Transition *Transition
	Source *Source
}

type AuxDestCol map[int]AuxDest

func (col *AuxDestCol) UnmarshalXML(d *xml.Decoder, e xml.StartElement) error {
	return unmarshalXMLCol(col, d, e)
}

func (col AuxDestCol) MarshalJSON() ([]byte, error) {
	return marshalJSONMap(col)
}

func (col AuxDestCol) List() (items []AuxDest) {
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
