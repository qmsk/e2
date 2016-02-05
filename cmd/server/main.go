
package main

import (
    "github.com/qmsk/e2/client"
    "github.com/jessevdk/go-flags"
    "net/http"
    "log"
    "github.com/qmsk/e2/server"
)

var options = struct{
    ClientOptions   client.Options      `group:"E2 JSON-RPC"`
    ServerOptions   server.Options      `group:"E2 Server"`

    HTTPListen      string              `long:"http-listen" value-name:"[HOST]:PORT" default:":8284"`
    HTTPStatic      string              `long:"http-static" value-name:"PATH"`
}{}

var parser = flags.NewParser(&options, flags.Default)

func main() {
    if _, err := parser.Parse(); err != nil {
        log.Fatalf("%v\n", err)
    }

    client, err := options.ClientOptions.Client()
    if err != nil {
        log.Fatalf("Client %#v: %v\n", options.ClientOptions, err)
    }

    server, err := options.ServerOptions.Server(client)
    if err != nil {
        log.Fatalf("Server %#v: %v\n", options.ServerOptions, err)
    }

    http.Handle("/api/", http.StripPrefix("/api/", server))

    if options.HTTPStatic != "" {
        log.Printf("Serve /static from %v\n", options.HTTPStatic)

        http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(options.HTTPStatic))))
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
