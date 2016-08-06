package client

// XML
type DestMgr struct {
	ID int `xml:"id,attr"`

	AuxDestCol    AuxDestCol    `xml:"AuxDestCol"`
	ScreenDestCol ScreenDestCol `xml:"ScreenDestCol"`
}
