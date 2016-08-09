package server

import (
	"github.com/qmsk/e2/web"
)

func (status Status) Get() (interface{}, error) {
	return status, nil
}
func (server *Server) Get() (interface{}, error) {
	return server.GetState(), nil
}

func (server *Server) WebAPI() web.API {
	return web.MakeAPI(server)
}

func (server *Server) Index(name string) (web.Resource, error) {
	switch name {
	case "":
		return server, nil

	case "status":
		return server.GetStatus(), nil

	case "presets":

		return server.Presets()

	default:
		return nil, nil
	}
}

// Configure server to distribute web events.
//
// Call this before Start()
func (server *Server) WebEvents() *web.Events {
	server.eventChan = make(chan web.Event)

	return web.MakeEvents(server.eventChan)
}
