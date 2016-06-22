// Implement the HETEC V-Switch quad Device Control Protocol
//  http://www.hetec.de/fileadmin/hetec_share/Produkte/Hetec/V-Switch_quad/DCP_Handbuch_1_2_En_Hetec.pdf
package dcp

type Device struct {
	Type    string `xml:"type"`
	Version struct {
		DCPProtocol string `xml:"dcp-protocol"`
		Hardware    string `xml:"hardware"`
		Software    string `xml:"software"`
	} `xml:"version"`

	Mode struct {
		Console Console `xml:"console"`
		Video   Video   `xml:"video"`
	} `xml:"mode"`
}

type Console struct {
	Channel int `xml:"channel"`
}

type Video struct {
	Channel int    `xml:"channel"`
	Layout  string `xml:"layout"`
	PIP     struct {
		Layout string `xml:"layout"`

		// TODO

	} `xml:"pip"`
}
