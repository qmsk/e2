package server

import (
	"github.com/qmsk/e2/client"
	"github.com/qmsk/e2/web"
	"log"
)

type Event struct {
	System	*client.System
}

func (server *Server) WebEvents() *web.Events {
	eventChan := make(chan web.Event)

	go func() {
		defer close(eventChan)

		for {
			if system, err := server.xmlClient.Read(); err != nil {
				log.Printf("server:WebEvents: xmlClient.Read: %v", err)
				return
			} else {
				var event = Event{System: &system}

				eventChan <- event
			}
		}
	}()

	return web.MakeEvents(eventChan)
}
