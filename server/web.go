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
	case "system":
		return server, nil

	case "":
		index := Index{}

		return &index, index.load(server.client)

	case "status":
		status := Status{
			client: server.client,
		}

		return &status, nil

	case "sources":
		sources := Sources{}

		return &sources, sources.load(server.client)

	case "screens":
		screens := Screens{
			client: server.client,
		}

		return &screens, screens.load(server.client)

	case "auxes":
		auxes := Auxes{}

		return &auxes, auxes.load(server.client)

	case "presets":
		presets := Presets{
			client: server.client,
		}

		return &presets, presets.load(server.client)

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
