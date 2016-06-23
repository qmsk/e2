package main

import (
	"github.com/jessevdk/go-flags"
	"github.com/qmsk/e2/client"
	"github.com/qmsk/e2/discovery"
	"github.com/qmsk/e2/tally"
	"github.com/qmsk/e2/web"
	"log"
	"os"
	"os/signal"
)

var options = struct {
	DiscoveryOptions discovery.Options `group:"E2 Discovery"`
	ClientOptions    client.Options    `group:"E2 XML"`
	TallyOptions     tally.Options     `group:"Tally"`
	GPIOOptions      tally.GPIOOptions `group:"Tally GPIO"`
	WebOptions		 web.Options       `group:"Web API"`

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

	// GPIO
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

	go func() {
		<-stopChan

		// second SIGINT kills
		signal.Stop(stopChan)

		tally.Stop()

	}()

	// Web
	go options.WebOptions.Server(
		web.RoutePrefix("/api/tally/", tally.WebAPI()),
		web.RoutePrefix("/events/", tally.WebEvents()),
		options.WebOptions.RouteStatic("/static/"),
		options.WebOptions.RouteFile("/", "tally.html"),
	)

	// run
	if err := tally.Run(); err != nil {
		log.Fatalf("Tally.Run: %v\n", err)
	} else {
		log.Printf("Exit")
	}
}
