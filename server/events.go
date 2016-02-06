package server

import (
    "github.com/qmsk/e2/client"
    "fmt"
    "net/http"
    "log"
    "golang.org/x/net/websocket"
)

const EVENTS_BUFFER = 100

type Event struct {
    Type        string      `json:"type"`
    Data        interface{} `json:"data"`
    Line        string      `json:"line"`
}

type clientSet map[chan Event]bool

func (clientSet clientSet) register(clientChan chan Event) {
    clientSet[clientChan] = true
}
func (clientSet clientSet) unregister(clientChan chan Event) {
    delete(clientSet, clientChan)
}
func (clientSet clientSet) publish(event Event) {
    for clientChan, _ := range clientSet {
        select {
        case clientChan <- event:

        default:
            // dropped
        }
    }
}

type Events struct {
    eventChan   chan client.Event

    registerChan    chan chan Event
    unregisterChan  chan chan Event
}

func (server *Server) Events() (*Events, error) {
    events := Events{
        registerChan:       make(chan chan Event),
        unregisterChan:     make(chan chan Event),
    }

    if eventChan, err := server.client.ListenEvents(); err != nil {
        return nil, err
    } else {
        events.eventChan = eventChan
    }

    go events.run()

    return &events, nil
}
func (events *Events) run() {
    clients := make(clientSet)

    for {
        select {
        case clientChan := <-events.registerChan:
            clients.register(clientChan)

        case clientChan := <-events.unregisterChan:
            clients.unregister(clientChan)

        case clientEvent := <-events.eventChan:
            event := Event{
                Type:   fmt.Sprintf("%T", clientEvent),
                Data:   clientEvent,
                Line:   clientEvent.String(),
            }

            log.Printf("Events: %v\n", event)

            clients.publish(event)
        }
    }
}

func (events *Events) register() chan Event {
    eventChan := make(chan Event, EVENTS_BUFFER)

    events.registerChan <- eventChan

    return eventChan
}

func (events *Events) unregister(eventChan chan Event) {
    events.unregisterChan <- eventChan
}

// goroutine-safe websocket subscriber
func (events *Events) ServeWebsocket(websocketConn *websocket.Conn) {
    eventChan := events.register()
    defer events.unregister(eventChan)

    for event := range eventChan {
        if err := websocket.JSON.Send(websocketConn, event); err != nil {
            return
        }
    }
}

func (events *Events) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    websocket.Handler(events.ServeWebsocket).ServeHTTP(w, r)
}
