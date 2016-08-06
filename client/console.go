package client

import (
	"encoding/xml"
	"fmt"
)

type ConsoleButtonType int

const (
	ConsleButtonTypeDestination ConsoleButtonType = 1
	ConsoleButtonTypeSource                       = 4
	ConsoleButtonTypeUserKey                      = 6
	ConsoleButtonTypePreset                       = 7
)

func (value ConsoleButtonType) String() string {
	switch (value) {
	case ConsleButtonTypeDestination:
		return "destination"
	case ConsoleButtonTypeSource:
		return "source"
	case ConsoleButtonTypeUserKey:
		return "user-key"
	case ConsoleButtonTypePreset:
		return "preset"
	default:
		return fmt.Sprintf("%d", value)
	}
}

func (value ConsoleButtonType) MarshalJSON() ([]byte, error) {
	return marshalJSONString(value)
}

type ConsoleButton struct {
	ID                     int		`xml:"id,attr"`
	ConsoleButtonType      ConsoleButtonType
	ConsoleButtonTypeIndex int
}

// Presets
type PresetBusColl map[int]ConsoleButton

func (col *PresetBusColl) UnmarshalXML(d *xml.Decoder, e xml.StartElement) error {
	return unmarshalXMLCol(col, d, e)
}

func (col PresetBusColl) MarshalJSON() ([]byte, error) {
	return marshalJSONMap(col)
}

// UserKeys
type UserKeyBusColl map[int]ConsoleButton

func (col *UserKeyBusColl) UnmarshalXML(d *xml.Decoder, e xml.StartElement) error {
	return unmarshalXMLCol(col, d, e)
}

func (col UserKeyBusColl) MarshalJSON() ([]byte, error) {
	return marshalJSONMap(col)
}

// Destinations
type DestinationBusColl map[int]ConsoleButton

func (col *DestinationBusColl) UnmarshalXML(d *xml.Decoder, e xml.StartElement) error {
	return unmarshalXMLCol(col, d, e)
}

func (col DestinationBusColl) MarshalJSON() ([]byte, error) {
	return marshalJSONMap(col)
}

// InputSource
type InputSourceBusColl map[int]ConsoleButton

func (col *InputSourceBusColl) UnmarshalXML(d *xml.Decoder, e xml.StartElement) error {
	return unmarshalXMLCol(col, d, e)
}

func (col InputSourceBusColl) MarshalJSON() ([]byte, error) {
	return marshalJSONMap(col)
}



type ConsoleLayout struct {
	PresetBusColl	   PresetBusColl
	UserKeyBusColl     UserKeyBusColl
	DestinationBusColl DestinationBusColl
	InputSourceBusColl InputSourceBusColl
}

type ConsoleLayoutMgr struct {
	ConsoleLayout ConsoleLayout
}
