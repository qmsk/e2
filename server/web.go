package server

import (
	"github.com/qmsk/e2/web"
)

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
		status := Status{
			clientOptions: server.clientOptions,
		}

		return &status, nil

	case "presets":
		presets := Presets{
			system: server.GetState().System,
			jsonClient: server.jsonClient,
			tcpClient: server.tcpClient,
		}

		return &presets, presets.load()

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
