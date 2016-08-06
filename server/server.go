package server

import (
	"github.com/qmsk/e2/client"
	"github.com/qmsk/e2/discovery"
	"github.com/qmsk/e2/web"
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
	} else {
		server.clientOptions = clientOptions
	}

	if client, err := server.clientOptions.Client(); err != nil {
		log.Fatalf("Client %#v: %v\n", server.clientOptions, err)
	} else {
		log.Printf("Client %#v: %v\n", server.clientOptions, client)

		server.client = client
	}

	if xmlClient, err := server.clientOptions.XMLClient(); err != nil {
		log.Fatalf("Client %#v: XMLClient: %v\n", server.clientOptions, err)
	} else {
		log.Printf("Client %#v: XMLClient: %v\n", server.clientOptions, xmlClient)

		server.xmlClient = xmlClient
	}


	return server, nil
}

type Server struct {
	options       Options
	clientOptions client.Options
	client        *client.Client
	xmlClient	  *client.XMLClient
}

func (server *Server) WebAPI() web.API {
	return web.MakeAPI(server)
}

func (server *Server) Index(name string) (web.Resource, error) {
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
