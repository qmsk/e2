
package main

import (
    "github.com/jessevdk/go-flags"
    "net/http"
    "log"
    "github.com/qmsk/e2/server"
)

var options = struct{
    ServerOptions       server.Options          `group:"E2 Server"`

    HTTPListen      string              `long:"http-listen" value-name:"[HOST]:PORT" default:":8284"`
    HTTPStatic      string              `long:"http-static" value-name:"PATH"`
}{}

var parser = flags.NewParser(&options, flags.Default)

func main() {
    if _, err := parser.Parse(); err != nil {
        log.Fatalf("%v\n", err)
    }

    server, err := options.ServerOptions.Server()
    if err != nil {
        log.Fatalf("Server %#v: %v\n", options.ServerOptions, err)
    }

    http.Handle("/api/", http.StripPrefix("/api/", server))

    if events, err := server.Events(); err != nil {
        log.Fatalf("Server.Events: %v\n", err)
    } else {
        http.Handle("/events", events)
    }

    if options.HTTPStatic != "" {
        log.Printf("Serve / from %v\n", options.HTTPStatic)

        http.Handle("/", http.FileServer(http.Dir(options.HTTPStatic)))
    }

    if options.HTTPListen == "" {
        log.Fatalf("No --http-listen")
    }

    log.Printf("Serve http://%v/\n", options.HTTPListen)

    if err := http.ListenAndServe(options.HTTPListen, nil); err != nil {
        log.Fatalf("Exit: %v\n", err)
    } else {
        log.Printf("Exit\n")
    }
}
