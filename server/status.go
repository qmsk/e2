package server

type Status struct {
	options Options

	Server string `json:"server"`
	Mode   string `json:"mode"`
}

func (status Status) Get() (interface{}, error) {
	status.Server = status.options.ClientOptions.String()
	status.Mode = "live"

	return status, nil
}
