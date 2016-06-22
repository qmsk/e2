package server

import (
	"github.com/qmsk/e2/client"
	"github.com/qmsk/e2/discovery"
	"log"
)

type Options struct {
	DiscoveryOptions discovery.Options `group:"E2 Discovery"`
	ClientOptions    client.Options    `group:"E2 JSON-RPC"`
}

func (options Options) Server() (*Server, error) {
	server := &Server{
		options: options,
	}

	if clientOptions, err := options.ClientOptions.DiscoverClient(options.DiscoveryOptions); err != nil {
		log.Fatalf("Client %#v: Discover %#v: %v\n", options.ClientOptions, options.DiscoveryOptions, err)
	} else if client, err := clientOptions.Client(); err != nil {
		log.Fatalf("Client %#v: %v\n", clientOptions, err)
	} else {
		log.Printf("Client %#v: %v\n", clientOptions, client)

		server.clientOptions = clientOptions
		server.client = client
	}

	return server, nil
}

type Server struct {
	options       Options
	clientOptions client.Options
	client        *client.Client
}

func (server *Server) Index(name string) (apiResource, error) {
	switch name {
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
