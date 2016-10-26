package client

import "encoding/xml"

type TransitionProgress float64 // 0.0 .. 1.0

// Zero value is false
func (value TransitionProgress) InProgress() bool {
	return value > 0.0
}

// Return float64 factor 0.0 .. 1.0
func (value TransitionProgress) Factor() float64 {
	return float64(value)
}

const TransitionPosMax = 4096

type Transition struct {
	ID int `xml:"id,attr"`

	ArmMode         int `xml:"ArmMode"`
	TransPos        int `xml:"TransPos"`
	AutoTransInProg int `xml:"AutoTransInProg"`
	TransInProg     int `xml:"TransInProg"`
}

// Workaroud missing TransInProg=0 update after Cut
func (transition Transition) InProgress() bool {
	return (transition.AutoTransInProg > 0) || (transition.TransInProg > 0 && transition.TransPos > 0)
}

// Return float 0.0 .. 1.0
func (transition Transition) Progress() TransitionProgress {
	return TransitionProgress(float64(transition.TransPos) / float64(TransitionPosMax))
}

type Transitions map[int]Transition

func (col *Transitions) UnmarshalXML(d *xml.Decoder, e xml.StartElement) error {
	return unmarshalXMLItem(col, d, e)
}

func (col Transitions) MarshalJSON() ([]byte, error) {
	return marshalJSONMap(col)
}
