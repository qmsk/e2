package main

import (
	"github.com/jessevdk/go-flags"
	"github.com/qmsk/e2/server"
	"github.com/qmsk/e2/web"
	"log"
)

var options = struct {
	WebOptions	  web.Options    `group:"Web"`
	ServerOptions server.Options `group:"E2 Server"`
}{}

var parser = flags.NewParser(&options, flags.Default)

func main() {
	if _, err := parser.Parse(); err != nil {
		log.Fatalf("%v", err)
	}

	server, err := options.ServerOptions.Server()
	if err != nil {
		log.Fatalf("Server %#v: %v", options.ServerOptions, err)
	}

	go options.WebOptions.Server(
		web.RoutePrefix("/api/", server.WebAPI()),
		web.RoutePrefix("/events", server.WebEvents()),
		options.WebOptions.RouteStatic("/static/"),
		options.WebOptions.RouteFile("/", "server.html"),
	)

	if err := server.Run(); err != nil {
		log.Fatalf("server.Run: %v", err)
	} else {
		log.Printf("Exit")
	}
}
