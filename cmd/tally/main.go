package main

import (
	"github.com/jessevdk/go-flags"
	"github.com/qmsk/e2/client"
	"github.com/qmsk/e2/discovery"
	"github.com/qmsk/e2/tally"
	"log"
	"os"
	"os/signal"
)

var options = struct {
	DiscoveryOptions discovery.Options `group:"E2 Discovery"`
	ClientOptions    client.Options    `group:"E2 XML"`
	TallyOptions     tally.Options     `group:"Tally"`
	GPIOOptions      tally.GPIOOptions `group:"Tally GPIO"`

	GPIO bool `long:"gpio"`
}{}

var parser = flags.NewParser(&options, flags.Default)

func main() {
	if _, err := parser.Parse(); err != nil {
		log.Fatalf("%v\n", err)
	}

	tally, err := options.TallyOptions.Tally(options.ClientOptions, options.DiscoveryOptions)
	if err != nil {
		log.Fatalf("Tally: %v\n", err)
	}

	if options.GPIO {
		if tallyGPIO, err := options.GPIOOptions.Make(tally); err != nil {
			log.Fatalf("Start GPIO: %v", err)
		} else {
			defer tallyGPIO.Close()
		}
	}

	// stopping
	stopChan := make(chan os.Signal)

	signal.Notify(stopChan, os.Interrupt)

	go func() { <-stopChan; tally.Stop() }()

	// run
	if err := tally.Run(); err != nil {
		log.Fatalf("Tally.Run: %v\n", err)
	} else {
		log.Printf("Exit")
	}
}
