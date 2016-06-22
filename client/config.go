package client

import (
	"encoding/xml"
	"os"
)

func LoadSettingsFile(path string) (system System, err error) {
	if file, err := os.Open(path); err != nil {
		return system, err
	} else if err := xml.NewDecoder(file).Decode(&system); err != nil {
		return system, err
	} else {
		return system, nil
	}
}
