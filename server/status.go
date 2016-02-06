package server

import (
    "github.com/qmsk/e2/client"
)

type Status struct {
    client      *client.Client

    Server      string      `json:"server"`
    Mode        string      `json:"mode"`
}

func (status Status) Get() (interface{}, error) {
    status.Server = status.client.String()
    status.Mode = "live"

    return status, nil
}
