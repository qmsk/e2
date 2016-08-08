package server

import (
	"github.com/qmsk/e2/client"
)

type Status struct {
	clientOptions client.Options

	Server string `json:"server"`
	Mode   string `json:"mode"`
}

func (status Status) Get() (interface{}, error) {
	status.Server = status.clientOptions.String()

	if status.clientOptions.ReadOnly {
		status.Mode = "read"
	} else if status.clientOptions.Safe {
		status.Mode = "safe"
	} else {
		status.Mode = "live"
	}

	return status, nil
}
