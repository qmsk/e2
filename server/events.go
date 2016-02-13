package server

import (
    "github.com/qmsk/e2/client"
    "net/http"
    "log"
    "golang.org/x/net/websocket"
)

const EVENTS_BUFFER = 100

type Event struct {
    System      client.System   `json:"system"`
    String      string          `json:"string"`
}

type clientSet map[chan Event]bool

func (clientSet clientSet) register(clientChan chan Event) {
    clientSet[clientChan] = true
}
func (clientSet clientSet) unregister(clientChan chan Event) {
    // remove from set on behalf of client; the clientChan may already be closed
    delete(clientSet, clientChan)
}
func (clientSet clientSet) drop(clientChan chan Event) {
    // close and remove a dead client
    // the client may trigger .unregister() later, but that's okay
    close(clientChan)
    delete(clientSet, clientChan)
}
func (clientSet clientSet) send(clientChan chan Event, event Event) {
    select {
    case clientChan <- event:

    default:
        // client dropped behind
        clientSet.drop(clientChan)
    }
}

func (clientSet clientSet) publish(event Event) {
    for clientChan, _ := range clientSet {
        clientSet.send(clientChan, event)
    }
}

func (clientSet clientSet) close() {
    for clientChan, _ := range clientSet {
        clientSet.drop(clientChan)
    }
}

type Events struct {
    listenChan   chan client.System

    registerChan    chan chan Event
    unregisterChan  chan chan Event
}

func (server *Server) Events() (*Events, error) {
    events := Events{
        registerChan:       make(chan chan Event),
        unregisterChan:     make(chan chan Event),
    }

    if xmlClient, err := server.clientOptions.XMLClient(); err != nil {
        return nil, err
    } else if listenChan, err := xmlClient.Listen(); err != nil {
        return nil, err
    } else {
        events.listenChan = listenChan
    }

    go events.run()

    return &events, nil
}

func (events *Events) run() {
    var event Event

    clients := make(clientSet)
    defer clients.close()

    // panics any subscribed clients
    defer close(events.registerChan)
    defer close(events.unregisterChan)

    for {
        select {
        case clientChan := <-events.registerChan:
            clients.register(clientChan)

            // initial state
            clients.send(clientChan, event)

        case clientChan := <-events.unregisterChan:
            clients.unregister(clientChan)

        case system, ok := <-events.listenChan:
            if !ok {
                // TODO: recover..
                panic("Events died")
            }

            // update state
            event = Event{
                System:     system,
                String:     system.String(),
            }

            // log.Printf("Events: %v\n", event)

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
            log.Printf("webSocket.JSON.Send: %v\n", err)
            return
        }
    }
}

func (events *Events) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    websocket.Handler(events.ServeWebsocket).ServeHTTP(w, r)
}
