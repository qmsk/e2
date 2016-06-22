package main

import (
	"github.com/jessevdk/go-flags"
	"github.com/qmsk/e2/hetec-dcp"
	"log"
)

var options = struct {
	ClientOptions dcp.Options `group:"Hetec DCP Serial client"`
}{}

var parser = flags.NewParser(&options, flags.Default)

func main() {
	if _, err := parser.Parse(); err != nil {
		log.Fatalf("%v\n", err)
	}
}
