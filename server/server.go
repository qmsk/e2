package server

import (
	"fmt"
	"github.com/qmsk/e2/client"
	"github.com/qmsk/e2/discovery"
	"github.com/qmsk/e2/web"
	"log"
	"sync/atomic"
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

	if jsonClient, err := server.clientOptions.JSONClient(); err != nil {
		log.Fatalf("Client %#v: %v\n", server.clientOptions, err)
	} else {
		log.Printf("Client %#v: %v\n", server.clientOptions, jsonClient)

		server.jsonClient = jsonClient
	}

	if xmlClient, err := server.clientOptions.XMLClient(); err != nil {
		log.Fatalf("Client %#v: XMLClient: %v\n", server.clientOptions, err)
	} else {
		log.Printf("Client %#v: XMLClient: %v\n", server.clientOptions, xmlClient)

		server.xmlClient = xmlClient
	}

	if tcpClient, err := server.clientOptions.TCPClient(); err != nil {
		log.Fatalf("Client %v: TCPClient %v", server.clientOptions, err)
	} else {
		log.Printf("Client %v: TCPClient: %v", server.clientOptions, tcpClient)

		server.tcpClient = tcpClient
	}

	return server, nil
}

type Server struct {
	options       Options
	clientOptions client.Options
	jsonClient    *client.JSONClient
	xmlClient	  *client.XMLClient
	tcpClient	  *client.TCPClient

	state		  atomic.Value
	eventChan     chan web.Event
}

type State struct {
	System	*client.System
}

func (server *Server) Run() error {
	if server.eventChan != nil {
		defer close(server.eventChan)
	}

	for {
		if system, err := server.xmlClient.Read(); err != nil {
			return fmt.Errorf("xmlClient.Read: %v", err)
		} else {
			var state = State{System: &system}

			server.state.Store(state)

			if server.eventChan != nil {
				server.eventChan <- state
			}
		}
	}
}

func (server *Server) GetState() State {
	return server.state.Load().(State)
}
