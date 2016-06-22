package client

import (
	"encoding/xml"
	"fmt"
	"sort"
)

type SourceType int

const SourceTypeInput SourceType = 0
const SourceTypeStill SourceType = 1
const SourceTypeDest SourceType = 2

func (sourceType SourceType) String() string {
	switch sourceType {
	case SourceTypeInput:
		return "input"
	case SourceTypeStill:
		return "still"
	case SourceTypeDest:
		return "dest"
	default:
		return fmt.Sprintf("%d", int(sourceType))
	}
}

type Source struct {
	ID int `json:"id" xml:"id,attr"`

	Name    string
	HSize   int `json:"HSize" xml:"AOIRect>HSize"`
	VSize   int `json:"VSize" xml:"AOIRect>VSize"`
	SrcType SourceType

	UserKeyIndex int

	// -1 unless Type == SourceTypeInput
	InputCfgIndex    int
	InputVideoStatus InputVideoStatus `json:"InputCfgVideoStatus" xml:"-"` // XXX: JSON isi different

	// -1 unless Type == SourceTypeStill
	StillIndex int

	// -1 unless Type == SourceTypeDest
	DestIndex int
}

type listSources struct {
	Type int `json:"type"`
}

const listSourcesTypeInput = 0
const listSourcesTypeBackground = 1

// Default is to return all
func (client *Client) ListSources() (sourceList []Source, err error) {
	request := Request{
		Method: "listSources",
		Params: struct{}{},
	}

	if err := client.doResult(&request, &sourceList); err != nil {
		return nil, err
	} else {
		return sourceList, nil
	}
}

// XML
type SourceCol map[int]Source

func (col *SourceCol) UnmarshalXML(d *xml.Decoder, e xml.StartElement) error {
	return unmarshalXMLCol(col, d, e)
}

func (col SourceCol) MarshalJSON() ([]byte, error) {
	return marshalJSONMap(col)
}

func (col SourceCol) Get(id int) *Source {
	if item, exists := col[id]; exists {
		return &item
	} else {
		return nil
	}
}

func (col SourceCol) List() (items []Source) {
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

type SrcMgr struct {
	ID int `xml:"id,attr"`

	SourceCol SourceCol `xml:"SourceCol"`
	//BGSourceCol
	InputCfgCol InputCfgCol `xml:"InputCfgCol"`
	//SavedInputCfgCol
}
