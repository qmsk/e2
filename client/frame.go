package client

import (
	"encoding/xml"
)

type Frame struct {
	ID            string `xml:"id,attr"`
	FrameType     int
	OpMode        int
	Name          string
	Contact       string
	OSVersion     string
	Version       string
	IsConnected   int
	SyncedSetting int

	// Enet
	// Slot
	// Genlock
	// SysCard
	// MultiViewer
	// OutMiniDin3D
}

// XML
type FrameCollection map[string]Frame

func (col *FrameCollection) UnmarshalXML(d *xml.Decoder, e xml.StartElement) error {
	return unmarshalXMLCol(col, d, e)
}

func (col FrameCollection) MarshalJSON() ([]byte, error) {
	return marshalJSONMap(col)
}
