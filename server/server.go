package server

import (
    "github.com/qmsk/e2/client"
)

type Options struct {

}

func (options Options) Server(clientClient *client.Client) (*Server, error) {
    server := &Server{
        client:     clientClient,

        sources:    Sources{
            client:         clientClient,
        },
        screens:    Screens{
            client:         clientClient,
        },
    }

    return server, nil
}

type Server struct {
    client      *client.Client

    sources     Sources
    screens     Screens
}

func (server *Server) Index(name string) (apiResource, error) {
    switch name {
    case "sources":
        return &server.sources, nil
    case "screens":
        return &server.screens, nil
    default:
        return nil, nil
    }
}

func (server *Server) Get() (interface{}, error) {
    return "Hello World", nil
}
