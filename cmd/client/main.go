package main

import (
	"github.com/jessevdk/go-flags"
	"github.com/qmsk/e2/client"
	"log"
)

var options = struct {
	ClientOptions client.Options `group:"E2 JSON-RPC"`
}{}

var parser = flags.NewParser(&options, flags.Default)

func main() {
	if _, err := parser.Parse(); err != nil {
		log.Fatalf("%v\n", err)
	}
}
